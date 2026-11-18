package service

import (
	"context"
	"errors"
	"strconv"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/setting/domain"
	"employee-portal/modules/shared"
)

const keyHistoryDays = "history_days"

// Service runs the settings use cases. It never imports GORM or Gin.
type Service struct {
	repo          Repository
	retentionDays int // HISTORY_RETENTION_DAYS: the longest window an admin may set
}

func New(repo Repository, retentionDays int) *Service {
	return &Service{repo: repo, retentionDays: retentionDays}
}

func (svc *Service) GetHistory(ctx context.Context, actor shared.Actor) (domain.HistorySettings, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return domain.HistorySettings{}, err
	}
	raw, s, err := svc.repo.Get(ctx, keyHistoryDays)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return domain.HistorySettings{}, err
	}
	// Same rule as the history page, so admins see the window in effect.
	s.Days = historydomain.HistoryWindow(raw, svc.retentionDays)
	s.MaxDays = svc.retentionDays
	return s, nil
}

func (svc *Service) SetHistoryDays(ctx context.Context, actor shared.Actor, days int) error {
	if err := shared.RequireAdmin(actor); err != nil {
		return err
	}
	if err := domain.ValidateHistoryDays(days, svc.retentionDays); err != nil {
		return err
	}
	return svc.repo.Set(ctx, keyHistoryDays, strconv.Itoa(days), actor)
}
