// Package model holds the GORM structs for the tables profile reads and writes.
//
// The rules for every feature's model package: the structs mirror the SQL
// migrations (GORM never migrates the schema) and stay inside the feature. Other
// features declare their own structs for what they need, so model and repository
// packages never import another feature's. Fields with a database default carry
// a `default` tag so GORM leaves them out of INSERTs instead of writing the Go
// zero value. Only repository uses these; it converts them with ToDomain.
package model

import (
	"time"

	"gorm.io/gorm"

	"employee-portal/modules/profile/domain"
	skilldomain "employee-portal/modules/skill/domain"
)

type Employee struct {
	ID           int64 `gorm:"primaryKey"`
	UserID       int64
	FirstName    string
	LastName     string
	PhotoPath    *string
	MobilePhone  *string
	ContactEmail *string
	Address      *string
	Version      int32     `gorm:"default:1"`
	UpdatedAt    time.Time `gorm:"default:now()"`
}

func (Employee) TableName() string { return "employees" }

func (e Employee) ToDomain() domain.Employee {
	return domain.Employee{
		ID: e.ID, UserID: e.UserID, FirstName: e.FirstName, LastName: e.LastName,
		PhotoPath: e.PhotoPath, MobilePhone: e.MobilePhone, ContactEmail: e.ContactEmail,
		Address: e.Address, Version: e.Version,
	}
}

type Skill struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	IsActive  bool
	SortOrder int32
}

func (Skill) TableName() string { return "skills" }

func ToSkills(rows []Skill) []skilldomain.Skill {
	out := make([]skilldomain.Skill, len(rows))
	for i, r := range rows {
		out[i] = skilldomain.Skill{ID: r.ID, Name: r.Name, IsActive: r.IsActive}
	}
	return out
}

type EmployeeSkill struct {
	EmployeeID int64 `gorm:"primaryKey"`
	SkillID    int64 `gorm:"primaryKey"`
}

func (EmployeeSkill) TableName() string { return "employee_skills" }

// ChangeLog is append-only: the database role may only INSERT and SELECT it.
type ChangeLog struct {
	ID         int64 `gorm:"primaryKey"`
	EmployeeID int64
	Field      string
	OldValue   []byte `gorm:"type:jsonb"` // nil = SQL NULL
	NewValue   []byte `gorm:"type:jsonb"`
	ChangedBy  int64
	ChangedAt  time.Time `gorm:"default:now()"`
}

func (ChangeLog) TableName() string { return "profile_change_logs" }

// SkillsOf limits a skills query to one employee's selection, in display order.
func SkillsOf(employeeID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN employee_skills es ON es.skill_id = skills.id").
			Where("es.employee_id = ?", employeeID).
			Scopes(SkillOrder)
	}
}

func SkillOrder(db *gorm.DB) *gorm.DB {
	return db.Order("skills.sort_order").Order("skills.name")
}
