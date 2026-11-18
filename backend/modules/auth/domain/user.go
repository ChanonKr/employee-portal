// Package domain holds the account rules: users, sign-in codes, roles and the
// password policy. No framework imports.
package domain

import (
	"regexp"
	"strings"
	"unicode"

	profiledomain "employee-portal/modules/profile/domain"
	"employee-portal/modules/shared"
)

type User struct {
	ID                 int64
	EmployeeCode       string
	PasswordHash       string
	Role               shared.Role
	IsActive           bool
	MustChangePassword bool
	EmployeeID         *int64 // nil when the account has no employee profile
}

var employeeCodePattern = regexp.MustCompile(`^[A-Z0-9-]{3,20}$`)

// NormalizeEmployeeCode upper-cases and validates a sign-in code (e.g. EMP-00042).
func NormalizeEmployeeCode(raw string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if !employeeCodePattern.MatchString(code) {
		return "", shared.NewValidationError("employeeCode", "must be 3–20 letters, digits or dashes")
	}
	return code, nil
}

func ParseRole(s string) (shared.Role, error) {
	switch r := shared.Role(s); r {
	case shared.RoleEmployee, shared.RoleAdmin:
		return r, nil
	case "":
		return shared.RoleEmployee, nil
	}
	return "", shared.NewValidationError("role", "must be employee or admin")
}

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72 // bcrypt ignores anything longer
)

// ValidateNewPassword is the password policy: 8+ characters, at most 72 bytes,
// at least one letter and one digit.
func ValidateNewPassword(field, pw string) error {
	if len([]rune(pw)) < minPasswordLength {
		return shared.NewValidationError(field, "must be at least 8 characters")
	}
	if len(pw) > maxPasswordBytes {
		return shared.NewValidationError(field, "must be at most 72 bytes")
	}
	var letter, digit bool
	for _, r := range pw {
		letter = letter || unicode.IsLetter(r)
		digit = digit || unicode.IsDigit(r)
	}
	if !letter || !digit {
		return shared.NewValidationError(field, "must contain a letter and a digit")
	}
	return nil
}

// NewAccount is what an admin enters to create an employee account.
type NewAccount struct {
	EmployeeCode string
	FirstName    string
	LastName     string
	Role         shared.Role
}

// NormalizeNewAccount validates every field and returns the cleaned values.
func NormalizeNewAccount(code, first, last, role string) (NewAccount, error) {
	var a NewAccount
	var err error
	if a.EmployeeCode, err = NormalizeEmployeeCode(code); err != nil {
		return a, err
	}
	if a.FirstName, err = profiledomain.NormalizeName("firstName", first); err != nil {
		return a, err
	}
	if a.LastName, err = profiledomain.NormalizeName("lastName", last); err != nil {
		return a, err
	}
	if a.Role, err = ParseRole(role); err != nil {
		return a, err
	}
	return a, nil
}
