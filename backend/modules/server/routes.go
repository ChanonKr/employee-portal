package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	authapi "employee-portal/modules/auth/api"
	authrepo "employee-portal/modules/auth/repo"
	authservice "employee-portal/modules/auth/service"
	employeeapi "employee-portal/modules/employee/api"
	employeerepo "employee-portal/modules/employee/repo"
	employeeservice "employee-portal/modules/employee/service"
	"employee-portal/modules/file"
	historyapi "employee-portal/modules/history/api"
	historyrepo "employee-portal/modules/history/repo"
	historyservice "employee-portal/modules/history/service"
	"employee-portal/modules/profile"
	settingapi "employee-portal/modules/setting/api"
	settingrepo "employee-portal/modules/setting/repo"
	settingservice "employee-portal/modules/setting/service"
	"employee-portal/modules/shared"
	skillapi "employee-portal/modules/skill/api"
	skillrepo "employee-portal/modules/skill/repo"
	skillservice "employee-portal/modules/skill/service"
	"employee-portal/pkg/database"
	"employee-portal/pkg/imaging"
	"employee-portal/pkg/middleware"
	"employee-portal/pkg/router"
)

// routes builds the shared groups and middleware, wires each module
// (repository → use case → handler) and lets the module register its own routes.
// This is the only place that knows about every module.
func (s *Server) routes() {
	// JSON endpoints get a small body limit. The photo upload uses Root and sets its own.
	api := s.engine.Group("/api/v1", middleware.BodyLimit(jsonBodyLimit))
	api.GET("/healthz", s.health)

	signedIn := middleware.Auth(s.tokens)
	ready := middleware.RequirePasswordChanged()
	r := router.Groups{
		Root:     &s.engine.RouterGroup,
		API:      api,
		Admin:    api.Group("/admin", signedIn, ready, middleware.RequireRole(shared.RoleAdmin)), // use cases check again
		SignedIn: signedIn,
		Ready:    ready,
	}

	auth := authapi.New(
		authservice.New(authrepo.New(s.db), s.tokens),
		authapi.CookieOptions{Secure: s.cfg.IsProduction()},
	)
	// Limits sign-in and password-change attempts (0 = default 5 per minute).
	loginPerMinute := s.cfg.LoginRateLimit
	if loginPerMinute < 1 {
		loginPerMinute = 5 // zero value (e.g. a Config built in tests)
	}
	loginLimiter := middleware.NewKeyedLimiter(loginPerMinute, time.Minute)
	// /auth/* only needs SignedIn, so a user on a temporary password can still
	// see who they are and change it.
	a := r.API.Group("/auth")
	a.POST("/login", middleware.RateLimit(loginLimiter, middleware.ByIP), auth.Login)
	a.POST("/logout", auth.Logout)
	a.GET("/me", r.SignedIn, auth.Me)
	a.POST("/change-password", r.SignedIn, middleware.RateLimit(loginLimiter, middleware.ByUser), auth.ChangePassword)
	r.Admin.POST("/employees", auth.AdminCreateAccount)
	r.Admin.POST("/employees/:id/reset-password", auth.AdminResetPassword)

	profile.NewHandler(profile.NewUsecase(profile.Deps{
		Repo:   profile.NewRepository(s.db),
		Photos: imaging.NewProcessor(s.cfg.UploadMaxBytes),
		Files:  s.files,
		Log:    s.log,
	})).RegisterRoutes(r, s.cfg.UploadMaxBytes)

	file.NewHandler(s.files).RegisterRoutes(r)

	// Admin directory and dashboard (read-only; edits go through profile).
	employees := employeeapi.New(employeeservice.New(employeerepo.New(s.db)))
	r.Admin.GET("/stats", employees.Stats)
	r.Admin.GET("/employees", employees.List)
	r.Admin.GET("/employees/:id", employees.Get)

	// History is read-only; writes happen inside the profile transactions.
	history := historyapi.New(historyservice.New(historyrepo.New(s.db), s.cfg.HistoryDays))
	r.API.GET("/profile/history", r.SignedIn, r.Ready, middleware.RequireProfile(), history.ListOwn)
	r.Admin.GET("/employees/:id/history", history.AdminList)

	skills := skillapi.New(skillservice.New(skillrepo.New(s.db)))
	r.API.GET("/skills", r.SignedIn, r.Ready, skills.List)
	r.Admin.GET("/skills", skills.AdminList)
	r.Admin.POST("/skills", skills.Create)
	r.Admin.PATCH("/skills/:id", skills.Update)
	r.Admin.PUT("/skills/order", skills.Reorder)

	settings := settingapi.New(settingservice.New(settingrepo.New(s.db), s.cfg.HistoryDays))
	r.Admin.GET("/settings", settings.Get)
	r.Admin.PUT("/settings/history-days", settings.SetHistoryDays)
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
