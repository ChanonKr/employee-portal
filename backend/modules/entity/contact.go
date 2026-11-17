package entity

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
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

// Same rule as the CHECK constraint on employees.mobile_phone.
var mobilePattern = regexp.MustCompile(`^\+?[0-9]{9,15}$`)

var phoneSeparators = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")

func ParseContactField(s string) (ContactField, error) {
	switch f := ContactField(s); f {
	case ContactMobile, ContactEmail, ContactAddress:
		return f, nil
	}
	return "", NewValidationError("field", "unknown contact field")
}

// LogField is the name stored in the change history.
func (f ContactField) LogField() FieldName {
	switch f {
	case ContactMobile:
		return FieldMobilePhone
	case ContactEmail:
		return FieldContactEmail
	default:
		return FieldAddress
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
			return NewValidationError(string(f), "must be 9–15 digits, optionally starting with +")
		}
	case ContactEmail:
		if len(v) > maxEmailLength {
			return NewValidationError(string(f), "must be at most 254 characters")
		}
		// ParseAddress also accepts `Name <a@b.c>`; require a bare address.
		addr, err := mail.ParseAddress(v)
		if err != nil || addr.Address != v {
			return NewValidationError(string(f), "must be a valid email address")
		}
	case ContactAddress:
		n := utf8.RuneCountInString(v)
		if n == 0 {
			return NewValidationError(string(f), "is required")
		}
		if n > maxAddressLength {
			return NewValidationError(string(f), "must be at most 500 characters")
		}
	}
	return nil
}
