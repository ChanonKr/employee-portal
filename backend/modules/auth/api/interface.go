package api

import (
	"context"

	"employee-portal/modules/auth/domain"
	"employee-portal/modules/auth/service"
	"employee-portal/modules/shared"
)

// Service is what the HTTP layer calls; service.Service implements it.
type Service interface {
	Login(ctx context.Context, employeeCode, password string) (service.Session, error)
	Me(ctx context.Context, actor shared.Actor) (domain.User, error)
	// ChangePassword replaces the password and returns a fresh session without the
	// must-change flag.
	ChangePassword(ctx context.Context, actor shared.Actor, current, next string) (service.Session, error)

	// Admin
	CreateAccount(ctx context.Context, actor shared.Actor, in service.CreateAccountInput) (service.CreatedAccount, error)
	ResetPassword(ctx context.Context, actor shared.Actor, employeeID int64) (temporaryPassword string, err error)
}

var _ Service = (*service.Service)(nil)
