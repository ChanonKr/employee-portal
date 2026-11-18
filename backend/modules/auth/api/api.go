package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/auth/service"
	"employee-portal/modules/shared"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

// CookieOptions configures the session cookie.
type CookieOptions struct {
	Secure bool // true in production (HTTPS only)
}

// API is the HTTP adapter for auth. Routes are registered in modules/server.
type API struct {
	svc    Service
	cookie CookieOptions
}

func New(svc Service, cookie CookieOptions) *API {
	return &API{svc: svc, cookie: cookie}
}

// Login godoc: POST /auth/login → 204 + Set-Cookie
func (h *API) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.ErrUnauthorized)
		return
	}
	s, err := h.svc.Login(c.Request.Context(), req.EmployeeCode, req.Password)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	h.setSession(c, s.Token, int(time.Until(s.ExpiresAt).Seconds()))
	c.Status(http.StatusNoContent)
}

// Logout clears the cookie. It works without a valid session.
func (h *API) Logout(c *gin.Context) {
	h.setSession(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (h *API) Me(c *gin.Context) {
	u, err := h.svc.Me(c.Request.Context(), middleware.ActorFrom(c))
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, meResponse{
		UserID: u.ID, EmployeeID: u.EmployeeID, EmployeeCode: u.EmployeeCode,
		Role: u.Role, MustChangePassword: u.MustChangePassword,
	})
}

// ChangePassword: POST /auth/change-password → 204 + a new session cookie
// (the old token still carries the must-change flag).
func (h *API) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("body", "must be JSON with currentPassword and newPassword"))
		return
	}
	s, err := h.svc.ChangePassword(c.Request.Context(), middleware.ActorFrom(c), req.CurrentPassword, req.NewPassword)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	h.setSession(c, s.Token, int(time.Until(s.ExpiresAt).Seconds()))
	c.Status(http.StatusNoContent)
}

// AdminCreateAccount: POST /admin/employees → 201 with a one-time temporary password.
func (h *API) AdminCreateAccount(c *gin.Context) {
	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("body", "must be JSON with employeeCode, firstName, lastName"))
		return
	}
	acc, err := h.svc.CreateAccount(c.Request.Context(), middleware.ActorFrom(c), service.CreateAccountInput{
		EmployeeCode: req.EmployeeCode, FirstName: req.FirstName, LastName: req.LastName, Role: req.Role,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, createdAccountResponse{
		ID: acc.EmployeeID, EmployeeCode: acc.EmployeeCode, TemporaryPassword: acc.TemporaryPassword,
	})
}

// AdminResetPassword: POST /admin/employees/:id/reset-password → {temporaryPassword}
func (h *API) AdminResetPassword(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	temp, err := h.svc.ResetPassword(c.Request.Context(), middleware.ActorFrom(c), id)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"temporaryPassword": temp})
}

func (h *API) setSession(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.SessionCookie,
		Value:    value,
		Path:     "/api",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	})
}
