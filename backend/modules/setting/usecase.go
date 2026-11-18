package setting

import (
	"context"
	"errors"
	"strconv"

	historydomain "employee-portal/modules/history/domain"
	settingdomain "employee-portal/modules/setting/domain"
	"employee-portal/modules/shared"
)

const keyHistoryDays = "history_days"

type usecase struct {
	repo         Repository
	fallbackDays int
}

func NewUsecase(repo Repository, fallbackDays int) Usecase {
	return &usecase{repo: repo, fallbackDays: fallbackDays}
}

func (u *usecase) GetHistory(ctx context.Context, actor shared.Actor) (settingdomain.HistorySettings, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return settingdomain.HistorySettings{}, err
	}
	raw, s, err := u.repo.Get(ctx, keyHistoryDays)
	if errors.Is(err, shared.ErrNotFound) {
		return settingdomain.HistorySettings{Days: u.fallbackDays}, nil
	}
	if err != nil {
		return settingdomain.HistorySettings{}, err
	}
	// Same fallback rule as the history page, so admins see the value in effect.
	if s.Days, err = historydomain.ParseHistoryDays(raw); err != nil {
		s.Days = u.fallbackDays
	}
	return s, nil
}

func (u *usecase) SetHistoryDays(ctx context.Context, actor shared.Actor, days int) error {
	if err := shared.RequireAdmin(actor); err != nil {
		return err
	}
	if err := settingdomain.ValidateHistoryDays(days); err != nil {
		return err
	}
	return u.repo.Set(ctx, keyHistoryDays, strconv.Itoa(days), actor)
}
