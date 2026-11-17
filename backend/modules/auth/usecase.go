package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"employee-portal/modules/entity"
	"employee-portal/pkg/security"
)

// Repository is implemented in repository.go.
type Repository interface {
	GetUserByCode(ctx context.Context, code string) (entity.User, error)
	GetUserByID(ctx context.Context, id int64) (entity.User, error)
}

type TokenIssuer interface {
	Issue(a entity.Actor) (token string, expiresAt time.Time, err error)
}

type Session struct {
	Token     string
	ExpiresAt time.Time
	User      entity.User
}

type Usecase interface {
	Login(ctx context.Context, employeeCode, password string) (Session, error)
	Me(ctx context.Context, actor entity.Actor) (entity.User, error)
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
	code := strings.ToUpper(strings.TrimSpace(employeeCode))

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

	actor := entity.Actor{UserID: user.ID, Role: user.Role}
	if user.EmployeeID != nil {
		actor.EmployeeID = *user.EmployeeID
	}
	token, exp, err := u.tokens.Issue(actor)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, ExpiresAt: exp, User: user}, nil
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
