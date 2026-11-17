// Package employee is the admin directory: list, search, detail and dashboard stats.
// Editing a profile (including the admin-only name change) stays in module profile,
// which owns the row lock, version and change log.
package employee

import (
	"context"

	"employee-portal/modules/entity"
)

const (
	topSkillsLimit     = 5
	recentChangesLimit = 5
)

type ListInput struct {
	Search   string
	Page     int
	PageSize int
}

type ListResult struct {
	Items []entity.EmployeeSummary
	Page  entity.PageRequest
	Total int64
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) List(ctx context.Context, actor entity.Actor, in ListInput) (ListResult, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return ListResult{}, err
	}
	page, err := entity.NewPageRequest(in.Page, in.PageSize)
	if err != nil {
		return ListResult{}, err
	}
	items, total, err := u.repo.List(ctx, entity.EmployeeQuery{Search: in.Search, PageRequest: page})
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Page: page, Total: total}, nil
}

func (u *usecase) Get(ctx context.Context, actor entity.Actor, id int64) (entity.EmployeeDetail, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return entity.EmployeeDetail{}, err
	}
	return u.repo.Get(ctx, id)
}

func (u *usecase) Stats(ctx context.Context, actor entity.Actor) (entity.DashboardStats, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return entity.DashboardStats{}, err
	}
	return u.repo.Stats(ctx, topSkillsLimit, recentChangesLimit)
}
