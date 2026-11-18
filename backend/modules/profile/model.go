package profile

import (
	"time"

	"gorm.io/gorm"

	profiledomain "employee-portal/modules/profile/domain"
	skilldomain "employee-portal/modules/skill/domain"
)

// GORM structs for the tables this module reads and writes. They mirror the SQL
// migrations (GORM never migrates the schema) and stay inside the module: other
// modules declare their own structs for what they need, so modules never import
// each other. Fields with a database default carry a `default` tag so GORM leaves
// them out of INSERTs instead of writing the Go zero value.

type employeeModel struct {
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

func (employeeModel) TableName() string { return "employees" }

func (e employeeModel) toEntity() profiledomain.Employee {
	return profiledomain.Employee{
		ID: e.ID, UserID: e.UserID, FirstName: e.FirstName, LastName: e.LastName,
		PhotoPath: e.PhotoPath, MobilePhone: e.MobilePhone, ContactEmail: e.ContactEmail,
		Address: e.Address, Version: e.Version,
	}
}

type skillModel struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	IsActive  bool
	SortOrder int32
}

func (skillModel) TableName() string { return "skills" }

func toSkills(rows []skillModel) []skilldomain.Skill {
	out := make([]skilldomain.Skill, len(rows))
	for i, r := range rows {
		out[i] = skilldomain.Skill{ID: r.ID, Name: r.Name, IsActive: r.IsActive}
	}
	return out
}

type employeeSkillModel struct {
	EmployeeID int64 `gorm:"primaryKey"`
	SkillID    int64 `gorm:"primaryKey"`
}

func (employeeSkillModel) TableName() string { return "employee_skills" }

// changeLogModel is append-only: the database role may only INSERT and SELECT it.
type changeLogModel struct {
	ID         int64 `gorm:"primaryKey"`
	EmployeeID int64
	Field      string
	OldValue   []byte `gorm:"type:jsonb"` // nil = SQL NULL
	NewValue   []byte `gorm:"type:jsonb"`
	ChangedBy  int64
	ChangedAt  time.Time `gorm:"default:now()"`
}

func (changeLogModel) TableName() string { return "profile_change_logs" }

// skillsOf limits a skills query to one employee's selection, in display order.
func skillsOf(employeeID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN employee_skills es ON es.skill_id = skills.id").
			Where("es.employee_id = ?", employeeID).
			Scopes(skillOrder)
	}
}

func skillOrder(db *gorm.DB) *gorm.DB {
	return db.Order("skills.sort_order").Order("skills.name")
}
