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

type fakeTokens struct{ issued entity.Actor }

func (f *fakeTokens) Issue(a entity.Actor) (string, time.Time, error) {
	f.issued = a
	return "token", time.Now().Add(time.Hour), nil
}

func newUsecase(t *testing.T) (Usecase, *fakeTokens) {
	t.Helper()
	hash, err := security.HashPassword("demo1234")
	require.NoError(t, err)
	emp := int64(7)
	repo := &fakeRepo{users: map[string]entity.User{
		"EMP-00001": {ID: 1, EmployeeCode: "EMP-00001", PasswordHash: hash, Role: entity.RoleEmployee, IsActive: true, EmployeeID: &emp},
		"EMP-00009": {ID: 9, EmployeeCode: "EMP-00009", PasswordHash: hash, Role: entity.RoleEmployee, IsActive: false},
	}}
	tokens := &fakeTokens{}
	return NewUsecase(repo, tokens), tokens
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
