package entity

// Actor is the authenticated user performing a request. It comes from the session
// token, never from the URL or request body.
type Actor struct {
	UserID     int64
	EmployeeID int64 // 0 when the account has no employee profile
	Role       Role
	// MustChangePassword is set while the account uses a temporary password.
	MustChangePassword bool
}

func (a Actor) IsAdmin() bool { return a.Role == RoleAdmin }

func (a Actor) HasProfile() bool { return a.EmployeeID != 0 }
