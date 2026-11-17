package auth

import (
	"context"
	"errors"
	"time"

	"employee-portal/modules/entity"
	"employee-portal/pkg/security"
)

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

type Session struct {
	Token     string
	ExpiresAt time.Time
	User      entity.User
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

type usecase struct {
	repo   Repository
	tokens TokenIssuer
}

func NewUsecase(repo Repository, tokens TokenIssuer) Usecase {
	return &usecase{repo: repo, tokens: tokens}
}

// Login returns the same ErrUnauthorized for an unknown code, a wrong password and
// a disabled account, so callers cannot tell which employee codes exist.
func (u *usecase) Login(ctx context.Context, employeeCode, password string) (Session, error) {
	code, err := entity.NormalizeEmployeeCode(employeeCode)
	if err != nil {
		security.BurnPasswordCheck(password)
		return Session{}, entity.ErrUnauthorized
	}

	user, err := u.repo.GetUserByCode(ctx, code)
	if errors.Is(err, entity.ErrNotFound) {
		security.BurnPasswordCheck(password)
		return Session{}, entity.ErrUnauthorized
	}
	if err != nil {
		return Session{}, err
	}
	if !security.CheckPassword(user.PasswordHash, password) || !user.IsActive {
		return Session{}, entity.ErrUnauthorized
	}
	return u.issue(user)
}

// Me re-reads the user so a disabled account loses access on the next app load.
func (u *usecase) Me(ctx context.Context, actor entity.Actor) (entity.User, error) {
	user, err := u.repo.GetUserByID(ctx, actor.UserID)
	if errors.Is(err, entity.ErrNotFound) {
		return entity.User{}, entity.ErrUnauthorized
	}
	if err != nil {
		return entity.User{}, err
	}
	if !user.IsActive {
		return entity.User{}, entity.ErrUnauthorized
	}
	return user, nil
}

func (u *usecase) ChangePassword(ctx context.Context, actor entity.Actor, current, next string) (Session, error) {
	user, err := u.Me(ctx, actor)
	if err != nil {
		return Session{}, err
	}
	// 400, not 401: a wrong current password must not look like an expired session.
	if !security.CheckPassword(user.PasswordHash, current) {
		return Session{}, entity.NewValidationError("currentPassword", "is incorrect")
	}
	if err := entity.ValidateNewPassword("newPassword", next); err != nil {
		return Session{}, err
	}
	if current == next {
		return Session{}, entity.NewValidationError("newPassword", "must be different from the current password")
	}

	hash, err := security.HashPassword(next)
	if err != nil {
		return Session{}, err
	}
	if err := u.repo.SetPassword(ctx, user.ID, hash, false); err != nil {
		return Session{}, err
	}
	user.MustChangePassword = false
	return u.issue(user)
}

func (u *usecase) CreateAccount(ctx context.Context, actor entity.Actor, in CreateAccountInput) (CreatedAccount, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return CreatedAccount{}, err
	}
	acc, err := entity.NormalizeNewAccount(in.EmployeeCode, in.FirstName, in.LastName, in.Role)
	if err != nil {
		return CreatedAccount{}, err
	}
	temp, hash, err := newTempPassword()
	if err != nil {
		return CreatedAccount{}, err
	}
	userID, empID, err := u.repo.CreateAccount(ctx, acc, hash)
	if err != nil {
		return CreatedAccount{}, err
	}
	return CreatedAccount{UserID: userID, EmployeeID: empID, EmployeeCode: acc.EmployeeCode, TemporaryPassword: temp}, nil
}

// ResetPassword sets a new temporary password that must be changed at next sign-in.
func (u *usecase) ResetPassword(ctx context.Context, actor entity.Actor, employeeID int64) (string, error) {
	if err := entity.RequireAdmin(actor); err != nil {
		return "", err
	}
	userID, err := u.repo.UserIDByEmployee(ctx, employeeID)
	if err != nil {
		return "", err
	}
	temp, hash, err := newTempPassword()
	if err != nil {
		return "", err
	}
	if err := u.repo.SetPassword(ctx, userID, hash, true); err != nil {
		return "", err
	}
	return temp, nil
}

func (u *usecase) issue(user entity.User) (Session, error) {
	actor := entity.Actor{UserID: user.ID, Role: user.Role, MustChangePassword: user.MustChangePassword}
	if user.EmployeeID != nil {
		actor.EmployeeID = *user.EmployeeID
	}
	token, exp, err := u.tokens.Issue(actor)
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
