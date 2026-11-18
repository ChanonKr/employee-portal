package skill

import (
	"context"

	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	ListActive(ctx context.Context) ([]skilldomain.Skill, error)
	ListAll(ctx context.Context) ([]skilldomain.SkillUsage, error)
	// Create returns shared.ErrDuplicate when the name exists (case-insensitive).
	Create(ctx context.Context, name string) (int64, error)
	// Update changes only non-nil fields; shared.ErrNotFound for an unknown id.
	Update(ctx context.Context, id int64, name *string, isActive *bool) error
	// Reorder applies all positions or none; shared.ErrNotFound if any id is unknown.
	Reorder(ctx context.Context, items []skilldomain.SkillOrder) error
}

type Usecase interface {
	// ListActive returns the catalog shown in the skills picker.
	ListActive(ctx context.Context) ([]skilldomain.Skill, error)

	// Admin → Skills. Every method checks the actor's role again (the router does too).
	AdminList(ctx context.Context, actor shared.Actor) ([]skilldomain.SkillUsage, error)
	Create(ctx context.Context, actor shared.Actor, name string) (int64, error)
	Update(ctx context.Context, actor shared.Actor, in UpdateInput) error
	Reorder(ctx context.Context, actor shared.Actor, items []skilldomain.SkillOrder) error
}
