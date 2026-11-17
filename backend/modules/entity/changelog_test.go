package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFieldName(t *testing.T) {
	for _, s := range []string{"first_name", "last_name", "mobile_phone", "contact_email", "address", "photo", "skills"} {
		f, err := ParseFieldName(s)
		assert.NoError(t, err, s)
		assert.Equal(t, FieldName(s), f)
	}
	for _, s := range []string{"", "mobile", "email", "password_hash", "SKILLS"} {
		_, err := ParseFieldName(s)
		assert.ErrorIs(t, err, ErrValidation, s)
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
		assert.ErrorIs(t, err, ErrValidation, in)
	}
}

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
