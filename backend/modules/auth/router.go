package auth

import (
	"time"

	"employee-portal/pkg/middleware"
	"employee-portal/pkg/router"
)

// RegisterRoutes mounts /auth/* and the admin account routes.
// loginPerMinute limits sign-in and password-change attempts (0 = default 5).
func (h *Handler) RegisterRoutes(r router.Groups, loginPerMinute int) {
	if loginPerMinute < 1 {
		loginPerMinute = 5 // zero value (e.g. a Config built in tests)
	}
	limiter := middleware.NewKeyedLimiter(loginPerMinute, time.Minute)

	// /auth/* only needs SignedIn, so a user on a temporary password can still
	// see who they are and change it.
	a := r.API.Group("/auth")
	a.POST("/login", middleware.RateLimit(limiter, middleware.ByIP), h.Login)
	a.POST("/logout", h.Logout)
	a.GET("/me", r.SignedIn, h.Me)
	a.POST("/change-password", r.SignedIn, middleware.RateLimit(limiter, middleware.ByUser), h.ChangePassword)

	r.Admin.POST("/employees", h.AdminCreateAccount)
	r.Admin.POST("/employees/:id/reset-password", h.AdminResetPassword)
}
