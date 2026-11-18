package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/employee/domain"
	"employee-portal/modules/shared"
)

type fakeRepo struct {
	gotQuery domain.EmployeeQuery
	calls    int
}

func (f *fakeRepo) List(_ context.Context, q domain.EmployeeQuery) ([]domain.EmployeeSummary, int64, error) {
	f.gotQuery = q
	f.calls++
	return []domain.EmployeeSummary{{ID: 1}}, 41, nil
}

func (f *fakeRepo) Get(context.Context, int64) (domain.EmployeeDetail, error) {
	f.calls++
	return domain.EmployeeDetail{}, nil
}

func (f *fakeRepo) Stats(context.Context, int, int) (domain.DashboardStats, error) {
	f.calls++
	return domain.DashboardStats{Employees: 3}, nil
}

var (
	admin    = shared.Actor{UserID: 1, Role: shared.RoleAdmin}
	employee = shared.Actor{UserID: 2, Role: shared.RoleEmployee}
)

func TestList(t *testing.T) {
	repo := &fakeRepo{}
	res, err := New(repo).List(context.Background(), admin, ListInput{Search: "som", Page: 2, PageSize: 20})
	require.NoError(t, err)
	assert.EqualValues(t, 41, res.Total)
	assert.Equal(t, "som", repo.gotQuery.Search)
	assert.Equal(t, 20, repo.gotQuery.Offset())

	_, err = New(repo).List(context.Background(), admin, ListInput{PageSize: 500})
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestEmployeesCannotUseTheDirectory(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{}
	svc := New(repo)

	_, err := svc.List(ctx, employee, ListInput{})
	assert.ErrorIs(t, err, shared.ErrForbidden)
	_, err = svc.Get(ctx, employee, 1)
	assert.ErrorIs(t, err, shared.ErrForbidden)
	_, err = svc.Stats(ctx, employee)
	assert.ErrorIs(t, err, shared.ErrForbidden)
	assert.Zero(t, repo.calls, "the repository is never reached")
}
