package employee

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/entity"
)

type fakeRepo struct {
	gotQuery entity.EmployeeQuery
	calls    int
}

func (f *fakeRepo) List(_ context.Context, q entity.EmployeeQuery) ([]entity.EmployeeSummary, int64, error) {
	f.gotQuery = q
	f.calls++
	return []entity.EmployeeSummary{{ID: 1}}, 41, nil
}

func (f *fakeRepo) Get(context.Context, int64) (entity.EmployeeDetail, error) {
	f.calls++
	return entity.EmployeeDetail{}, nil
}

func (f *fakeRepo) Stats(context.Context, int, int) (entity.DashboardStats, error) {
	f.calls++
	return entity.DashboardStats{Employees: 3}, nil
}

var (
	admin    = entity.Actor{UserID: 1, Role: entity.RoleAdmin}
	employee = entity.Actor{UserID: 2, Role: entity.RoleEmployee}
)

func TestList(t *testing.T) {
	repo := &fakeRepo{}
	res, err := NewUsecase(repo).List(context.Background(), admin, ListInput{Search: "som", Page: 2, PageSize: 20})
	require.NoError(t, err)
	assert.EqualValues(t, 41, res.Total)
	assert.Equal(t, "som", repo.gotQuery.Search)
	assert.Equal(t, 20, repo.gotQuery.Offset())

	_, err = NewUsecase(repo).List(context.Background(), admin, ListInput{PageSize: 500})
	assert.ErrorIs(t, err, entity.ErrValidation)
}

func TestEmployeesCannotUseTheDirectory(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{}
	uc := NewUsecase(repo)

	_, err := uc.List(ctx, employee, ListInput{})
	assert.ErrorIs(t, err, entity.ErrForbidden)
	_, err = uc.Get(ctx, employee, 1)
	assert.ErrorIs(t, err, entity.ErrForbidden)
	_, err = uc.Stats(ctx, employee)
	assert.ErrorIs(t, err, entity.ErrForbidden)
	assert.Zero(t, repo.calls, "the repository is never reached")
}
