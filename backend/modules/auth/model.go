package auth

import "time"

// GORM structs for what this module touches (see profile/model.go for the rules).

type userModel struct {
	ID                 int64 `gorm:"primaryKey"`
	EmployeeCode       string
	PasswordHash       string
	Role               string `gorm:"default:employee"`
	IsActive           bool   `gorm:"default:true"`
	MustChangePassword bool
	CreatedAt          time.Time `gorm:"->"` // set by the database
}

func (userModel) TableName() string { return "users" }

// employeeModel is only what account creation writes: the empty profile row.
// Version and the contact fields keep their database defaults.
type employeeModel struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64
	FirstName string
	LastName  string
}

func (employeeModel) TableName() string { return "employees" }
