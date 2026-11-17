package setting

import (
	"context"
	"errors"
	"strconv"

	"employee-portal/modules/entity"
)

const keyHistoryDays = "history_days"

type usecase struct {
	repo         Repository
	fallbackDays int
}

func NewUsecase(repo Repository, fallbackDays int) Usecase {
	return &usecase{repo: repo, fallbackDays: fallbackDays}
}

func (u *usecase) GetHistory(ctx context.Context, actor entity.Actor) (entity.HistorySettings, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return entity.HistorySettings{}, err
	}
	raw, s, err := u.repo.Get(ctx, keyHistoryDays)
	if errors.Is(err, entity.ErrNotFound) {
		return entity.HistorySettings{Days: u.fallbackDays}, nil
	}
	if err != nil {
		return entity.HistorySettings{}, err
	}
	// Same fallback rule as the history page, so admins see the value in effect.
	if s.Days, err = entity.ParseHistoryDays(raw); err != nil {
		s.Days = u.fallbackDays
	}
	return s, nil
}

func (u *usecase) SetHistoryDays(ctx context.Context, actor entity.Actor, days int) error {
	if err := entity.RequireAdmin(actor); err != nil {
		return err
	}
	if err := entity.ValidateHistoryDays(days); err != nil {
		return err
	}
	return u.repo.Set(ctx, keyHistoryDays, strconv.Itoa(days), actor)
}
