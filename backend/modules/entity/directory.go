package entity

import (
	"strings"
	"time"
	"unicode/utf8"
)

// ── name (read-only for employees; only admins change it)

const maxNameLength = 100

func normalizeName(field, raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	n := utf8.RuneCountInString(name)
	if n == 0 {
		return "", NewValidationError(field, "is required")
	}
	if n > maxNameLength {
		return "", NewValidationError(field, "must be at most 100 characters")
	}
	return name, nil
}

// ChangeName sets the legal name. Each part that changed is recorded separately
// (first_name, last_name); the result is empty when nothing changed.
func (e *Employee) ChangeName(first, last string, expectedVersion int32) ([]Change, error) {
	first, err := normalizeName("firstName", first)
	if err != nil {
		return nil, err
	}
	last, err = normalizeName("lastName", last)
	if err != nil {
		return nil, err
	}
	if err := e.CheckVersion(expectedVersion); err != nil {
		return nil, err
	}

	changes := []Change{}
	if first != e.FirstName {
		changes = append(changes, Change{Field: FieldFirstName, Old: e.FirstName, New: first})
		e.FirstName = first
	}
	if last != e.LastName {
		changes = append(changes, Change{Field: FieldLastName, Old: e.LastName, New: last})
		e.LastName = last
	}
	return changes, nil
}

// ── admin directory

// EmployeeDetail is the admin view of one employee.
type EmployeeDetail struct {
	Profile
	Role     Role
	IsActive bool
}

type EmployeeSummary struct {
	ID           int64
	EmployeeCode string
	FirstName    string
	LastName     string
	PhotoPath    *string
	Role         Role
	IsActive     bool
}

type EmployeeQuery struct {
	Search string // matches employee code or "first last", case-insensitive
	PageRequest
}

// LikePattern turns user search text into a safe ILIKE "contains" pattern:
// %, _ and \ typed by the user match literally. Returns nil for blank input.
func (q EmployeeQuery) LikePattern() *string {
	s := strings.TrimSpace(q.Search)
	if s == "" {
		return nil
	}
	s = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
	p := "%" + s + "%"
	return &p
}

// ── dashboard

type SkillCount struct {
	Skill
	UsedBy int64
}

type RecentChange struct {
	ID            int64
	EmployeeID    int64
	EmployeeCode  string
	FirstName     string
	LastName      string
	Field         FieldName
	ChangedAt     time.Time
	ChangedByRole Role
}

type DashboardStats struct {
	Employees        int64
	CompleteProfiles int64 // photo + all contact fields + at least one skill
	TopSkills        []SkillCount
	RecentChanges    []RecentChange
}
