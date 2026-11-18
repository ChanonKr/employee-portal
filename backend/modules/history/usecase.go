package history

import (
	"context"
	"errors"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
)

type ListInput struct {
	Actor    shared.Actor
	Field    string // "" = all fields
	Page     int    // 0 = default
	PageSize int    // 0 = default
}

type Result struct {
	AllTime  bool // admin view: no time window (Days is 0)
	Days     int  // the window used, so the UI never hard-codes "7"
	HasOlder bool // true when changes exist before the window
	Items    []historydomain.ChangeLog
	Page     shared.PageRequest
	HasNext  bool
}

type usecase struct {
	repo         Repository
	fallbackDays int // HISTORY_DAYS from env, used when the setting is missing or invalid
}

func NewUsecase(repo Repository, fallbackDays int) Usecase {
	return &usecase{repo: repo, fallbackDays: fallbackDays}
}

func (u *usecase) ListOwn(ctx context.Context, in ListInput) (Result, error) {
	q, err := buildQuery(in.Actor.EmployeeID, in)
	if err != nil {
		return Result{}, err
	}
	if q.Days, err = u.historyDays(ctx); err != nil {
		return Result{}, err
	}
	res, err := u.list(ctx, q)
	if err != nil {
		return Result{}, err
	}
	if res.HasOlder, err = u.repo.HasOlder(ctx, q.EmployeeID, q.Days); err != nil {
		return Result{}, err
	}
	return res, nil
}

func (u *usecase) ListForEmployee(ctx context.Context, employeeID int64, in ListInput) (Result, error) {
	if err := shared.RequireAdmin(in.Actor); err != nil {
		return Result{}, err
	}
	q, err := buildQuery(employeeID, in) // Days stays 0: no window
	if err != nil {
		return Result{}, err
	}
	exists, err := u.repo.EmployeeExists(ctx, employeeID)
	if err != nil {
		return Result{}, err
	}
	if !exists {
		return Result{}, shared.ErrNotFound
	}
	res, err := u.list(ctx, q)
	if err != nil {
		return Result{}, err
	}
	res.AllTime = true
	return res, nil
}

func buildQuery(employeeID int64, in ListInput) (historydomain.HistoryQuery, error) {
	page, err := shared.NewPageRequest(in.Page, in.PageSize)
	if err != nil {
		return historydomain.HistoryQuery{}, err
	}
	q := historydomain.HistoryQuery{EmployeeID: employeeID, PageRequest: page}
	if in.Field != "" {
		f, err := historydomain.ParseFieldName(in.Field)
		if err != nil {
			return historydomain.HistoryQuery{}, err
		}
		q.Field = &f
	}
	return q, nil
}

// list fetches one extra row to tell whether there is a next page without COUNT.
func (u *usecase) list(ctx context.Context, q historydomain.HistoryQuery) (Result, error) {
	items, err := u.repo.List(ctx, q, q.PageSize+1)
	if err != nil {
		return Result{}, err
	}
	hasNext := len(items) > q.PageSize
	if hasNext {
		items = items[:q.PageSize]
	}
	return Result{Days: q.Days, Items: items, Page: q.PageRequest, HasNext: hasNext}, nil
}

// historyDays reads the admin setting (no restart needed after a change).
func (u *usecase) historyDays(ctx context.Context) (int, error) {
	raw, err := u.repo.HistoryDays(ctx)
	if errors.Is(err, shared.ErrNotFound) {
		return u.fallbackDays, nil
	}
	if err != nil {
		return 0, err
	}
	days, err := historydomain.ParseHistoryDays(raw)
	if err != nil {
		return u.fallbackDays, nil // a bad stored value must not break the page
	}
	return days, nil
}
