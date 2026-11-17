package employee

import "time"

// Read-only views for the admin directory and dashboard
// (see profile/model.go for the rules). Queries name their tables explicitly.

type summaryRow struct {
	ID           int64
	EmployeeCode string
	FirstName    string
	LastName     string
	PhotoPath    *string
	Role         string
	IsActive     bool
}

type detailRow struct {
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

type skillRow struct {
	ID       int64
	Name     string
	IsActive bool
}

type topSkillRow struct {
	ID     int64
	Name   string
	UsedBy int64
}

type recentChangeRow struct {
	ID            int64
	EmployeeID    int64
	EmployeeCode  string
	FirstName     string
	LastName      string
	Field         string
	ChangedAt     time.Time
	ChangedByRole string
}
