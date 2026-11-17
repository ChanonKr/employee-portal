package entity

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// ErrPasswordChangeRequired blocks every API except /auth/* until the user
// replaces a temporary password.
var ErrPasswordChangeRequired = errors.New("password change required")

var employeeCodePattern = regexp.MustCompile(`^[A-Z0-9-]{3,20}$`)

// NormalizeEmployeeCode upper-cases and validates a sign-in code (e.g. EMP-00042).
func NormalizeEmployeeCode(raw string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if !employeeCodePattern.MatchString(code) {
		return "", NewValidationError("employeeCode", "must be 3–20 letters, digits or dashes")
	}
	return code, nil
}

func ParseRole(s string) (Role, error) {
	switch r := Role(s); r {
	case RoleEmployee, RoleAdmin:
		return r, nil
	case "":
		return RoleEmployee, nil
	}
	return "", NewValidationError("role", "must be employee or admin")
}

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72 // bcrypt ignores anything longer
)

// ValidateNewPassword is the password policy: 8+ characters, at most 72 bytes,
// at least one letter and one digit.
func ValidateNewPassword(field, pw string) error {
	if len([]rune(pw)) < minPasswordLength {
		return NewValidationError(field, "must be at least 8 characters")
	}
	if len(pw) > maxPasswordBytes {
		return NewValidationError(field, "must be at most 72 bytes")
	}
	var letter, digit bool
	for _, r := range pw {
		letter = letter || unicode.IsLetter(r)
		digit = digit || unicode.IsDigit(r)
	}
	if !letter || !digit {
		return NewValidationError(field, "must contain a letter and a digit")
	}
	return nil
}

// NewAccount is what an admin enters to create an employee account.
type NewAccount struct {
	EmployeeCode string
	FirstName    string
	LastName     string
	Role         Role
}

// NormalizeNewAccount validates every field and returns the cleaned values.
func NormalizeNewAccount(code, first, last, role string) (NewAccount, error) {
	var a NewAccount
	var err error
	if a.EmployeeCode, err = NormalizeEmployeeCode(code); err != nil {
		return a, err
	}
	if a.FirstName, err = normalizeName("firstName", first); err != nil {
		return a, err
	}
	if a.LastName, err = normalizeName("lastName", last); err != nil {
		return a, err
	}
	if a.Role, err = ParseRole(role); err != nil {
		return a, err
	}
	return a, nil
}
