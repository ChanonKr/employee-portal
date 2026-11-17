package employee

import (
	"context"

	"employee-portal/modules/entity"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	List(ctx context.Context, q entity.EmployeeQuery) (items []entity.EmployeeSummary, total int64, err error)
	Get(ctx context.Context, id int64) (entity.EmployeeDetail, error)
	Stats(ctx context.Context, topSkills, recentChanges int) (entity.DashboardStats, error)
}

type Usecase interface {
	List(ctx context.Context, actor entity.Actor, in ListInput) (ListResult, error)
	Get(ctx context.Context, actor entity.Actor, id int64) (entity.EmployeeDetail, error)
	Stats(ctx context.Context, actor entity.Actor) (entity.DashboardStats, error)
}
