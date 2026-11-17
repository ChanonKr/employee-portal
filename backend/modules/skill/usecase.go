package skill

import (
	"context"

	"employee-portal/modules/entity"
)

// Repository is implemented in repository.go.
type Repository interface {
	ListActive(ctx context.Context) ([]entity.Skill, error)
}

type Usecase interface {
	// ListActive returns the catalog shown in the skills picker.
	ListActive(ctx context.Context) ([]entity.Skill, error)
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
