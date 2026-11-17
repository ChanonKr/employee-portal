package setting

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/entity"
)

type fakeRepo struct {
	value  string
	getErr error
	setTo  string
	setBy  entity.Actor
}

func (f *fakeRepo) Get(context.Context, string) (string, entity.HistorySettings, error) {
	return f.value, entity.HistorySettings{}, f.getErr
}

func (f *fakeRepo) Set(_ context.Context, _ string, value string, by entity.Actor) error {
	f.setTo, f.setBy = value, by
	return nil
}

var (
	admin    = entity.Actor{UserID: 1, Role: entity.RoleAdmin}
	employee = entity.Actor{UserID: 2, Role: entity.RoleEmployee}
)

func TestSetHistoryDays(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{}
	uc := NewUsecase(repo, 7)

	require.NoError(t, uc.SetHistoryDays(ctx, admin, 30))
	assert.Equal(t, "30", repo.setTo)
	assert.Equal(t, admin, repo.setBy)

	for _, bad := range []int{0, -1, 366} {
		assert.ErrorIs(t, uc.SetHistoryDays(ctx, admin, bad), entity.ErrValidation, bad)
	}
	assert.ErrorIs(t, uc.SetHistoryDays(ctx, employee, 30), entity.ErrForbidden)
}

func TestGetHistory(t *testing.T) {
	ctx := context.Background()

	s, err := NewUsecase(&fakeRepo{value: "14"}, 7).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 14, s.Days)

	s, err = NewUsecase(&fakeRepo{getErr: entity.ErrNotFound}, 7).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 7, s.Days, "env fallback when the row is missing")

	s, err = NewUsecase(&fakeRepo{value: "garbage"}, 7).GetHistory(ctx, admin)
	require.NoError(t, err)
	assert.Equal(t, 7, s.Days, "same fallback as the history page")

	_, err = NewUsecase(&fakeRepo{value: "14"}, 7).GetHistory(ctx, employee)
	assert.ErrorIs(t, err, entity.ErrForbidden)
}
