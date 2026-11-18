package service

import (
	"context"
	"time"

	"employee-portal/modules/auth/domain"
	"employee-portal/modules/shared"
)

// Repository is the port this service needs; repo.Repository implements it.
type Repository interface {
	GetUserByCode(ctx context.Context, code string) (domain.User, error)
	GetUserByID(ctx context.Context, id int64) (domain.User, error)
	// CreateAccount inserts the user and the empty profile in one transaction.
	// shared.ErrDuplicate when the employee code is taken.
	CreateAccount(ctx context.Context, a domain.NewAccount, passwordHash string) (userID, employeeID int64, err error)
	SetPassword(ctx context.Context, userID int64, passwordHash string, mustChange bool) error
	UserIDByEmployee(ctx context.Context, employeeID int64) (int64, error)
}

type TokenIssuer interface {
	Issue(a shared.Actor) (token string, expiresAt time.Time, err error)
}
