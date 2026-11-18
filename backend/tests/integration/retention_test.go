//go:build integration

package integration

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	historyrepo "employee-portal/modules/history/repo"
	historyservice "employee-portal/modules/history/service"
	"employee-portal/pkg/storage"
)

// addChange inserts a change made daysAgo (owner role: the app may not backdate).
// old and new are photo file names or text values; "" is NULL.
func addChange(t *testing.T, empID, userID int64, field, old, new string, daysAgo int) {
	t.Helper()
	_, err := ownerPool.Exec(context.Background(), `
		INSERT INTO profile_change_logs (employee_id, field, old_value, new_value, changed_by, changed_at)
		VALUES ($1, $2, to_jsonb(NULLIF($3, '')), to_jsonb(NULLIF($4, '')), $5, now() - make_interval(days => $6))`,
		empID, field, old, new, userID, daysAgo)
	require.NoError(t, err)
}

func photoLogs(t *testing.T, empID int64) []string {
	t.Helper()
	rows, err := ownerPool.Query(context.Background(),
		`SELECT new_value #>> '{}' FROM profile_change_logs WHERE employee_id = $1 AND field = 'photo' ORDER BY changed_at`, empID)
	require.NoError(t, err)
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		require.NoError(t, rows.Scan(&s))
		out = append(out, s)
	}
	return out
}

func TestRetention_KeepsWindowAndThreeOlderPhotos(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	files, err := storage.NewLocal(dir)
	require.NoError(t, err)
	save := func() string {
		name, err := files.Save(ctx, []byte("jpeg"), ".jpg")
		require.NoError(t, err)
		return name
	}

	// Employee A: 7 photos, uploaded 200, 150, 120, 100, 95, 30 and 10 days ago.
	userA, empA, _ := newEmployee(t, "0811111111")
	agesA := []int{200, 150, 120, 100, 95, 30, 10}
	pa := make([]string, len(agesA))
	for i, age := range agesA {
		pa[i] = save()
		prev := ""
		if i > 0 {
			prev = pa[i-1]
		}
		addChange(t, empA, userA, "photo", prev, pa[i], age)
	}
	_, err = ownerPool.Exec(ctx, `UPDATE employees SET photo_path = $1 WHERE id = $2`, pa[6], empA)
	require.NoError(t, err)
	addChange(t, empA, userA, "mobile_phone", "0811111111", "0822222222", 100) // purged
	addChange(t, empA, userA, "mobile_phone", "0822222222", "0811111111", 5)   // kept

	// Employee B: every photo is older than the retention; the current one is the newest.
	userB, empB, _ := newEmployee(t, "0833333333")
	agesB := []int{300, 200, 150, 120}
	pb := make([]string, len(agesB))
	for i, age := range agesB {
		pb[i] = save()
		prev := ""
		if i > 0 {
			prev = pb[i-1]
		}
		addChange(t, empB, userB, "photo", prev, pb[i], age)
	}
	_, err = ownerPool.Exec(ctx, `UPDATE employees SET photo_path = $1 WHERE id = $2`, pb[3], empB)
	require.NoError(t, err)

	job := historyservice.NewRetention(historyrepo.NewRetention(retentionDB), files, 90, slog.New(slog.NewTextHandler(io.Discard, nil)))
	res, err := job.Run(ctx)
	require.NoError(t, err)
	assert.True(t, res.Ran)

	// A: the 2 photos in the window + the 3 newest before it; the 2 oldest are gone.
	assert.Equal(t, pa[2:], photoLogs(t, empA))
	// B: the 3 newest (including the current photo); the oldest is gone.
	assert.Equal(t, pb[1:], photoLogs(t, empB))
	assert.Equal(t, 6, countLogs(t, empA), "5 photos + the recent mobile change")
	assert.Equal(t, 3, countLogs(t, empB))

	assert.ElementsMatch(t, append(append([]string{}, pa[2:]...), pb[1:]...), filesIn(t, dir),
		"only files no kept change or profile uses are deleted")
	assert.ElementsMatch(t, []string{pa[0], pa[1], pb[0]}, res.UnusedPhotos)

	// A second run has nothing left to do.
	res, err = job.Run(ctx)
	require.NoError(t, err)
	assert.Empty(t, res.UnusedPhotos)
	assert.Equal(t, 6, countLogs(t, empA))
}

func TestRetention_SkipsWhileAnotherInstanceIsPurging(t *testing.T) {
	ctx := context.Background()
	conn, err := ownerPool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()
	_, err = conn.Exec(ctx, `SELECT pg_advisory_lock(hashtext('employee-portal:history-retention'))`)
	require.NoError(t, err)
	defer conn.Exec(ctx, `SELECT pg_advisory_unlock(hashtext('employee-portal:history-retention'))`) //nolint:errcheck

	userID, empID, _ := newEmployee(t, "0811111111")
	addChange(t, empID, userID, "address", "", "Old Road", 200)

	res, err := historyrepo.NewRetention(retentionDB).Purge(ctx, timeDaysAgo(90), 3)
	require.NoError(t, err)
	assert.False(t, res.Ran)
	assert.Equal(t, 1, countLogs(t, empID), "nothing deleted")
}

func TestRetention_RolePermissions(t *testing.T) {
	ctx := context.Background()
	err := retentionDB.WithContext(ctx).Exec(`UPDATE profile_change_logs SET field = 'photo' WHERE false`).Error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied", "the retention role may delete history, never rewrite it")
	err = retentionDB.WithContext(ctx).Exec(`DELETE FROM employees WHERE false`).Error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func timeDaysAgo(days int) time.Time { return time.Now().AddDate(0, 0, -days) }
