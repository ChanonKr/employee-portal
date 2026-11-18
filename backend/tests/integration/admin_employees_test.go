//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type employeeList struct {
	Items []struct {
		ID           int64  `json:"id"`
		EmployeeCode string `json:"employeeCode"`
		FirstName    string `json:"firstName"`
		Role         string `json:"role"`
	} `json:"items"`
	Page struct {
		Page     int  `json:"page"`
		PageSize int  `json:"pageSize"`
		HasNext  bool `json:"hasNext"`
	} `json:"page"`
	Total int64 `json:"total"`
}

func TestAdmin_EmployeeDirectory(t *testing.T) {
	base := newTestServer(t)
	admin := login(t, base, newAdmin(t))
	_, empID, code := newEmployee(t, "0811111111")

	// Search by code (case-insensitive)
	res := admin.do("GET", "/api/v1/admin/employees?q="+url.QueryEscape(code[2:]), nil, nil)
	require.Equal(t, http.StatusOK, res.StatusCode)
	list := decodeJSON[employeeList](t, res)
	require.Len(t, list.Items, 1)
	assert.Equal(t, empID, list.Items[0].ID)
	assert.EqualValues(t, 1, list.Total)

	// "%" typed by the user is literal, not "match everything"
	list = decodeJSON[employeeList](t, admin.do("GET", "/api/v1/admin/employees?q=%25", nil, nil))
	assert.Empty(t, list.Items)
	assert.Zero(t, list.Total)

	// Paging
	all := decodeJSON[employeeList](t, admin.do("GET", "/api/v1/admin/employees?pageSize=2", nil, nil))
	require.GreaterOrEqual(t, all.Total, int64(3))
	assert.Len(t, all.Items, 2)
	assert.True(t, all.Page.HasNext)

	// Detail with ETag; unknown and malformed ids are 404
	res = admin.do("GET", fmt.Sprintf("/api/v1/admin/employees/%d", empID), nil, nil)
	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `"1"`, res.Header.Get("ETag"))
	detail := decodeJSON[struct {
		EmployeeCode string `json:"employeeCode"`
		Role         string `json:"role"`
		IsActive     bool   `json:"isActive"`
	}](t, res)
	assert.Equal(t, code, detail.EmployeeCode)
	assert.Equal(t, "employee", detail.Role)
	assert.True(t, detail.IsActive)
	assert.Equal(t, http.StatusNotFound, admin.do("GET", "/api/v1/admin/employees/999999999", nil, nil).StatusCode)
	assert.Equal(t, http.StatusNotFound, admin.do("GET", "/api/v1/admin/employees/abc", nil, nil).StatusCode)
}

func TestAdmin_RenameEmployeeAndSeeAllHistory(t *testing.T) {
	base := newTestServer(t)
	admin := login(t, base, newAdmin(t))
	userID, empID, code := newEmployee(t, "0811111111")
	addLog(t, empID, userID, "address", `null`, `"Old"`, "60 days") // outside any employee window

	path := fmt.Sprintf("/api/v1/admin/employees/%d/name", empID)
	res := admin.do("PUT", path, map[string]string{"firstName": "Somsri", "lastName": "User"}, nil)
	assert.Equal(t, http.StatusPreconditionRequired, res.StatusCode)

	res = admin.do("PUT", path, map[string]string{"firstName": " Somsri ", "lastName": "User"}, map[string]string{"If-Match": `"1"`})
	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `"2"`, res.Header.Get("ETag"))

	res = admin.do("PUT", path, map[string]string{"firstName": "", "lastName": "User"}, map[string]string{"If-Match": `"2"`})
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	res = admin.do("PUT", path, map[string]string{"firstName": "X", "lastName": "Y"}, map[string]string{"If-Match": `"1"`})
	assert.Equal(t, http.StatusConflict, res.StatusCode)

	// The employee sees the new name and the admin's edit in their history.
	emp := login(t, base, code)
	profile := decodeJSON[struct{ FirstName string }](t, emp.do("GET", "/api/v1/profile", nil, nil))
	assert.Equal(t, "Somsri", profile.FirstName)

	type history struct {
		Days  *int `json:"days"`
		Items []struct {
			Field     string          `json:"field"`
			NewValue  json.RawMessage `json:"newValue"`
			ChangedBy struct {
				Role string `json:"role"`
			} `json:"changedBy"`
		} `json:"items"`
	}
	own := decodeJSON[history](t, emp.do("GET", "/api/v1/profile/history", nil, nil))
	require.NotNil(t, own.Days)
	require.Len(t, own.Items, 1, "only the rename is inside the employee's window")
	assert.Equal(t, "first_name", own.Items[0].Field)
	assert.Equal(t, "admin", own.Items[0].ChangedBy.Role)

	// Admin sees everything, with no window.
	res = admin.do("GET", fmt.Sprintf("/api/v1/admin/employees/%d/history", empID), nil, nil)
	require.Equal(t, http.StatusOK, res.StatusCode)
	all := decodeJSON[history](t, res)
	assert.Nil(t, all.Days, "null = no window: everything still kept")
	assert.Len(t, all.Items, 2)

	assert.Equal(t, http.StatusNotFound,
		admin.do("GET", "/api/v1/admin/employees/999999999/history", nil, nil).StatusCode)
	assert.Equal(t, http.StatusForbidden,
		emp.do("GET", fmt.Sprintf("/api/v1/admin/employees/%d/history", empID), nil, nil).StatusCode)
}

func TestAdmin_DashboardStats(t *testing.T) {
	base := newTestServer(t)
	admin := login(t, base, newAdmin(t))

	type stats struct {
		Employees        int64 `json:"employees"`
		CompleteProfiles int64 `json:"completeProfiles"`
		TopSkills        []struct {
			Name   string `json:"name"`
			UsedBy int64  `json:"usedBy"`
		} `json:"topSkills"`
		RecentChanges []struct {
			EmployeeCode string `json:"employeeCode"`
			Field        string `json:"field"`
		} `json:"recentChanges"`
	}
	before := decodeJSON[stats](t, admin.do("GET", "/api/v1/admin/stats", nil, nil))

	// Complete one profile: photo + all contact fields + a skill.
	userID, empID, code := newEmployee(t, "0811111111")
	_, err := ownerPool.Exec(t.Context(), `
		UPDATE employees SET photo_path = 'x.jpg', contact_email = 'a@example.com', address = 'A'
		WHERE id = $1`, empID)
	require.NoError(t, err)
	_, err = ownerPool.Exec(t.Context(),
		`INSERT INTO employee_skills (employee_id, skill_id) SELECT $1, id FROM skills ORDER BY id LIMIT 1`, empID)
	require.NoError(t, err)
	addLog(t, empID, userID, "address", `null`, `"A"`, "0 seconds")

	after := decodeJSON[stats](t, admin.do("GET", "/api/v1/admin/stats", nil, nil))
	assert.Equal(t, before.Employees+1, after.Employees)
	assert.Equal(t, before.CompleteProfiles+1, after.CompleteProfiles)
	assert.LessOrEqual(t, len(after.TopSkills), 5)
	assert.LessOrEqual(t, len(after.RecentChanges), 5)
	require.NotEmpty(t, after.RecentChanges)
	assert.Equal(t, code, after.RecentChanges[0].EmployeeCode, "newest change first")
}
