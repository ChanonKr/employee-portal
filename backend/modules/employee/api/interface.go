package api

import (
	"context"

	"employee-portal/modules/employee/domain"
	"employee-portal/modules/employee/service"
	"employee-portal/modules/shared"
)

// Service is what the HTTP layer calls; service.Service implements it.
type Service interface {
	List(ctx context.Context, actor shared.Actor, in service.ListInput) (service.ListResult, error)
	Get(ctx context.Context, actor shared.Actor, id int64) (domain.EmployeeDetail, error)
	Stats(ctx context.Context, actor shared.Actor) (domain.DashboardStats, error)
}

var _ Service = (*service.Service)(nil)
