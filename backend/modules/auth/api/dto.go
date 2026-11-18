package api

import (
	"employee-portal/modules/shared"
)

// Request and response bodies (JSON).

type loginRequest struct {
	EmployeeCode string `json:"employeeCode" binding:"required,max=20"`
	Password     string `json:"password" binding:"required,max=72"`
}

type meResponse struct {
	UserID             int64       `json:"userId"`
	EmployeeID         *int64      `json:"employeeId"`
	EmployeeCode       string      `json:"employeeCode"`
	Role               shared.Role `json:"role"`
	MustChangePassword bool        `json:"mustChangePassword"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type createAccountRequest struct {
	EmployeeCode string `json:"employeeCode"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Role         string `json:"role"`
}

type createdAccountResponse struct {
	ID                int64  `json:"id"` // employee id, as in /admin/employees/:id
	EmployeeCode      string `json:"employeeCode"`
	TemporaryPassword string `json:"temporaryPassword"`
}
