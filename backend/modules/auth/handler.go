package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

type CookieOptions struct {
	Secure bool // true in production (HTTPS only)
}

type Handler struct {
	uc     Usecase
	cookie CookieOptions
}

func NewHandler(uc Usecase, cookie CookieOptions) *Handler {
	return &Handler{uc: uc, cookie: cookie}
}

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

// Login godoc: POST /auth/login → 204 + Set-Cookie
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.ErrUnauthorized)
		return
	}
	s, err := h.uc.Login(c.Request.Context(), req.EmployeeCode, req.Password)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	h.setSession(c, s.Token, int(time.Until(s.ExpiresAt).Seconds()))
	c.Status(http.StatusNoContent)
}

// Logout clears the cookie. It works without a valid session.
func (h *Handler) Logout(c *gin.Context) {
	h.setSession(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (h *Handler) Me(c *gin.Context) {
	u, err := h.uc.Me(c.Request.Context(), middleware.ActorFrom(c))
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, meResponse{
		UserID: u.ID, EmployeeID: u.EmployeeID, EmployeeCode: u.EmployeeCode,
		Role: u.Role, MustChangePassword: u.MustChangePassword,
	})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangePassword: POST /auth/change-password → 204 + a new session cookie
// (the old token still carries the must-change flag).
func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("body", "must be JSON with currentPassword and newPassword"))
		return
	}
	s, err := h.uc.ChangePassword(c.Request.Context(), middleware.ActorFrom(c), req.CurrentPassword, req.NewPassword)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	h.setSession(c, s.Token, int(time.Until(s.ExpiresAt).Seconds()))
	c.Status(http.StatusNoContent)
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

// AdminCreateAccount: POST /admin/employees → 201 with a one-time temporary password.
func (h *Handler) AdminCreateAccount(c *gin.Context) {
	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("body", "must be JSON with employeeCode, firstName, lastName"))
		return
	}
	acc, err := h.uc.CreateAccount(c.Request.Context(), middleware.ActorFrom(c), CreateAccountInput{
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
func (h *Handler) AdminResetPassword(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	temp, err := h.uc.ResetPassword(c.Request.Context(), middleware.ActorFrom(c), id)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"temporaryPassword": temp})
}

func (h *Handler) setSession(c *gin.Context, value string, maxAge int) {
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
