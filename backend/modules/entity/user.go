package entity

type Role string

const (
	RoleEmployee Role = "employee"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID                 int64
	EmployeeCode       string
	PasswordHash       string
	Role               Role
	IsActive           bool
	MustChangePassword bool
	EmployeeID         *int64 // nil when the account has no employee profile
}
