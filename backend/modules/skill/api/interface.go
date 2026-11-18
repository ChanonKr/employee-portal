package api

import (
	"context"

	"employee-portal/modules/shared"
	"employee-portal/modules/skill/domain"
	"employee-portal/modules/skill/service"
)

// Service is what the HTTP layer calls; service.Service implements it.
type Service interface {
	// ListActive returns the catalog shown in the skills picker.
	ListActive(ctx context.Context) ([]domain.Skill, error)

	// Admin → Skills. Every method checks the actor's role again (the router does too).
	AdminList(ctx context.Context, actor shared.Actor) ([]domain.SkillUsage, error)
	Create(ctx context.Context, actor shared.Actor, name string) (int64, error)
	Update(ctx context.Context, actor shared.Actor, in service.UpdateInput) error
	Reorder(ctx context.Context, actor shared.Actor, items []domain.SkillOrder) error
}

var _ Service = (*service.Service)(nil)
