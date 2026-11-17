package auth

import (
	"context"
	"time"

	"employee-portal/modules/entity"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go.
type Repository interface {
	GetUserByCode(ctx context.Context, code string) (entity.User, error)
	GetUserByID(ctx context.Context, id int64) (entity.User, error)
	// CreateAccount inserts the user and the empty profile in one transaction.
	// entity.ErrDuplicate when the employee code is taken.
	CreateAccount(ctx context.Context, a entity.NewAccount, passwordHash string) (userID, employeeID int64, err error)
	SetPassword(ctx context.Context, userID int64, passwordHash string, mustChange bool) error
	UserIDByEmployee(ctx context.Context, employeeID int64) (int64, error)
}

type TokenIssuer interface {
	Issue(a entity.Actor) (token string, expiresAt time.Time, err error)
}

type Usecase interface {
	Login(ctx context.Context, employeeCode, password string) (Session, error)
	Me(ctx context.Context, actor entity.Actor) (entity.User, error)
	// ChangePassword replaces the password and returns a fresh session without the
	// must-change flag.
	ChangePassword(ctx context.Context, actor entity.Actor, current, next string) (Session, error)

	// Admin
	CreateAccount(ctx context.Context, actor entity.Actor, in CreateAccountInput) (CreatedAccount, error)
	ResetPassword(ctx context.Context, actor entity.Actor, employeeID int64) (temporaryPassword string, err error)
}
