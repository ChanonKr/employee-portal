package service

import (
	"context"

	"employee-portal/modules/employee/domain"
)

// Repository is the port this service needs; repository.Repository implements it.
type Repository interface {
	List(ctx context.Context, q domain.EmployeeQuery) (items []domain.EmployeeSummary, total int64, err error)
	Get(ctx context.Context, id int64) (domain.EmployeeDetail, error)
	Stats(ctx context.Context, topSkills, recentChanges int) (domain.DashboardStats, error)
}
