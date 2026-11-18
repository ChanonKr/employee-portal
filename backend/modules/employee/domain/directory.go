// Package domain holds the admin directory views: employee list, detail and
// dashboard stats. No framework imports.
package domain

import (
	"strings"
	"time"

	historydomain "employee-portal/modules/history/domain"
	profiledomain "employee-portal/modules/profile/domain"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

// EmployeeDetail is the admin view of one employee.
type EmployeeDetail struct {
	profiledomain.Profile
	Role     shared.Role
	IsActive bool
}

type EmployeeSummary struct {
	ID           int64
	EmployeeCode string
	FirstName    string
	LastName     string
	PhotoPath    *string
	Role         shared.Role
	IsActive     bool
}

type EmployeeQuery struct {
	Search string // matches employee code or "first last", case-insensitive
	shared.PageRequest
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
	skilldomain.Skill
	UsedBy int64
}

type RecentChange struct {
	ID            int64
	EmployeeID    int64
	EmployeeCode  string
	FirstName     string
	LastName      string
	Field         historydomain.FieldName
	ChangedAt     time.Time
	ChangedByRole shared.Role
}

type DashboardStats struct {
	Employees        int64
	CompleteProfiles int64 // photo + all contact fields + at least one skill
	TopSkills        []SkillCount
	RecentChanges    []RecentChange
}
