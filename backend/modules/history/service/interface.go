package service

import (
	"context"

	"employee-portal/modules/history/domain"
)

// Repository is the port this service needs; repo.Repository implements it.
type Repository interface {
	// HistoryDays returns the raw app_settings value, or shared.ErrNotFound.
	HistoryDays(ctx context.Context) (string, error)
	// List returns at most limit rows, newest first.
	List(ctx context.Context, q domain.HistoryQuery, limit int) ([]domain.ChangeLog, error)
	HasOlder(ctx context.Context, employeeID int64, days int) (bool, error)
	EmployeeExists(ctx context.Context, employeeID int64) (bool, error)
}
