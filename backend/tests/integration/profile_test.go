//go:build integration

package integration

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/profile"
	"employee-portal/modules/shared"
)

func TestUpdateContact_WritesValueAndLogTogether(t *testing.T) {
	ctx := context.Background()
	userID, empID, _ := newEmployee(t, "0811111111")
	uc := newProfileUsecase(t)
	actor := shared.Actor{UserID: userID, EmployeeID: empID}

	v, err := uc.UpdateContact(ctx, profile.UpdateContactInput{Actor: actor, Field: "mobile", Value: "0812345678", Version: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 2, v)
	assert.Equal(t, 1, countLogs(t, empID))

	var oldV, newV string
	require.NoError(t, ownerPool.QueryRow(ctx,
		`SELECT old_value::text, new_value::text FROM profile_change_logs WHERE employee_id = $1`, empID).Scan(&oldV, &newV))
	assert.Equal(t, `"0811111111"`, oldV)
	assert.Equal(t, `"0812345678"`, newV)

	// Same value again: version moves, no new log.
	v, err = uc.UpdateContact(ctx, profile.UpdateContactInput{Actor: actor, Field: "mobile", Value: "081-234-5678", Version: 2})
	require.NoError(t, err)
	assert.EqualValues(t, 3, v)
	assert.Equal(t, 1, countLogs(t, empID))
}

func TestUpdateContact_ConcurrentSameVersion_OneWinsOneConflicts(t *testing.T) {
	ctx := context.Background()
	userID, empID, _ := newEmployee(t, "0811111111")
	uc := newProfileUsecase(t)
	actor := shared.Actor{UserID: userID, EmployeeID: empID}

	const n = 5
	errs := make([]error, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, errs[i] = uc.UpdateContact(ctx, profile.UpdateContactInput{
				Actor: actor, Field: "address", Value: "Address from request", Version: 1,
			})
		}()
	}
	close(start)
	wg.Wait()

	var ok, conflict int
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case assert.ErrorIs(t, err, shared.ErrConflict):
			conflict++
		}
	}
	assert.Equal(t, 1, ok)
	assert.Equal(t, n-1, conflict)
	assert.Equal(t, 1, countLogs(t, empID))
}

func TestUpdateContact_LogFailureRollsBackUpdate(t *testing.T) {
	ctx := context.Background()
	_, empID, _ := newEmployee(t, "0811111111")
	uc := newProfileUsecase(t)

	// changed_by references a user that does not exist → the log insert fails (FK).
	ghost := shared.Actor{UserID: 999_999_999, EmployeeID: empID}
	_, err := uc.UpdateContact(ctx, profile.UpdateContactInput{Actor: ghost, Field: "mobile", Value: "0899999999", Version: 1})
	require.Error(t, err)

	var mobile string
	var version int32
	require.NoError(t, ownerPool.QueryRow(ctx,
		`SELECT mobile_phone, version FROM employees WHERE id = $1`, empID).Scan(&mobile, &version))
	assert.Equal(t, "0811111111", mobile, "update must roll back with the failed log")
	assert.EqualValues(t, 1, version)
	assert.Equal(t, 0, countLogs(t, empID))
}

func TestAppRole_CannotRewriteHistory(t *testing.T) {
	ctx := context.Background()
	_, err := appPool.Exec(ctx, `UPDATE profile_change_logs SET new_value = '"x"'`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	_, err = appPool.Exec(ctx, `DELETE FROM profile_change_logs`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
