//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type adminSkill struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"isActive"`
	SortOrder int32  `json:"sortOrder"`
	UsedBy    int64  `json:"usedBy"`
}

func findSkill(list []adminSkill, id int64) *adminSkill {
	for i := range list {
		if list[i].ID == id {
			return &list[i]
		}
	}
	return nil
}

func TestAdmin_RoutesRequireAdminRole(t *testing.T) {
	base := newTestServer(t)
	_, _, code := newEmployee(t, "0811111111")
	emp := login(t, base, code)

	for _, r := range []struct{ method, path string }{
		{"GET", "/api/v1/admin/skills"},
		{"POST", "/api/v1/admin/skills"},
		{"PATCH", "/api/v1/admin/skills/1"},
		{"PUT", "/api/v1/admin/skills/order"},
		{"GET", "/api/v1/admin/settings"},
		{"PUT", "/api/v1/admin/settings/history-days"},
	} {
		res := emp.do(r.method, r.path, map[string]any{}, nil)
		assert.Equal(t, http.StatusForbidden, res.StatusCode, r.path)
	}
}

func TestAdmin_ManageSkills(t *testing.T) {
	base := newTestServer(t)
	admin := login(t, base, newAdmin(t))
	name := fmt.Sprintf("Rust%d", time.Now().UnixNano()%1_000_000)

	// Create
	res := admin.do("POST", "/api/v1/admin/skills", map[string]string{"name": "  " + name + " "}, nil)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	id := decodeJSON[struct{ ID int64 }](t, res).ID

	// Duplicate, ignoring case
	res = admin.do("POST", "/api/v1/admin/skills", map[string]string{"name": "rust" + name[4:]}, nil)
	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "ALREADY_EXISTS", decodeJSON[struct{ Error struct{ Code string } }](t, res).Error.Code)

	// Appears at the end of the list, unused, and in the employee picker
	list := decodeJSON[[]adminSkill](t, admin.do("GET", "/api/v1/admin/skills", nil, nil))
	s := findSkill(list, id)
	require.NotNil(t, s)
	assert.Equal(t, name, s.Name, "name was trimmed")
	assert.True(t, s.IsActive)
	assert.Zero(t, s.UsedBy)
	assert.Equal(t, id, list[len(list)-1].ID)

	// An employee picks it → usedBy = 1
	_, _, code := newEmployee(t, "0811111111")
	emp := login(t, base, code)
	res = emp.do("PUT", "/api/v1/profile/skills", map[string]any{"skillIds": []int64{id}}, map[string]string{"If-Match": `"1"`})
	require.Equal(t, http.StatusOK, res.StatusCode)
	list = decodeJSON[[]adminSkill](t, admin.do("GET", "/api/v1/admin/skills", nil, nil))
	assert.EqualValues(t, 1, findSkill(list, id).UsedBy)

	// Turn it off → gone from the picker, kept by the employee
	res = admin.do("PATCH", fmt.Sprintf("/api/v1/admin/skills/%d", id), map[string]bool{"isActive": false}, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	picker := decodeJSON[[]struct{ ID int64 }](t, emp.do("GET", "/api/v1/skills", nil, nil))
	for _, p := range picker {
		assert.NotEqual(t, id, p.ID)
	}
	profile := decodeJSON[struct{ Skills []struct{ ID int64 } }](t, emp.do("GET", "/api/v1/profile", nil, nil))
	require.Len(t, profile.Skills, 1)
	assert.Equal(t, id, profile.Skills[0].ID)

	// Rename onto an existing name → 409; unknown id → 404; empty patch → 400
	res = admin.do("PATCH", fmt.Sprintf("/api/v1/admin/skills/%d", id), map[string]string{"name": "GO"}, nil)
	assert.Equal(t, http.StatusConflict, res.StatusCode)
	res = admin.do("PATCH", "/api/v1/admin/skills/999999999", map[string]bool{"isActive": true}, nil)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	res = admin.do("PATCH", fmt.Sprintf("/api/v1/admin/skills/%d", id), map[string]any{}, nil)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestAdmin_ReorderIsAllOrNothing(t *testing.T) {
	base := newTestServer(t)
	admin := login(t, base, newAdmin(t))
	before := decodeJSON[[]adminSkill](t, admin.do("GET", "/api/v1/admin/skills", nil, nil))
	require.GreaterOrEqual(t, len(before), 2)
	a, b := before[0], before[1]

	// One unknown id → 404 and nothing moves.
	res := admin.do("PUT", "/api/v1/admin/skills/order", []map[string]any{
		{"id": a.ID, "sortOrder": 500}, {"id": 999999999, "sortOrder": 501},
	}, nil)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	after := decodeJSON[[]adminSkill](t, admin.do("GET", "/api/v1/admin/skills", nil, nil))
	assert.Equal(t, a.SortOrder, findSkill(after, a.ID).SortOrder)

	// Swap the first two.
	res = admin.do("PUT", "/api/v1/admin/skills/order", []map[string]any{
		{"id": a.ID, "sortOrder": b.SortOrder}, {"id": b.ID, "sortOrder": a.SortOrder},
	}, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	t.Cleanup(func() {
		admin.do("PUT", "/api/v1/admin/skills/order", []map[string]any{
			{"id": a.ID, "sortOrder": a.SortOrder}, {"id": b.ID, "sortOrder": b.SortOrder},
		}, nil)
	})
	after = decodeJSON[[]adminSkill](t, admin.do("GET", "/api/v1/admin/skills", nil, nil))
	assert.Equal(t, b.SortOrder, findSkill(after, a.ID).SortOrder)
	assert.Equal(t, a.SortOrder, findSkill(after, b.ID).SortOrder)
}

func TestAdmin_HistoryDaysTakesEffectImmediately(t *testing.T) {
	setHistoryDays(t, "7") // restores the original value when the test ends
	base := newTestServer(t)
	adminCode := newAdmin(t)
	admin := login(t, base, adminCode)
	_, _, code := newEmployee(t, "0811111111")
	emp := login(t, base, code)

	type settings struct {
		HistoryDays    int `json:"historyDays"`
		MaxHistoryDays int `json:"maxHistoryDays"`
		UpdatedBy      *struct {
			EmployeeCode string `json:"employeeCode"`
		} `json:"updatedBy"`
		UpdatedAt *time.Time `json:"updatedAt"`
	}

	for _, bad := range []any{0, 91, "30", nil} { // 91: longer than the retention (90)
		res := admin.do("PUT", "/api/v1/admin/settings/history-days", map[string]any{"value": bad}, nil)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode, bad)
	}

	res := admin.do("PUT", "/api/v1/admin/settings/history-days", map[string]int{"value": 30}, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	s := decodeJSON[settings](t, admin.do("GET", "/api/v1/admin/settings", nil, nil))
	assert.Equal(t, 30, s.HistoryDays)
	assert.Equal(t, 90, s.MaxHistoryDays)
	require.NotNil(t, s.UpdatedBy)
	assert.Equal(t, adminCode, s.UpdatedBy.EmployeeCode)
	assert.NotNil(t, s.UpdatedAt)

	hist := decodeJSON[struct{ Days int }](t, emp.do("GET", "/api/v1/profile/history", nil, nil))
	assert.Equal(t, 30, hist.Days, "no restart needed")
}
