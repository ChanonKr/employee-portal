package setting

import (
	"context"

	settingdomain "employee-portal/modules/setting/domain"
	"employee-portal/modules/shared"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	// Get returns the value with who changed it last; shared.ErrNotFound if missing.
	Get(ctx context.Context, key string) (value string, s settingdomain.HistorySettings, err error)
	Set(ctx context.Context, key, value string, by shared.Actor) error
}

type Usecase interface {
	GetHistory(ctx context.Context, actor shared.Actor) (settingdomain.HistorySettings, error)
	// SetHistoryDays takes effect on the next history request (no restart).
	SetHistoryDays(ctx context.Context, actor shared.Actor, days int) error
}
