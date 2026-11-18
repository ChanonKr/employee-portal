package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/setting/domain"
	"employee-portal/modules/shared"
)

type fakeRepo struct {
	value  string
	getErr error
	setTo  string
	setBy  shared.Actor
}

func (f *fakeRepo) Get(context.Context, string) (string, domain.HistorySettings, error) {
	return f.value, domain.HistorySettings{}, f.getErr
}

func (f *fakeRepo) Set(_ context.Context, _ string, value string, by shared.Actor) error {
	f.setTo, f.setBy = value, by
	return nil
}

var (
	admin    = shared.Actor{UserID: 1, Role: shared.RoleAdmin}
	employee = shared.Actor{UserID: 2, Role: shared.RoleEmployee}
)

func TestSetHistoryDays(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{}
	svc := New(repo, 90)

	require.NoError(t, svc.SetHistoryDays(ctx, admin, 30))
	assert.Equal(t, "30", repo.setTo)
	assert.Equal(t, admin, repo.setBy)

	for _, bad := range []int{0, -1, 91} {
		assert.ErrorIs(t, svc.SetHistoryDays(ctx, admin, bad), shared.ErrValidation, bad)
	}
	assert.ErrorIs(t, svc.SetHistoryDays(ctx, employee, 30), shared.ErrForbidden)
}

func TestGetHistory(t *testing.T) {
	ctx := context.Background()

	s, err := New(&fakeRepo{value: "14"}, 90).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 14, s.Days)
	assert.Equal(t, 90, s.MaxDays, "the retention is the longest window")

	s, err = New(&fakeRepo{getErr: shared.ErrNotFound}, 90).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 7, s.Days, "default when the row is missing")

	s, err = New(&fakeRepo{value: "garbage"}, 90).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 7, s.Days, "same default as the history page")

	s, err = New(&fakeRepo{value: "365"}, 90).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 90, s.Days, "capped at the retention")

	_, err = New(&fakeRepo{getErr: errors.New("db down")}, 90).GetHistory(ctx, admin)
	assert.Error(t, err)

	_, err = New(&fakeRepo{value: "14"}, 90).GetHistory(ctx, employee)
	assert.ErrorIs(t, err, shared.ErrForbidden)
}
