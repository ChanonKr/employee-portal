package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/entity"
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
	Role               entity.Role `json:"role"`
	MustChangePassword bool        `json:"mustChangePassword"`
}

// Login godoc: POST /auth/login → 204 + Set-Cookie
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, entity.ErrUnauthorized)
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
