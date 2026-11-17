package history

import (
	"context"

	"employee-portal/modules/entity"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	// HistoryDays returns the raw app_settings value, or entity.ErrNotFound.
	HistoryDays(ctx context.Context) (string, error)
	// List returns at most limit rows, newest first.
	List(ctx context.Context, q entity.HistoryQuery, limit int) ([]entity.ChangeLog, error)
	HasOlder(ctx context.Context, employeeID int64, days int) (bool, error)
	EmployeeExists(ctx context.Context, employeeID int64) (bool, error)
}

type Usecase interface {
	// ListOwn returns the signed-in employee's changes within the history window.
	ListOwn(ctx context.Context, in ListInput) (Result, error)
	// ListForEmployee is the admin view: any employee, all history, no time window.
	ListForEmployee(ctx context.Context, employeeID int64, in ListInput) (Result, error)
}
