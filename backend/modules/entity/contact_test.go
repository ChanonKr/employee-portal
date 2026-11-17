package entity

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseContactField(t *testing.T) {
	for _, s := range []string{"mobile", "email", "address"} {
		f, err := ParseContactField(s)
		assert.NoError(t, err)
		assert.Equal(t, ContactField(s), f)
	}
	for _, s := range []string{"", "first_name", "mobile_phone", "Mobile"} {
		_, err := ParseContactField(s)
		assert.ErrorIs(t, err, ErrValidation, s)
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
				var ve *ValidationError
				assert.True(t, errors.As(err, &ve))
				assert.Contains(t, ve.Fields, string(tt.field))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
