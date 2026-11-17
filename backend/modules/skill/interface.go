package skill

import (
	"context"

	"employee-portal/modules/entity"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	ListActive(ctx context.Context) ([]entity.Skill, error)
	ListAll(ctx context.Context) ([]entity.SkillUsage, error)
	// Create returns entity.ErrDuplicate when the name exists (case-insensitive).
	Create(ctx context.Context, name string) (int64, error)
	// Update changes only non-nil fields; entity.ErrNotFound for an unknown id.
	Update(ctx context.Context, id int64, name *string, isActive *bool) error
	// Reorder applies all positions or none; entity.ErrNotFound if any id is unknown.
	Reorder(ctx context.Context, items []entity.SkillOrder) error
}

type Usecase interface {
	// ListActive returns the catalog shown in the skills picker.
	ListActive(ctx context.Context) ([]entity.Skill, error)

	// Admin → Skills. Every method checks the actor's role again (the router does too).
	AdminList(ctx context.Context, actor entity.Actor) ([]entity.SkillUsage, error)
	Create(ctx context.Context, actor entity.Actor, name string) (int64, error)
	Update(ctx context.Context, actor entity.Actor, in UpdateInput) error
	Reorder(ctx context.Context, actor entity.Actor, items []entity.SkillOrder) error
}
