package api

import (
	"context"

	"employee-portal/modules/setting/domain"
	"employee-portal/modules/setting/service"
	"employee-portal/modules/shared"
)

// Service is what the HTTP layer calls; service.Service implements it.
type Service interface {
	GetHistory(ctx context.Context, actor shared.Actor) (domain.HistorySettings, error)
	// SetHistoryDays takes effect on the next history request (no restart).
	SetHistoryDays(ctx context.Context, actor shared.Actor, days int) error
}

var _ Service = (*service.Service)(nil)
