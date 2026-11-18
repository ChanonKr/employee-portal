package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

var (
	goSkill      = skilldomain.Skill{ID: 1, Name: "Go", IsActive: true}
	reactSkill   = skilldomain.Skill{ID: 2, Name: "React", IsActive: true}
	flutterSkill = skilldomain.Skill{ID: 3, Name: "Flutter", IsActive: false} // turned off by admin
)

func ptr(s string) *string { return &s }

func TestEmployee_ChangeContact(t *testing.T) {
	t.Run("new value is recorded", func(t *testing.T) {
		e := Employee{Version: 3, MobilePhone: ptr("0811111111")}
		c, err := e.ChangeContact(ContactMobile, "081-234-5678", 3)
		require.NoError(t, err)
		assert.Equal(t, &historydomain.Change{Field: historydomain.FieldMobilePhone, Old: "0811111111", New: "0812345678"}, c)
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
		assert.ErrorIs(t, err, shared.ErrConflict)
		assert.Equal(t, "old", *e.Address, "employee must not change on error")
	})

	t.Run("invalid value is reported before version", func(t *testing.T) {
		e := Employee{Version: 5}
		_, err := e.ChangeContact(ContactMobile, "abc", 4)
		assert.ErrorIs(t, err, shared.ErrValidation)
	})
}

func TestEmployee_ChangeSkills(t *testing.T) {
	// Employee has Go + Flutter (Flutter was turned off after they picked it).
	current := []skilldomain.Skill{goSkill, flutterSkill}

	t.Run("records names before and after", func(t *testing.T) {
		e := Employee{Version: 3}
		c, err := e.ChangeSkills([]int64{1, 2}, current, []skilldomain.Skill{goSkill, reactSkill}, 3)
		require.NoError(t, err)
		assert.Equal(t, &historydomain.Change{Field: historydomain.FieldSkills, Old: []string{"Go", "Flutter"}, New: []string{"Go", "React"}}, c)
	})

	t.Run("same set in a different order is not recorded", func(t *testing.T) {
		e := Employee{Version: 3}
		c, err := e.ChangeSkills([]int64{3, 1}, current, []skilldomain.Skill{goSkill, flutterSkill}, 3)
		require.NoError(t, err)
		assert.Nil(t, c)
	})

	t.Run("keeps an inactive skill the employee already has", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{2, 3}, current, []skilldomain.Skill{reactSkill, flutterSkill}, 3)
		assert.NoError(t, err)
	})

	t.Run("cannot newly add an inactive skill", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1, 3}, []skilldomain.Skill{goSkill}, []skilldomain.Skill{goSkill, flutterSkill}, 3)
		assert.ErrorIs(t, err, shared.ErrValidation)
		assert.Contains(t, err.Error(), "Flutter")
	})

	t.Run("unknown id is rejected", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1, 99}, current, []skilldomain.Skill{goSkill}, 3)
		assert.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("empty list clears and stores [] not null", func(t *testing.T) {
		e := Employee{Version: 3}
		c, err := e.ChangeSkills([]int64{}, current, []skilldomain.Skill{}, 3)
		require.NoError(t, err)
		assert.Equal(t, []string{}, c.New)

		c, err = e.ChangeSkills([]int64{1}, nil, []skilldomain.Skill{goSkill}, 3)
		require.NoError(t, err)
		assert.Equal(t, []string{}, c.Old)
	})

	t.Run("stale version is a conflict", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1}, current, []skilldomain.Skill{goSkill}, 2)
		assert.ErrorIs(t, err, shared.ErrConflict)
	})

	t.Run("malformed list is reported before version", func(t *testing.T) {
		e := Employee{Version: 3}
		_, err := e.ChangeSkills([]int64{1, 1}, current, []skilldomain.Skill{goSkill}, 2)
		assert.ErrorIs(t, err, shared.ErrValidation)
	})
}
func TestParseContactField(t *testing.T) {
	for _, s := range []string{"mobile", "email", "address"} {
		f, err := ParseContactField(s)
		assert.NoError(t, err)
		assert.Equal(t, ContactField(s), f)
	}
	for _, s := range []string{"", "first_name", "mobile_phone", "Mobile"} {
		_, err := ParseContactField(s)
		assert.ErrorIs(t, err, shared.ErrValidation, s)
	}
}

func TestContactField_NormalizeAndValidate(t *testing.T) {
	tests := []struct {
		name    string
		field   ContactField
		input   string
		want    string
		wantErr bool
	}{
		{"mobile plain", ContactMobile, "0812345678", "0812345678", false},
		{"mobile with separators", ContactMobile, " 081-234 5678 ", "0812345678", false},
		{"mobile international", ContactMobile, "+66812345678", "+66812345678", false},
		{"mobile too short", ContactMobile, "08123", "08123", true},
		{"mobile letters", ContactMobile, "08123abcde", "08123abcde", true},
		{"mobile empty", ContactMobile, "  ", "", true},

		{"email lowercased", ContactEmail, "  Somchai@Example.COM ", "somchai@example.com", false},
		{"email display name rejected", ContactEmail, "Somchai <s@example.com>", "somchai <s@example.com>", true},
		{"email missing at", ContactEmail, "somchai.example.com", "somchai.example.com", true},
		{"email empty", ContactEmail, "", "", true},
		{"email too long", ContactEmail, strings.Repeat("a", 250) + "@x.co", strings.Repeat("a", 250) + "@x.co", true},

		{"address trimmed", ContactAddress, "  99 Road  ", "99 Road", false},
		{"address thai counts runes", ContactAddress, strings.Repeat("ก", 500), strings.Repeat("ก", 500), false},
		{"address too long", ContactAddress, strings.Repeat("a", 501), strings.Repeat("a", 501), true},
		{"address empty", ContactAddress, "   ", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.field.Normalize(tt.input)
			assert.Equal(t, tt.want, got)

			err := tt.field.Validate(got)
			if tt.wantErr {
				var ve *shared.ValidationError
				assert.True(t, errors.As(err, &ve))
				assert.Contains(t, ve.Fields, string(tt.field))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestEmployee_ChangeName(t *testing.T) {
	t.Run("records each changed part", func(t *testing.T) {
		e := Employee{FirstName: "Somchai", LastName: "Jaidi", Version: 2}
		changes, err := e.ChangeName(" Somchai ", "Jaidee", 2)
		require.NoError(t, err)
		assert.Equal(t, []historydomain.Change{{Field: historydomain.FieldLastName, Old: "Jaidi", New: "Jaidee"}}, changes)
		assert.Equal(t, "Jaidee", e.LastName)
	})

	t.Run("both parts", func(t *testing.T) {
		e := Employee{FirstName: "A", LastName: "B", Version: 1}
		changes, err := e.ChangeName("C", "D", 1)
		require.NoError(t, err)
		assert.Len(t, changes, 2)
	})

	t.Run("no change returns an empty list", func(t *testing.T) {
		e := Employee{FirstName: "Suda", LastName: "Rakdee", Version: 1}
		changes, err := e.ChangeName("Suda", "  Rakdee", 1)
		require.NoError(t, err)
		assert.Empty(t, changes)
	})

	t.Run("validation and version", func(t *testing.T) {
		e := Employee{FirstName: "A", LastName: "B", Version: 3}
		_, err := e.ChangeName("", "B", 3)
		assert.ErrorIs(t, err, shared.ErrValidation)
		_, err = e.ChangeName("A", strings.Repeat("x", 101), 3)
		assert.ErrorIs(t, err, shared.ErrValidation)
		_, err = e.ChangeName("New", "B", 2)
		assert.ErrorIs(t, err, shared.ErrConflict)
		assert.Equal(t, "A", e.FirstName, "unchanged on error")
	})
}
