package setting

import (
	"context"

	"employee-portal/modules/entity"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	// Get returns the value with who changed it last; entity.ErrNotFound if missing.
	Get(ctx context.Context, key string) (value string, s entity.HistorySettings, err error)
	Set(ctx context.Context, key, value string, by entity.Actor) error
}

type Usecase interface {
	GetHistory(ctx context.Context, actor entity.Actor) (entity.HistorySettings, error)
	// SetHistoryDays takes effect on the next history request (no restart).
	SetHistoryDays(ctx context.Context, actor entity.Actor, days int) error
}
