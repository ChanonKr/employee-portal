package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPageRequest(t *testing.T) {
	p, err := NewPageRequest(0, 0)
	require.NoError(t, err)
	assert.Equal(t, PageRequest{Page: 1, PageSize: DefaultPageSize}, p)
	assert.Equal(t, 0, p.Offset())

	p, err = NewPageRequest(3, 10)
	require.NoError(t, err)
	assert.Equal(t, 20, p.Offset())

	for _, tc := range [][2]int{{-1, 10}, {1, -1}, {1, MaxPageSize + 1}} {
		_, err := NewPageRequest(tc[0], tc[1])
		assert.ErrorIs(t, err, ErrValidation, tc)
	}
}

func TestRequireAdmin(t *testing.T) {
	assert.NoError(t, RequireAdmin(Actor{Role: RoleAdmin}))
	assert.ErrorIs(t, RequireAdmin(Actor{Role: RoleEmployee}), ErrForbidden)
}
