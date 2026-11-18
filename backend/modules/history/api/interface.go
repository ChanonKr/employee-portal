package api

import (
	"context"

	"employee-portal/modules/history/service"
)

// Service is what the HTTP layer calls; service.Service implements it.
type Service interface {
	// ListOwn returns the signed-in employee's changes within the history window.
	ListOwn(ctx context.Context, in service.ListInput) (service.Result, error)
	// ListForEmployee is the admin view: any employee, all history, no time window.
	ListForEmployee(ctx context.Context, employeeID int64, in service.ListInput) (service.Result, error)
}

var _ Service = (*service.Service)(nil)
