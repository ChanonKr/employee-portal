package history

import (
	"context"
	"errors"

	"employee-portal/modules/entity"
)

// Repository is implemented in repository.go.
type Repository interface {
	// HistoryDays returns the raw app_settings value, or entity.ErrNotFound.
	HistoryDays(ctx context.Context) (string, error)
	// List returns at most limit rows, newest first.
	List(ctx context.Context, q entity.HistoryQuery, limit int) ([]entity.ChangeLog, error)
	HasOlder(ctx context.Context, employeeID int64, days int) (bool, error)
	EmployeeExists(ctx context.Context, employeeID int64) (bool, error)
}

type ListInput struct {
	Actor    entity.Actor
	Field    string // "" = all fields
	Page     int    // 0 = default
	PageSize int    // 0 = default
}

type Result struct {
	AllTime  bool // admin view: no time window (Days is 0)
	Days     int  // the window used, so the UI never hard-codes "7"
	HasOlder bool // true when changes exist before the window
	Items    []entity.ChangeLog
	Page     entity.PageRequest
	HasNext  bool
}

type Usecase interface {
	// ListOwn returns the signed-in employee's changes within the history window.
	ListOwn(ctx context.Context, in ListInput) (Result, error)
	// ListForEmployee is the admin view: any employee, all history, no time window.
	ListForEmployee(ctx context.Context, employeeID int64, in ListInput) (Result, error)
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
	if err := entity.RequireAdmin(in.Actor); err != nil {
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
		return Result{}, entity.ErrNotFound
	}
	res, err := u.list(ctx, q)
	if err != nil {
		return Result{}, err
	}
	res.AllTime = true
	return res, nil
}

func buildQuery(employeeID int64, in ListInput) (entity.HistoryQuery, error) {
	page, err := entity.NewPageRequest(in.Page, in.PageSize)
	if err != nil {
		return entity.HistoryQuery{}, err
	}
	q := entity.HistoryQuery{EmployeeID: employeeID, PageRequest: page}
	if in.Field != "" {
		f, err := entity.ParseFieldName(in.Field)
		if err != nil {
			return entity.HistoryQuery{}, err
		}
		q.Field = &f
	}
	return q, nil
}

// list fetches one extra row to tell whether there is a next page without COUNT.
func (u *usecase) list(ctx context.Context, q entity.HistoryQuery) (Result, error) {
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
	if errors.Is(err, entity.ErrNotFound) {
		return u.fallbackDays, nil
	}
	if err != nil {
		return 0, err
	}
	days, err := entity.ParseHistoryDays(raw)
	if err != nil {
		return u.fallbackDays, nil // a bad stored value must not break the page
	}
	return days, nil
}
