package service

import (
	"context"

	"employee-portal/modules/skill/domain"
)

// Repository is the port this service needs; repo.Repository implements it and
// tests use a fake.
type Repository interface {
	ListActive(ctx context.Context) ([]domain.Skill, error)
	ListAll(ctx context.Context) ([]domain.SkillUsage, error)
	// Create returns shared.ErrDuplicate when the name exists (case-insensitive).
	Create(ctx context.Context, name string) (int64, error)
	// Update changes only non-nil fields; shared.ErrNotFound for an unknown id.
	Update(ctx context.Context, id int64, name *string, isActive *bool) error
	// Reorder applies all positions or none; shared.ErrNotFound if any id is unknown.
	Reorder(ctx context.Context, items []domain.SkillOrder) error
}
