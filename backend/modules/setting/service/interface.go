package service

import (
	"context"

	"employee-portal/modules/setting/domain"
	"employee-portal/modules/shared"
)

// Repository is the port this service needs; repo.Repository implements it.
type Repository interface {
	// Get returns the value with who changed it last; shared.ErrNotFound if missing.
	Get(ctx context.Context, key string) (value string, s domain.HistorySettings, err error)
	Set(ctx context.Context, key, value string, by shared.Actor) error
}
