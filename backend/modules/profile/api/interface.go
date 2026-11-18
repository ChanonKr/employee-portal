package api

import (
	"context"

	"employee-portal/modules/profile/domain"
	"employee-portal/modules/profile/service"
	"employee-portal/modules/shared"
)

// Service is what the HTTP layer calls; service.Service implements it.
type Service interface {
	GetProfile(ctx context.Context, actor shared.Actor) (domain.Profile, error)
	UpdateContact(ctx context.Context, in service.UpdateContactInput) (newVersion int32, err error)
	SetSkills(ctx context.Context, in service.SetSkillsInput) (newVersion int32, err error)
	UpdatePhoto(ctx context.Context, in service.UpdatePhotoInput) (newVersion int32, photoName string, err error)
	UpdateName(ctx context.Context, in service.UpdateNameInput) (newVersion int32, err error)
}

var _ Service = (*service.Service)(nil)
