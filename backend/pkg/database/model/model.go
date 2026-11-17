// Package model holds the GORM persistence structs. They mirror the tables created by
// the SQL migrations (GORM never migrates the schema) and stay out of the domain:
// repositories convert them to and from entity types.
package model

import "time"

// Columns with a database default carry a `default` tag so GORM leaves them out of
// INSERTs when the Go value is zero, instead of writing false / 0 over the default.

type User struct {
	ID                 int64 `gorm:"primaryKey"`
	EmployeeCode       string
	PasswordHash       string
	Role               string `gorm:"default:employee"`
	IsActive           bool   `gorm:"default:true"`
	MustChangePassword bool
	CreatedAt          time.Time `gorm:"->"` // set by the database
}

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

type Skill struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	IsActive  bool `gorm:"default:true"`
	SortOrder int32
}

type EmployeeSkill struct {
	EmployeeID int64 `gorm:"primaryKey"`
	SkillID    int64 `gorm:"primaryKey"`
}

// ProfileChangeLog is append-only: the database role may only INSERT and SELECT it.
type ProfileChangeLog struct {
	ID         int64 `gorm:"primaryKey"`
	EmployeeID int64
	Field      string
	OldValue   []byte `gorm:"type:jsonb"` // nil = SQL NULL
	NewValue   []byte `gorm:"type:jsonb"`
	ChangedBy  int64
	ChangedAt  time.Time `gorm:"default:now()"`
}

type AppSetting struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedBy *int64
	UpdatedAt time.Time
}
