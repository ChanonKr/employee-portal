// Package model holds the read-only row shapes for the admin directory and
// dashboard (see modules/profile/model for the rules). Queries name their tables
// explicitly; only repository uses these.
package model

import "time"

type SummaryRow struct {
	ID           int64
	EmployeeCode string
	FirstName    string
	LastName     string
	PhotoPath    *string
	Role         string
	IsActive     bool
}

type DetailRow struct {
	ID           int64
	UserID       int64
	FirstName    string
	LastName     string
	PhotoPath    *string
	MobilePhone  *string
	ContactEmail *string
	Address      *string
	Version      int32
	EmployeeCode string
	Role         string
	IsActive     bool
}

type SkillRow struct {
	ID       int64
	Name     string
	IsActive bool
}

type TopSkillRow struct {
	ID     int64
	Name   string
	UsedBy int64
}

type RecentChangeRow struct {
	ID            int64
	EmployeeID    int64
	EmployeeCode  string
	FirstName     string
	LastName      string
	Field         string
	ChangedAt     time.Time
	ChangedByRole string
}
