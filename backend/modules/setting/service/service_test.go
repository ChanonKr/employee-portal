package service

import (
	"context"
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
	svc := New(repo, 7)

	require.NoError(t, svc.SetHistoryDays(ctx, admin, 30))
	assert.Equal(t, "30", repo.setTo)
	assert.Equal(t, admin, repo.setBy)

	for _, bad := range []int{0, -1, 366} {
		assert.ErrorIs(t, svc.SetHistoryDays(ctx, admin, bad), shared.ErrValidation, bad)
	}
	assert.ErrorIs(t, svc.SetHistoryDays(ctx, employee, 30), shared.ErrForbidden)
}

func TestGetHistory(t *testing.T) {
	ctx := context.Background()

	s, err := New(&fakeRepo{value: "14"}, 7).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 14, s.Days)

	s, err = New(&fakeRepo{getErr: shared.ErrNotFound}, 7).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 7, s.Days, "env fallback when the row is missing")

	s, err = New(&fakeRepo{value: "garbage"}, 7).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 7, s.Days, "same fallback as the history page")

	_, err = New(&fakeRepo{value: "14"}, 7).GetHistory(ctx, employee)
	assert.ErrorIs(t, err, shared.ErrForbidden)
}
