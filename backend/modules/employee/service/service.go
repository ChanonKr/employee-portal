// Package service runs the admin directory: list, search, detail and dashboard stats.
// Editing a profile (including the admin-only name change) stays in the profile feature,
// which owns the row lock, version and change log.
package service

import (
	"context"

	"employee-portal/modules/employee/domain"
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
	Items []domain.EmployeeSummary
	Page  shared.PageRequest
	Total int64
}

// Service runs the admin directory use cases. It never imports GORM or Gin.
type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (svc *Service) List(ctx context.Context, actor shared.Actor, in ListInput) (ListResult, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return ListResult{}, err
	}
	page, err := shared.NewPageRequest(in.Page, in.PageSize)
	if err != nil {
		return ListResult{}, err
	}
	items, total, err := svc.repo.List(ctx, domain.EmployeeQuery{Search: in.Search, PageRequest: page})
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Page: page, Total: total}, nil
}

func (svc *Service) Get(ctx context.Context, actor shared.Actor, id int64) (domain.EmployeeDetail, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return domain.EmployeeDetail{}, err
	}
	return svc.repo.Get(ctx, id)
}

func (svc *Service) Stats(ctx context.Context, actor shared.Actor) (domain.DashboardStats, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return domain.DashboardStats{}, err
	}
	return svc.repo.Stats(ctx, topSkillsLimit, recentChangesLimit)
}
