// Package employee is the admin directory: list, search, detail and dashboard stats.
// Editing a profile (including the admin-only name change) stays in module profile,
// which owns the row lock, version and change log.
package employee

import (
	"context"

	employeedomain "employee-portal/modules/employee/domain"
	"employee-portal/modules/shared"
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
	Items []employeedomain.EmployeeSummary
	Page  shared.PageRequest
	Total int64
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) List(ctx context.Context, actor shared.Actor, in ListInput) (ListResult, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return ListResult{}, err
	}
	page, err := shared.NewPageRequest(in.Page, in.PageSize)
	if err != nil {
		return ListResult{}, err
	}
	items, total, err := u.repo.List(ctx, employeedomain.EmployeeQuery{Search: in.Search, PageRequest: page})
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Page: page, Total: total}, nil
}

func (u *usecase) Get(ctx context.Context, actor shared.Actor, id int64) (employeedomain.EmployeeDetail, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return employeedomain.EmployeeDetail{}, err
	}
	return u.repo.Get(ctx, id)
}

func (u *usecase) Stats(ctx context.Context, actor shared.Actor) (employeedomain.DashboardStats, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return employeedomain.DashboardStats{}, err
	}
	return u.repo.Stats(ctx, topSkillsLimit, recentChangesLimit)
}
