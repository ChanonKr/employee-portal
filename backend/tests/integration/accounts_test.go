//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type created struct {
	ID                int64  `json:"id"`
	EmployeeCode      string `json:"employeeCode"`
	TemporaryPassword string `json:"temporaryPassword"`
}

// loginWith signs in with an explicit password (login() uses the demo password).
func loginWith(t *testing.T, base, code, password string) (*apiClient, int) {
	t.Helper()
	c := &apiClient{t: t, base: base}
	res := c.do("POST", "/api/v1/auth/login", map[string]string{"employeeCode": code, "password": password}, nil)
	c.cookies = res.Cookies()
	return c, res.StatusCode
}

func TestAccounts_CreateFirstLoginAndChangePassword(t *testing.T) {
	base := newTestServer(t)
	admin := login(t, base, newAdmin(t))
	code := fmt.Sprintf("NEW-%d", time.Now().UnixNano()%1_000_000)

	res := admin.do("POST", "/api/v1/admin/employees",
		map[string]string{"employeeCode": code, "firstName": "New", "lastName": "Hire"}, nil)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	acc := decodeJSON[created](t, res)
	assert.Equal(t, code, acc.EmployeeCode)
	require.Len(t, acc.TemporaryPassword, 12)

	// Same code again → 409
	res = admin.do("POST", "/api/v1/admin/employees",
		map[string]string{"employeeCode": code, "firstName": "X", "lastName": "Y"}, nil)
	assert.Equal(t, http.StatusConflict, res.StatusCode)

	// First sign-in with the temporary password
	user, status := loginWith(t, base, code, acc.TemporaryPassword)
	require.Equal(t, http.StatusNoContent, status)

	me := decodeJSON[struct{ MustChangePassword bool }](t, user.do("GET", "/api/v1/auth/me", nil, nil))
	assert.True(t, me.MustChangePassword)

	res = user.do("GET", "/api/v1/profile", nil, nil)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
	assert.Equal(t, "PASSWORD_CHANGE_REQUIRED", decodeJSON[struct{ Error struct{ Code string } }](t, res).Error.Code)
	assert.Equal(t, http.StatusForbidden, user.do("GET", "/api/v1/skills", nil, nil).StatusCode)

	// Wrong current password → 400 (not 401, which would log the user out)
	res = user.do("POST", "/api/v1/auth/change-password",
		map[string]string{"currentPassword": "wrong", "newPassword": "brandnew123"}, nil)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	res = user.do("POST", "/api/v1/auth/change-password",
		map[string]string{"currentPassword": acc.TemporaryPassword, "newPassword": "brandnew123"}, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	user.cookies = res.Cookies() // the new session no longer carries the flag

	res = user.do("GET", "/api/v1/profile", nil, nil)
	require.Equal(t, http.StatusOK, res.StatusCode)
	p := decodeJSON[struct {
		ID        int64  `json:"id"`
		FirstName string `json:"firstName"`
	}](t, res)
	assert.Equal(t, acc.ID, p.ID)
	assert.Equal(t, "New", p.FirstName)

	_, status = loginWith(t, base, code, acc.TemporaryPassword)
	assert.Equal(t, http.StatusUnauthorized, status, "temporary password is gone")
	_, status = loginWith(t, base, code, "brandnew123")
	assert.Equal(t, http.StatusNoContent, status)
}

func TestAccounts_ResetPassword(t *testing.T) {
	base := newTestServer(t)
	admin := login(t, base, newAdmin(t))
	_, empID, code := newEmployee(t, "0811111111")

	res := admin.do("POST", fmt.Sprintf("/api/v1/admin/employees/%d/reset-password", empID), nil, nil)
	require.Equal(t, http.StatusOK, res.StatusCode)
	temp := decodeJSON[struct{ TemporaryPassword string }](t, res).TemporaryPassword
	require.Len(t, temp, 12)

	_, status := loginWith(t, base, code, "demo1234")
	assert.Equal(t, http.StatusUnauthorized, status, "old password stops working")

	user, status := loginWith(t, base, code, temp)
	require.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, http.StatusForbidden, user.do("GET", "/api/v1/profile", nil, nil).StatusCode)

	assert.Equal(t, http.StatusNotFound,
		admin.do("POST", "/api/v1/admin/employees/999999999/reset-password", nil, nil).StatusCode)
}

func TestAccounts_EmployeesCannotManageAccounts(t *testing.T) {
	base := newTestServer(t)
	_, empID, code := newEmployee(t, "0811111111")
	emp := login(t, base, code)

	res := emp.do("POST", "/api/v1/admin/employees",
		map[string]string{"employeeCode": "HACK-1", "firstName": "A", "lastName": "B", "role": "admin"}, nil)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
	res = emp.do("POST", fmt.Sprintf("/api/v1/admin/employees/%d/reset-password", empID), nil, nil)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
}
