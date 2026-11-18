package skill

import (
	"context"

	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

type UpdateInput struct {
	ID       int64
	Name     *string
	IsActive *bool
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) ListActive(ctx context.Context) ([]skilldomain.Skill, error) {
	return u.repo.ListActive(ctx)
}

func (u *usecase) AdminList(ctx context.Context, actor shared.Actor) ([]skilldomain.SkillUsage, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return nil, err
	}
	return u.repo.ListAll(ctx)
}

func (u *usecase) Create(ctx context.Context, actor shared.Actor, raw string) (int64, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return 0, err
	}
	name, err := skilldomain.NormalizeSkillName(raw)
	if err != nil {
		return 0, err
	}
	return u.repo.Create(ctx, name)
}

// Update renames and/or turns a skill on or off. Skills are never deleted:
// turning one off hides it from the picker, and employees who have it keep it.
func (u *usecase) Update(ctx context.Context, actor shared.Actor, in UpdateInput) error {
	if err := shared.RequireAdmin(actor); err != nil {
		return err
	}
	if in.Name == nil && in.IsActive == nil {
		return shared.NewValidationError("body", "nothing to update")
	}
	if in.Name != nil {
		name, err := skilldomain.NormalizeSkillName(*in.Name)
		if err != nil {
			return err
		}
		in.Name = &name
	}
	return u.repo.Update(ctx, in.ID, in.Name, in.IsActive)
}

func (u *usecase) Reorder(ctx context.Context, actor shared.Actor, items []skilldomain.SkillOrder) error {
	if err := shared.RequireAdmin(actor); err != nil {
		return err
	}
	if err := skilldomain.ValidateSkillOrder(items); err != nil {
		return err
	}
	return u.repo.Reorder(ctx, items)
}
