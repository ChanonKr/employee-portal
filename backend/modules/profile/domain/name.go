package domain

import (
	"strings"
	"unicode/utf8"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
)

// The legal name is read-only for employees; only admins change it.

const maxNameLength = 100

// NormalizeName collapses whitespace and checks the length of one name part.
func NormalizeName(field, raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	n := utf8.RuneCountInString(name)
	if n == 0 {
		return "", shared.NewValidationError(field, "is required")
	}
	if n > maxNameLength {
		return "", shared.NewValidationError(field, "must be at most 100 characters")
	}
	return name, nil
}

// ChangeName sets the legal name. Each part that changed is recorded separately
// (first_name, last_name); the result is empty when nothing changed.
func (e *Employee) ChangeName(first, last string, expectedVersion int32) ([]historydomain.Change, error) {
	first, err := NormalizeName("firstName", first)
	if err != nil {
		return nil, err
	}
	last, err = NormalizeName("lastName", last)
	if err != nil {
		return nil, err
	}
	if err := e.CheckVersion(expectedVersion); err != nil {
		return nil, err
	}

	changes := []historydomain.Change{}
	if first != e.FirstName {
		changes = append(changes, historydomain.Change{Field: historydomain.FieldFirstName, Old: e.FirstName, New: first})
		e.FirstName = first
	}
	if last != e.LastName {
		changes = append(changes, historydomain.Change{Field: historydomain.FieldLastName, Old: e.LastName, New: last})
		e.LastName = last
	}
	return changes, nil
}
