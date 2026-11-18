// Package model holds the GORM structs auth reads and writes (see
// modules/profile/model for the rules); only repository uses them.
package model

import "time"

type User struct {
	ID                 int64 `gorm:"primaryKey"`
	EmployeeCode       string
	PasswordHash       string
	Role               string `gorm:"default:employee"`
	IsActive           bool   `gorm:"default:true"`
	MustChangePassword bool
	CreatedAt          time.Time `gorm:"->"` // set by the database
}

func (User) TableName() string { return "users" }

// Employee is only what account creation writes: the empty profile row.
// Version and the contact fields keep their database defaults.
type Employee struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64
	FirstName string
	LastName  string
}

func (Employee) TableName() string { return "employees" }
