package skill

import (
	"context"

	"employee-portal/modules/entity"
)

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

type UpdateInput struct {
	ID       int64
	Name     *string
	IsActive *bool
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

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) ListActive(ctx context.Context) ([]entity.Skill, error) {
	return u.repo.ListActive(ctx)
}

func (u *usecase) AdminList(ctx context.Context, actor entity.Actor) ([]entity.SkillUsage, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return nil, err
	}
	return u.repo.ListAll(ctx)
}

func (u *usecase) Create(ctx context.Context, actor entity.Actor, raw string) (int64, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return 0, err
	}
	name, err := entity.NormalizeSkillName(raw)
	if err != nil {
		return 0, err
	}
	return u.repo.Create(ctx, name)
}

// Update renames and/or turns a skill on or off. Skills are never deleted:
// turning one off hides it from the picker, and employees who have it keep it.
func (u *usecase) Update(ctx context.Context, actor entity.Actor, in UpdateInput) error {
	if err := entity.RequireAdmin(actor); err != nil {
		return err
	}
	if in.Name == nil && in.IsActive == nil {
		return entity.NewValidationError("body", "nothing to update")
	}
	if in.Name != nil {
		name, err := entity.NormalizeSkillName(*in.Name)
		if err != nil {
			return err
		}
		in.Name = &name
	}
	return u.repo.Update(ctx, in.ID, in.Name, in.IsActive)
}

func (u *usecase) Reorder(ctx context.Context, actor entity.Actor, items []entity.SkillOrder) error {
	if err := entity.RequireAdmin(actor); err != nil {
		return err
	}
	if err := entity.ValidateSkillOrder(items); err != nil {
		return err
	}
	return u.repo.Reorder(ctx, items)
}
