package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/entity"
	"employee-portal/pkg/security"
)

type fakeRepo struct{ users map[string]entity.User }

func (f *fakeRepo) GetUserByCode(_ context.Context, code string) (entity.User, error) {
	u, ok := f.users[code]
	if !ok {
		return entity.User{}, entity.ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) GetUserByID(_ context.Context, id int64) (entity.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return entity.User{}, entity.ErrNotFound
}

func (f *fakeRepo) CreateAccount(_ context.Context, a entity.NewAccount, hash string) (int64, int64, error) {
	if _, taken := f.users[a.EmployeeCode]; taken {
		return 0, 0, entity.ErrDuplicate
	}
	emp := int64(100 + len(f.users))
	f.users[a.EmployeeCode] = entity.User{ID: int64(len(f.users) + 1), EmployeeCode: a.EmployeeCode,
		PasswordHash: hash, Role: a.Role, IsActive: true, MustChangePassword: true, EmployeeID: &emp}
	return int64(len(f.users)), emp, nil
}

func (f *fakeRepo) SetPassword(_ context.Context, userID int64, hash string, mustChange bool) error {
	for code, u := range f.users {
		if u.ID == userID {
			u.PasswordHash, u.MustChangePassword = hash, mustChange
			f.users[code] = u
			return nil
		}
	}
	return entity.ErrNotFound
}

func (f *fakeRepo) UserIDByEmployee(_ context.Context, employeeID int64) (int64, error) {
	for _, u := range f.users {
		if u.EmployeeID != nil && *u.EmployeeID == employeeID {
			return u.ID, nil
		}
	}
	return 0, entity.ErrNotFound
}

type fakeTokens struct{ issued entity.Actor }

func (f *fakeTokens) Issue(a entity.Actor) (string, time.Time, error) {
	f.issued = a
	return "token", time.Now().Add(time.Hour), nil
}

func newUsecase(t *testing.T) (Usecase, *fakeTokens) {
	uc, tokens, _ := newUsecaseWithRepo(t)
	return uc, tokens
}

func newUsecaseWithRepo(t *testing.T) (Usecase, *fakeTokens, *fakeRepo) {
	t.Helper()
	hash, err := security.HashPassword("demo1234")
	require.NoError(t, err)
	emp := int64(7)
	repo := &fakeRepo{users: map[string]entity.User{
		"EMP-00001": {ID: 1, EmployeeCode: "EMP-00001", PasswordHash: hash, Role: entity.RoleEmployee, IsActive: true, EmployeeID: &emp},
		"EMP-00009": {ID: 9, EmployeeCode: "EMP-00009", PasswordHash: hash, Role: entity.RoleEmployee, IsActive: false},
	}}
	tokens := &fakeTokens{}
	return NewUsecase(repo, tokens), tokens, repo
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("success puts user and employee id in the token", func(t *testing.T) {
		uc, tokens := newUsecase(t)
		s, err := uc.Login(ctx, " emp-00001 ", "demo1234")
		require.NoError(t, err)
		assert.Equal(t, "token", s.Token)
		assert.Equal(t, entity.Actor{UserID: 1, EmployeeID: 7, Role: entity.RoleEmployee}, tokens.issued)
	})

	for name, tc := range map[string]struct{ code, password string }{
		"wrong password":   {"EMP-00001", "nope"},
		"unknown code":     {"EMP-99999", "demo1234"},
		"inactive account": {"EMP-00009", "demo1234"},
	} {
		t.Run(name+" is the same unauthorized error", func(t *testing.T) {
			uc, _ := newUsecase(t)
			_, err := uc.Login(ctx, tc.code, tc.password)
			assert.ErrorIs(t, err, entity.ErrUnauthorized)
		})
	}
}

func TestMe_InactiveUserLosesAccess(t *testing.T) {
	uc, _ := newUsecase(t)
	_, err := uc.Me(context.Background(), entity.Actor{UserID: 9})
	assert.ErrorIs(t, err, entity.ErrUnauthorized)
}

var admin = entity.Actor{UserID: 50, Role: entity.RoleAdmin}

func TestCreateAccount_ThenFirstLoginMustChangePassword(t *testing.T) {
	ctx := context.Background()
	uc, tokens, _ := newUsecaseWithRepo(t)

	acc, err := uc.CreateAccount(ctx, admin, CreateAccountInput{EmployeeCode: " emp-00077 ", FirstName: "New", LastName: "Hire"})
	require.NoError(t, err)
	assert.Equal(t, "EMP-00077", acc.EmployeeCode)
	assert.Len(t, acc.TemporaryPassword, 12)

	s, err := uc.Login(ctx, "EMP-00077", acc.TemporaryPassword)
	require.NoError(t, err)
	assert.True(t, tokens.issued.MustChangePassword, "temporary password → flag in the token")

	// Wrong current password is a 400-style validation error, not 401.
	actor := tokens.issued
	_, err = uc.ChangePassword(ctx, actor, "nope", "brandnew123")
	assert.ErrorIs(t, err, entity.ErrValidation)
	_, err = uc.ChangePassword(ctx, actor, acc.TemporaryPassword, "short")
	assert.ErrorIs(t, err, entity.ErrValidation)
	_, err = uc.ChangePassword(ctx, actor, acc.TemporaryPassword, acc.TemporaryPassword)
	assert.ErrorIs(t, err, entity.ErrValidation, "must differ")

	s, err = uc.ChangePassword(ctx, actor, acc.TemporaryPassword, "brandnew123")
	require.NoError(t, err)
	assert.NotEmpty(t, s.Token)
	assert.False(t, tokens.issued.MustChangePassword, "new session without the flag")

	_, err = uc.Login(ctx, "EMP-00077", acc.TemporaryPassword)
	assert.ErrorIs(t, err, entity.ErrUnauthorized, "old password no longer works")
	_, err = uc.Login(ctx, "EMP-00077", "brandnew123")
	assert.NoError(t, err)

	_, err = uc.CreateAccount(ctx, admin, CreateAccountInput{EmployeeCode: "EMP-00077", FirstName: "A", LastName: "B"})
	assert.ErrorIs(t, err, entity.ErrDuplicate)
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	uc, tokens, _ := newUsecaseWithRepo(t)

	temp, err := uc.ResetPassword(ctx, admin, 7) // EMP-00001 has employee id 7
	require.NoError(t, err)
	_, err = uc.Login(ctx, "EMP-00001", "demo1234")
	assert.ErrorIs(t, err, entity.ErrUnauthorized)
	_, err = uc.Login(ctx, "EMP-00001", temp)
	require.NoError(t, err)
	assert.True(t, tokens.issued.MustChangePassword)

	_, err = uc.ResetPassword(ctx, admin, 999)
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestAccountAdminRequiresAdmin(t *testing.T) {
	ctx := context.Background()
	uc, _, repo := newUsecaseWithRepo(t)
	employee := entity.Actor{UserID: 1, EmployeeID: 7, Role: entity.RoleEmployee}

	_, err := uc.CreateAccount(ctx, employee, CreateAccountInput{EmployeeCode: "EMP-5", FirstName: "A", LastName: "B"})
	assert.ErrorIs(t, err, entity.ErrForbidden)
	_, err = uc.ResetPassword(ctx, employee, 7)
	assert.ErrorIs(t, err, entity.ErrForbidden)
	assert.Len(t, repo.users, 2, "nothing created")
}
