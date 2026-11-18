package domain

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
)

// ContactField is a contact detail the employee can edit, one field per request.
type ContactField string

const (
	ContactMobile  ContactField = "mobile"
	ContactEmail   ContactField = "email"
	ContactAddress ContactField = "address"
)

const (
	maxEmailLength   = 254
	maxAddressLength = 500
)

// A mobile number is exactly 10 digits (e.g. 0812345678); the UI shows it as
// 081 234 5678. The CHECK constraint on employees.mobile_phone is a looser
// backstop (9–15 digits, optional +) so older rows stay valid.
var mobilePattern = regexp.MustCompile(`^[0-9]{10}$`)

// emailPattern is a practical address check on top of net/mail: a local part of
// letters, digits and . _ % + - (no leading, trailing or double dots) and a domain
// of dot-separated labels ending in a TLD of 2+ letters. Same rule as the web form.
var emailPattern = regexp.MustCompile(
	`^[a-z0-9_%+-]+(?:\.[a-z0-9_%+-]+)*@(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

var phoneSeparators = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")

func ParseContactField(s string) (ContactField, error) {
	switch f := ContactField(s); f {
	case ContactMobile, ContactEmail, ContactAddress:
		return f, nil
	}
	return "", shared.NewValidationError("field", "unknown contact field")
}

// LogField is the name stored in the change history.
func (f ContactField) LogField() historydomain.FieldName {
	switch f {
	case ContactMobile:
		return historydomain.FieldMobilePhone
	case ContactEmail:
		return historydomain.FieldContactEmail
	default:
		return historydomain.FieldAddress
	}
}

func (f ContactField) Normalize(v string) string {
	v = strings.TrimSpace(v)
	switch f {
	case ContactMobile:
		return phoneSeparators.Replace(v)
	case ContactEmail:
		return strings.ToLower(v)
	}
	return v
}

// Validate expects a normalized value.
func (f ContactField) Validate(v string) error {
	switch f {
	case ContactMobile:
		if !mobilePattern.MatchString(v) {
			return shared.NewValidationError(string(f), "must be 10 digits, for example 0812345678")
		}
	case ContactEmail:
		if len(v) > maxEmailLength {
			return shared.NewValidationError(string(f), "must be at most 254 characters")
		}
		// ParseAddress also accepts `Name <a@b.c>`; require a bare address.
		// Normalize lower-cases it, so the pattern only needs lower case.
		addr, err := mail.ParseAddress(v)
		if err != nil || addr.Address != v || !emailPattern.MatchString(v) {
			return shared.NewValidationError(string(f), "must be a valid email address")
		}
	case ContactAddress:
		n := utf8.RuneCountInString(v)
		if n == 0 {
			return shared.NewValidationError(string(f), "is required")
		}
		if n > maxAddressLength {
			return shared.NewValidationError(string(f), "must be at most 500 characters")
		}
	}
	return nil
}
