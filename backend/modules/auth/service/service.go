package service

import (
	"context"
	"errors"
	"time"

	"employee-portal/modules/auth/domain"
	"employee-portal/modules/shared"
	"employee-portal/pkg/security"
)

type Session struct {
	Token     string
	ExpiresAt time.Time
	User      domain.User
}

type CreateAccountInput struct {
	EmployeeCode string
	FirstName    string
	LastName     string
	Role         string // "" = employee
}

type CreatedAccount struct {
	UserID            int64
	EmployeeID        int64
	EmployeeCode      string
	TemporaryPassword string // shown to the admin once; never stored in plain text
}

// Service runs this feature's use cases. It never imports GORM or Gin.
type Service struct {
	repo   Repository
	tokens TokenIssuer
}

func New(repo Repository, tokens TokenIssuer) *Service {
	return &Service{repo: repo, tokens: tokens}
}

// Login returns the same ErrUnauthorized for an unknown code, a wrong password and
// a disabled account, so callers cannot tell which employee codes exist.
func (svc *Service) Login(ctx context.Context, employeeCode, password string) (Session, error) {
	code, err := domain.NormalizeEmployeeCode(employeeCode)
	if err != nil {
		security.BurnPasswordCheck(password)
		return Session{}, shared.ErrUnauthorized
	}

	user, err := svc.repo.GetUserByCode(ctx, code)
	if errors.Is(err, shared.ErrNotFound) {
		security.BurnPasswordCheck(password)
		return Session{}, shared.ErrUnauthorized
	}
	if err != nil {
		return Session{}, err
	}
	if !security.CheckPassword(user.PasswordHash, password) || !user.IsActive {
		return Session{}, shared.ErrUnauthorized
	}
	return svc.issue(user)
}

// Me re-reads the user so a disabled account loses access on the next app load.
func (svc *Service) Me(ctx context.Context, actor shared.Actor) (domain.User, error) {
	user, err := svc.repo.GetUserByID(ctx, actor.UserID)
	if errors.Is(err, shared.ErrNotFound) {
		return domain.User{}, shared.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, err
	}
	if !user.IsActive {
		return domain.User{}, shared.ErrUnauthorized
	}
	return user, nil
}

func (svc *Service) ChangePassword(ctx context.Context, actor shared.Actor, current, next string) (Session, error) {
	user, err := svc.Me(ctx, actor)
	if err != nil {
		return Session{}, err
	}
	// 400, not 401: a wrong current password must not look like an expired session.
	if !security.CheckPassword(user.PasswordHash, current) {
		return Session{}, shared.NewValidationError("currentPassword", "is incorrect")
	}
	if err := domain.ValidateNewPassword("newPassword", next); err != nil {
		return Session{}, err
	}
	if current == next {
		return Session{}, shared.NewValidationError("newPassword", "must be different from the current password")
	}

	hash, err := security.HashPassword(next)
	if err != nil {
		return Session{}, err
	}
	if err := svc.repo.SetPassword(ctx, user.ID, hash, false); err != nil {
		return Session{}, err
	}
	user.MustChangePassword = false
	return svc.issue(user)
}

func (svc *Service) CreateAccount(ctx context.Context, actor shared.Actor, in CreateAccountInput) (CreatedAccount, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return CreatedAccount{}, err
	}
	acc, err := domain.NormalizeNewAccount(in.EmployeeCode, in.FirstName, in.LastName, in.Role)
	if err != nil {
		return CreatedAccount{}, err
	}
	temp, hash, err := newTempPassword()
	if err != nil {
		return CreatedAccount{}, err
	}
	userID, empID, err := svc.repo.CreateAccount(ctx, acc, hash)
	if err != nil {
		return CreatedAccount{}, err
	}
	return CreatedAccount{UserID: userID, EmployeeID: empID, EmployeeCode: acc.EmployeeCode, TemporaryPassword: temp}, nil
}

// ResetPassword sets a new temporary password that must be changed at next sign-in.
func (svc *Service) ResetPassword(ctx context.Context, actor shared.Actor, employeeID int64) (string, error) {
	if err := shared.RequireAdmin(actor); err != nil {
		return "", err
	}
	userID, err := svc.repo.UserIDByEmployee(ctx, employeeID)
	if err != nil {
		return "", err
	}
	temp, hash, err := newTempPassword()
	if err != nil {
		return "", err
	}
	if err := svc.repo.SetPassword(ctx, userID, hash, true); err != nil {
		return "", err
	}
	return temp, nil
}

func (svc *Service) issue(user domain.User) (Session, error) {
	actor := shared.Actor{UserID: user.ID, Role: user.Role, MustChangePassword: user.MustChangePassword}
	if user.EmployeeID != nil {
		actor.EmployeeID = *user.EmployeeID
	}
	token, exp, err := svc.tokens.Issue(actor)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, ExpiresAt: exp, User: user}, nil
}

func newTempPassword() (plain, hash string, err error) {
	if plain, err = security.TempPassword(); err != nil {
		return "", "", err
	}
	if hash, err = security.HashPassword(plain); err != nil {
		return "", "", err
	}
	return plain, hash, nil
}
