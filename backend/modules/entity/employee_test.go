package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

func TestEmployee_ChangeContact(t *testing.T) {
	t.Run("new value is recorded", func(t *testing.T) {
		e := Employee{Version: 3, MobilePhone: ptr("0811111111")}
		c, err := e.ChangeContact(ContactMobile, "081-234-5678", 3)
		require.NoError(t, err)
		assert.Equal(t, &Change{Field: FieldMobilePhone, Old: "0811111111", New: "0812345678"}, c)
		assert.Equal(t, "0812345678", *e.MobilePhone)
	})

	t.Run("first value has nil old", func(t *testing.T) {
		e := Employee{Version: 1}
		c, err := e.ChangeContact(ContactAddress, "99 Road", 1)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Nil(t, c.Old) // untyped nil, so it is stored as SQL NULL
		assert.Equal(t, "99 Road", c.New)
	})

	t.Run("same value after normalization is not recorded", func(t *testing.T) {
		e := Employee{Version: 1, ContactEmail: ptr("a@example.com")}
		c, err := e.ChangeContact(ContactEmail, " A@Example.com ", 1)
		require.NoError(t, err)
		assert.Nil(t, c)
	})

	t.Run("stale version is a conflict", func(t *testing.T) {
		e := Employee{Version: 5, Address: ptr("old")}
		_, err := e.ChangeContact(ContactAddress, "new", 4)
		assert.ErrorIs(t, err, ErrConflict)
		assert.Equal(t, "old", *e.Address, "entity must not change on error")
	})

	t.Run("invalid value is reported before version", func(t *testing.T) {
		e := Employee{Version: 5}
		_, err := e.ChangeContact(ContactMobile, "abc", 4)
		assert.ErrorIs(t, err, ErrValidation)
	})
}
