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
	repo         Repository
	fallbackDays int
}

func New(repo Repository, fallbackDays int) *Service {
	return &Service{repo: repo, fallbackDays: fallbackDays}
}

func (svc *Service) GetHistory(ctx context.Context, actor shared.Actor) (domain.HistorySettings, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return domain.HistorySettings{}, err
	}
	raw, s, err := svc.repo.Get(ctx, keyHistoryDays)
	if errors.Is(err, shared.ErrNotFound) {
		return domain.HistorySettings{Days: svc.fallbackDays}, nil
	}
	if err != nil {
		return domain.HistorySettings{}, err
	}
	// Same fallback rule as the history page, so admins see the value in effect.
	if s.Days, err = historydomain.ParseHistoryDays(raw); err != nil {
		s.Days = svc.fallbackDays
	}
	return s, nil
}

func (svc *Service) SetHistoryDays(ctx context.Context, actor shared.Actor, days int) error {
	if err := shared.RequireAdmin(actor); err != nil {
		return err
	}
	if err := domain.ValidateHistoryDays(days); err != nil {
		return err
	}
	return svc.repo.Set(ctx, keyHistoryDays, strconv.Itoa(days), actor)
}
