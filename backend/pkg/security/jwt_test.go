package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authdomain "employee-portal/modules/auth/domain"
	"employee-portal/modules/shared"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestTokenManager_RoundTrip(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour)
	want := shared.Actor{UserID: 5, EmployeeID: 9, Role: shared.RoleAdmin, MustChangePassword: true}

	tok, exp, err := m.Issue(want)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), exp, time.Second)

	got, err := m.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestTokenManager_Rejects(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour)
	tok, _, err := m.Issue(shared.Actor{UserID: 1, Role: shared.RoleEmployee})
	require.NoError(t, err)

	t.Run("wrong secret", func(t *testing.T) {
		other := NewTokenManager("ffffffffffffffffffffffffffffffff", time.Hour)
		_, err := other.Parse(tok)
		assert.ErrorIs(t, err, shared.ErrUnauthorized)
	})

	t.Run("expired", func(t *testing.T) {
		late := NewTokenManager(testSecret, time.Hour)
		late.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
		_, err := late.Parse(tok)
		assert.ErrorIs(t, err, shared.ErrUnauthorized)
	})

	t.Run("garbage", func(t *testing.T) {
		_, err := m.Parse("not-a-token")
		assert.ErrorIs(t, err, shared.ErrUnauthorized)
	})
}

func TestPassword(t *testing.T) {
	h, err := HashPassword("demo1234")
	require.NoError(t, err)
	assert.True(t, CheckPassword(h, "demo1234"))
	assert.False(t, CheckPassword(h, "wrong"))
}

func TestTempPassword(t *testing.T) {
	seen := map[string]bool{}
	for range 50 {
		p, err := TempPassword()
		require.NoError(t, err)
		assert.Len(t, p, 12)
		assert.NoError(t, authdomain.ValidateNewPassword("p", p), p)
		assert.NotContains(t, p, "0")
		assert.NotContains(t, p, "l")
		seen[p] = true
	}
	assert.Len(t, seen, 50, "no repeats")
}
