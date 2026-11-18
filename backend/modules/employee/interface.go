package employee

import (
	"context"

	employeedomain "employee-portal/modules/employee/domain"
	"employee-portal/modules/shared"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	List(ctx context.Context, q employeedomain.EmployeeQuery) (items []employeedomain.EmployeeSummary, total int64, err error)
	Get(ctx context.Context, id int64) (employeedomain.EmployeeDetail, error)
	Stats(ctx context.Context, topSkills, recentChanges int) (employeedomain.DashboardStats, error)
}

type Usecase interface {
	List(ctx context.Context, actor shared.Actor, in ListInput) (ListResult, error)
	Get(ctx context.Context, actor shared.Actor, id int64) (employeedomain.EmployeeDetail, error)
	Stats(ctx context.Context, actor shared.Actor) (employeedomain.DashboardStats, error)
}
