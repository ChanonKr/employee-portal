package history

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
)

type fakeRepo struct {
	days     string
	daysErr  error
	logs     []historydomain.ChangeLog // newest first
	hasOlder bool

	gotQuery historydomain.HistoryQuery
	gotLimit int
}

func (f *fakeRepo) HistoryDays(context.Context) (string, error) { return f.days, f.daysErr }

func (f *fakeRepo) List(_ context.Context, q historydomain.HistoryQuery, limit int) ([]historydomain.ChangeLog, error) {
	f.gotQuery, f.gotLimit = q, limit
	start := min(q.Offset(), len(f.logs))
	end := min(start+limit, len(f.logs))
	return f.logs[start:end], nil
}

func (f *fakeRepo) HasOlder(context.Context, int64, int) (bool, error) { return f.hasOlder, nil }

func (f *fakeRepo) EmployeeExists(_ context.Context, id int64) (bool, error) { return id != 404, nil }

func logs(n int) []historydomain.ChangeLog {
	out := make([]historydomain.ChangeLog, n)
	for i := range out {
		out[i] = historydomain.ChangeLog{ID: int64(n - i)}
	}
	return out
}

var actor = shared.Actor{UserID: 1, EmployeeID: 10}

func TestListOwn(t *testing.T) {
	ctx := context.Background()

	t.Run("uses the admin setting and defaults paging", func(t *testing.T) {
		repo := &fakeRepo{days: "30", logs: logs(3), hasOlder: true}
		res, err := NewUsecase(repo, 7).ListOwn(ctx, ListInput{Actor: actor})
		require.NoError(t, err)
		assert.Equal(t, 30, res.Days)
		assert.Equal(t, 30, repo.gotQuery.Days)
		assert.EqualValues(t, 10, repo.gotQuery.EmployeeID, "employee comes from the actor")
		assert.True(t, res.HasOlder)
		assert.Len(t, res.Items, 3)
		assert.Equal(t, shared.PageRequest{Page: 1, PageSize: shared.DefaultPageSize}, res.Page)
		assert.False(t, res.HasNext)
	})

	t.Run("falls back to env days when the setting is missing or invalid", func(t *testing.T) {
		for _, repo := range []*fakeRepo{{daysErr: shared.ErrNotFound}, {days: "0"}, {days: "abc"}} {
			res, err := NewUsecase(repo, 7).ListOwn(ctx, ListInput{Actor: actor})
			require.NoError(t, err)
			assert.Equal(t, 7, res.Days)
		}
	})

	t.Run("a database error while reading the setting is returned", func(t *testing.T) {
		repo := &fakeRepo{daysErr: errors.New("db down")}
		_, err := NewUsecase(repo, 7).ListOwn(ctx, ListInput{Actor: actor})
		assert.Error(t, err)
	})

	t.Run("asks for one extra row and trims it to report hasNext", func(t *testing.T) {
		repo := &fakeRepo{days: "7", logs: logs(5)}
		res, err := NewUsecase(repo, 7).ListOwn(ctx, ListInput{Actor: actor, Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, 3, repo.gotLimit)
		assert.Len(t, res.Items, 2)
		assert.True(t, res.HasNext)

		res, err = NewUsecase(repo, 7).ListOwn(ctx, ListInput{Actor: actor, Page: 3, PageSize: 2})
		require.NoError(t, err)
		assert.Len(t, res.Items, 1)
		assert.False(t, res.HasNext)
	})

	t.Run("field filter", func(t *testing.T) {
		repo := &fakeRepo{days: "7"}
		_, err := NewUsecase(repo, 7).ListOwn(ctx, ListInput{Actor: actor, Field: "skills"})
		require.NoError(t, err)
		require.NotNil(t, repo.gotQuery.Field)
		assert.Equal(t, historydomain.FieldSkills, *repo.gotQuery.Field)

		_, err = NewUsecase(repo, 7).ListOwn(ctx, ListInput{Actor: actor, Field: "password_hash"})
		assert.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("invalid paging", func(t *testing.T) {
		_, err := NewUsecase(&fakeRepo{days: "7"}, 7).ListOwn(ctx, ListInput{Actor: actor, PageSize: 1000})
		assert.ErrorIs(t, err, shared.ErrValidation)
	})
}

func TestListForEmployee(t *testing.T) {
	ctx := context.Background()
	admin := shared.Actor{UserID: 9, Role: shared.RoleAdmin}

	repo := &fakeRepo{days: "7", logs: logs(3), hasOlder: true}
	res, err := NewUsecase(repo, 7).ListForEmployee(ctx, 55, ListInput{Actor: admin})
	require.NoError(t, err)
	assert.True(t, res.AllTime)
	assert.Zero(t, repo.gotQuery.Days, "no time window")
	assert.EqualValues(t, 55, repo.gotQuery.EmployeeID, "employee from the URL, not the actor")
	assert.Len(t, res.Items, 3)

	_, err = NewUsecase(repo, 7).ListForEmployee(ctx, 404, ListInput{Actor: admin})
	assert.ErrorIs(t, err, shared.ErrNotFound)

	_, err = NewUsecase(repo, 7).ListForEmployee(ctx, 55, ListInput{Actor: actor})
	assert.ErrorIs(t, err, shared.ErrForbidden)
}
