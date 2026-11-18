package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
	"employee-portal/pkg/database"
	"employee-portal/pkg/middleware"
)

const (
	defaultLoginPerMinute = 5
	uploadPerMinute       = 10
)

// routes is the single route table: every endpoint, its group and middleware.
func (s *Server) routes() {
	h := s.wire()

	// Shared middleware.
	signedIn := middleware.Auth(s.tokens)
	ready := middleware.RequirePasswordChanged() // rejects sessions still on a temporary password
	hasProfile := middleware.RequireProfile()

	loginPerMinute := s.cfg.LoginRateLimit
	if loginPerMinute < 1 {
		loginPerMinute = defaultLoginPerMinute // zero value (e.g. a Config built in tests)
	}
	loginLimiter := middleware.NewKeyedLimiter(loginPerMinute, time.Minute)
	uploadLimiter := middleware.NewKeyedLimiter(uploadPerMinute, time.Minute)

	// JSON endpoints get a small body limit. The photo upload sets its own below.
	api := s.engine.Group("/api/v1", middleware.BodyLimit(jsonBodyLimit))
	api.GET("/healthz", s.health)

	// /auth/* only needs signedIn, so a user on a temporary password can still
	// see who they are and change it.
	auth := api.Group("/auth")
	auth.POST("/login", middleware.RateLimit(loginLimiter, middleware.ByIP), h.auth.Login)
	auth.POST("/logout", h.auth.Logout)
	auth.GET("/me", signedIn, h.auth.Me)
	auth.POST("/change-password", signedIn, middleware.RateLimit(loginLimiter, middleware.ByUser), h.auth.ChangePassword)

	// /profile/* takes the employee id from the token; the URL has no id.
	// No route changes the name: F2 is enforced by the server, not only the UI.
	profile := api.Group("/profile", signedIn, ready, hasProfile)
	profile.GET("", h.profile.GetProfile)
	profile.PUT("/contact/:field", h.profile.UpdateContact)
	profile.PUT("/skills", h.profile.SetSkills)
	profile.GET("/history", h.history.ListOwn)

	// The upload has its own group: a body limit sized for the file, not for JSON.
	upload := s.engine.Group("/api/v1/profile",
		middleware.BodyLimit(s.cfg.UploadMaxBytes+multipartOverhead),
		signedIn, ready, hasProfile,
		middleware.RateLimit(uploadLimiter, middleware.ByUser))
	upload.PUT("/photo", h.profile.UpdatePhoto)

	signedInReady := api.Group("", signedIn, ready)
	signedInReady.GET("/skills", h.skill.List)
	signedInReady.GET("/files/photos/:name", h.file.Photo)

	// Admin: the router checks the role and every service checks it again.
	admin := api.Group("/admin", signedIn, ready, middleware.RequireRole(shared.RoleAdmin))
	admin.GET("/stats", h.employee.Stats)
	admin.GET("/employees", h.employee.List)
	admin.POST("/employees", h.auth.AdminCreateAccount)
	admin.GET("/employees/:id", h.employee.Get)
	admin.PUT("/employees/:id/name", h.profile.AdminUpdateName)
	admin.POST("/employees/:id/reset-password", h.auth.AdminResetPassword)
	admin.GET("/employees/:id/history", h.history.AdminList)

	admin.GET("/skills", h.skill.AdminList)
	admin.POST("/skills", h.skill.Create)
	admin.PATCH("/skills/:id", h.skill.Update)
	admin.PUT("/skills/order", h.skill.Reorder)

	admin.GET("/settings", h.setting.Get)
	admin.PUT("/settings/history-days", h.setting.SetHistoryDays)
}

func (s *Server) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := database.Ping(ctx, s.db); err != nil {
		s.log.WarnContext(ctx, "health check: database unreachable", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "db": "down"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "ok"})
}
