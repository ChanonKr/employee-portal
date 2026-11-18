package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/shared"
)

func TestParseFieldName(t *testing.T) {
	for _, s := range []string{"first_name", "last_name", "mobile_phone", "contact_email", "address", "photo", "skills"} {
		f, err := ParseFieldName(s)
		assert.NoError(t, err, s)
		assert.Equal(t, FieldName(s), f)
	}
	for _, s := range []string{"", "mobile", "email", "password_hash", "SKILLS"} {
		_, err := ParseFieldName(s)
		assert.ErrorIs(t, err, shared.ErrValidation, s)
	}
}

func TestParseHistoryDays(t *testing.T) {
	for in, want := range map[string]int{"1": 1, "7": 7, " 30 ": 30, "365": 365} {
		got, err := ParseHistoryDays(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got)
	}
	for _, in := range []string{"0", "366", "-1", "7.5", "seven", ""} {
		_, err := ParseHistoryDays(in)
		assert.ErrorIs(t, err, shared.ErrValidation, in)
	}
}
