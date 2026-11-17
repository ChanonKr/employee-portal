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

func TestEmployee_ChangeSkills(t *testing.T) {
	// Employee has Go + Flutter (Flutter was turned off after they picked it).
	current := []Skill{goSkill, flutterSkill}

	t.Run("records names before and after", func(t *testing.T) {
		e := Employee{Version: 3}
		c, err := e.ChangeSkills([]int64{1, 2}, current, []Skill{goSkill, reactSkill}, 3)
		require.NoError(t, err)
		assert.Equal(t, &Change{Field: FieldSkills, Old: []string{"Go", "Flutter"}, New: []string{"Go", "React"}}, c)
	})

	t.Run("same set in a different order is not recorded", func(t *testing.T) {
		e := Employee{Version: 3}
		c, err := e.ChangeSkills([]int64{3, 1}, current, []Skill{goSkill, flutterSkill}, 3)
		require.NoError(t, err)
		assert.Nil(t, c)
	})

	t.Run("keeps an inactive skill the employee already has", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{2, 3}, current, []Skill{reactSkill, flutterSkill}, 3)
		assert.NoError(t, err)
	})

	t.Run("cannot newly add an inactive skill", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1, 3}, []Skill{goSkill}, []Skill{goSkill, flutterSkill}, 3)
		assert.ErrorIs(t, err, ErrValidation)
		assert.Contains(t, err.Error(), "Flutter")
	})

	t.Run("unknown id is rejected", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1, 99}, current, []Skill{goSkill}, 3)
		assert.ErrorIs(t, err, ErrValidation)
	})

	t.Run("empty list clears and stores [] not null", func(t *testing.T) {
		e := Employee{Version: 3}
		c, err := e.ChangeSkills([]int64{}, current, []Skill{}, 3)
		require.NoError(t, err)
		assert.Equal(t, []string{}, c.New)

		c, err = e.ChangeSkills([]int64{1}, nil, []Skill{goSkill}, 3)
		require.NoError(t, err)
		assert.Equal(t, []string{}, c.Old)
	})

	t.Run("stale version is a conflict", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1}, current, []Skill{goSkill}, 2)
		assert.ErrorIs(t, err, ErrConflict)
	})

	t.Run("malformed list is reported before version", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1, 1}, current, []Skill{goSkill}, 2)
		assert.ErrorIs(t, err, ErrValidation)
	})
}
