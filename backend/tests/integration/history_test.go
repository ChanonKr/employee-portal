//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/entity"
	"employee-portal/modules/history"
)

// addLog inserts a change log `ago` in the past (e.g. "8 days").
func addLog(t *testing.T, empID, userID int64, field, oldJSON, newJSON, ago string) {
	t.Helper()
	_, err := ownerPool.Exec(context.Background(), `
		INSERT INTO profile_change_logs (employee_id, field, old_value, new_value, changed_by, changed_at)
		VALUES ($1, $2, $3::jsonb, $4::jsonb, $5, now() - $6::interval)`,
		empID, field, oldJSON, newJSON, userID, ago)
	require.NoError(t, err)
}

// setHistoryDays changes the admin setting for one test and restores it afterwards.
func setHistoryDays(t *testing.T, value string) {
	t.Helper()
	ctx := context.Background()
	var before string
	require.NoError(t, ownerPool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = 'history_days'`).Scan(&before))
	_, err := ownerPool.Exec(ctx, `UPDATE app_settings SET value = $1 WHERE key = 'history_days'`, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = ownerPool.Exec(ctx, `UPDATE app_settings SET value = $1 WHERE key = 'history_days'`, before)
	})
}

func TestHistory_WindowFollowsSettingWithoutRestart(t *testing.T) {
	ctx := context.Background()
	userID, empID, _ := newEmployee(t, "0811111111")
	addLog(t, empID, userID, "mobile_phone", `"0800000000"`, `"0811111111"`, "8 days")
	addLog(t, empID, userID, "address", `null`, `"New address"`, "2 days")

	uc := history.NewUsecase(history.NewRepository(appDB), 7)
	actor := entity.Actor{UserID: userID, EmployeeID: empID}

	setHistoryDays(t, "7")
	res, err := uc.ListOwn(ctx, history.ListInput{Actor: actor})
	require.NoError(t, err)
	assert.Equal(t, 7, res.Days)
	require.Len(t, res.Items, 1, "the 8-day-old change is outside a 7-day window")
	assert.Equal(t, entity.FieldAddress, res.Items[0].Field)
	assert.Equal(t, "New address", res.Items[0].New)
	assert.True(t, res.HasOlder)

	setHistoryDays(t, "10") // same use case instance: no restart
	res, err = uc.ListOwn(ctx, history.ListInput{Actor: actor})
	require.NoError(t, err)
	assert.Equal(t, 10, res.Days)
	assert.Len(t, res.Items, 2)
	assert.False(t, res.HasOlder)
	assert.Equal(t, entity.FieldAddress, res.Items[0].Field, "newest first")
}

func TestHistory_EmptyStates(t *testing.T) {
	ctx := context.Background()
	uc := history.NewUsecase(history.NewRepository(appDB), 7)

	// Never edited.
	userID, empID, _ := newEmployee(t, "0811111111")
	res, err := uc.ListOwn(ctx, history.ListInput{Actor: entity.Actor{UserID: userID, EmployeeID: empID}})
	require.NoError(t, err)
	assert.Empty(t, res.Items)
	assert.NotNil(t, res.Items)
	assert.False(t, res.HasOlder)

	// Edited, but only long ago.
	userID, empID, _ = newEmployee(t, "0811111111")
	addLog(t, empID, userID, "address", `null`, `"Old"`, "30 days")
	res, err = uc.ListOwn(ctx, history.ListInput{Actor: entity.Actor{UserID: userID, EmployeeID: empID}})
	require.NoError(t, err)
	assert.Empty(t, res.Items)
	assert.True(t, res.HasOlder)
}

func TestHistory_FilterPagingAndValues(t *testing.T) {
	ctx := context.Background()
	userID, empID, _ := newEmployee(t, "0811111111")
	addLog(t, empID, userID, "skills", `["Go","Flutter"]`, `["Go","React"]`, "1 hour")
	addLog(t, empID, userID, "mobile_phone", `"0811111111"`, `"0822222222"`, "2 hours")
	addLog(t, empID, userID, "mobile_phone", `"0800000000"`, `"0811111111"`, "3 hours")
	addLog(t, empID, userID, "address", `null`, `"A"`, "4 hours")

	uc := history.NewUsecase(history.NewRepository(appDB), 7)
	actor := entity.Actor{UserID: userID, EmployeeID: empID}

	res, err := uc.ListOwn(ctx, history.ListInput{Actor: actor, Field: "mobile_phone"})
	require.NoError(t, err)
	require.Len(t, res.Items, 2)
	for _, it := range res.Items {
		assert.Equal(t, entity.FieldMobilePhone, it.Field)
	}

	res, err = uc.ListOwn(ctx, history.ListInput{Actor: actor, Field: "skills"})
	require.NoError(t, err)
	require.Len(t, res.Items, 1)
	assert.Equal(t, []any{"Go", "Flutter"}, res.Items[0].Old)
	assert.Equal(t, []any{"Go", "React"}, res.Items[0].New)

	page1, err := uc.ListOwn(ctx, history.ListInput{Actor: actor, Page: 1, PageSize: 3})
	require.NoError(t, err)
	assert.Len(t, page1.Items, 3)
	assert.True(t, page1.HasNext)
	page2, err := uc.ListOwn(ctx, history.ListInput{Actor: actor, Page: 2, PageSize: 3})
	require.NoError(t, err)
	require.Len(t, page2.Items, 1)
	assert.False(t, page2.HasNext)
	assert.Nil(t, page2.Items[0].Old, "SQL NULL comes back as nil")
}

func TestHistory_OnlyOwnEmployee(t *testing.T) {
	ctx := context.Background()
	userA, empA, _ := newEmployee(t, "0811111111")
	userB, empB, _ := newEmployee(t, "0811111111")
	addLog(t, empB, userB, "address", `null`, `"B's address"`, "1 hour")

	uc := history.NewUsecase(history.NewRepository(appDB), 7)
	res, err := uc.ListOwn(ctx, history.ListInput{Actor: entity.Actor{UserID: userA, EmployeeID: empA}})
	require.NoError(t, err)
	assert.Empty(t, res.Items)
}
