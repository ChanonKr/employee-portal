package service

import (
	"context"

	"employee-portal/modules/shared"
	"employee-portal/modules/skill/domain"
)

type UpdateInput struct {
	ID       int64
	Name     *string
	IsActive *bool
}

// Service runs the skill catalog use cases. It never imports GORM or Gin.
type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListActive(ctx context.Context) ([]domain.Skill, error) {
	return s.repo.ListActive(ctx)
}

func (s *Service) AdminList(ctx context.Context, actor shared.Actor) ([]domain.SkillUsage, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return nil, err
	}
	return s.repo.ListAll(ctx)
}

func (s *Service) Create(ctx context.Context, actor shared.Actor, raw string) (int64, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return 0, err
	}
	name, err := domain.NormalizeSkillName(raw)
	if err != nil {
		return 0, err
	}
	return s.repo.Create(ctx, name)
}

// Update renames and/or turns a skill on or off. Skills are never deleted:
// turning one off hides it from the picker, and employees who have it keep it.
func (s *Service) Update(ctx context.Context, actor shared.Actor, in UpdateInput) error {
	if err := shared.RequireAdmin(actor); err != nil {
		return err
	}
	if in.Name == nil && in.IsActive == nil {
		return shared.NewValidationError("body", "nothing to update")
	}
	if in.Name != nil {
		name, err := domain.NormalizeSkillName(*in.Name)
		if err != nil {
			return err
		}
		in.Name = &name
	}
	return s.repo.Update(ctx, in.ID, in.Name, in.IsActive)
}

func (s *Service) Reorder(ctx context.Context, actor shared.Actor, items []domain.SkillOrder) error {
	if err := shared.RequireAdmin(actor); err != nil {
		return err
	}
	if err := domain.ValidateSkillOrder(items); err != nil {
		return err
	}
	return s.repo.Reorder(ctx, items)
}
