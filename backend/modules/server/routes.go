package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/auth"
	"employee-portal/modules/employee"
	"employee-portal/modules/entity"
	"employee-portal/modules/file"
	"employee-portal/modules/history"
	"employee-portal/modules/profile"
	"employee-portal/modules/setting"
	"employee-portal/modules/skill"
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
		Admin:    api.Group("/admin", signedIn, ready, middleware.RequireRole(entity.RoleAdmin)), // use cases check again
		SignedIn: signedIn,
		Ready:    ready,
	}

	auth.NewHandler(
		auth.NewUsecase(auth.NewRepository(s.db), s.tokens),
		auth.CookieOptions{Secure: s.cfg.IsProduction()},
	).RegisterRoutes(r, s.cfg.LoginRateLimit)

	profile.NewHandler(profile.NewUsecase(profile.Deps{
		Repo:   profile.NewRepository(s.db),
		Photos: imaging.NewProcessor(s.cfg.UploadMaxBytes),
		Files:  s.files,
		Log:    s.log,
	})).RegisterRoutes(r, s.cfg.UploadMaxBytes)

	history.NewHandler(history.NewUsecase(history.NewRepository(s.db), s.cfg.HistoryDays)).RegisterRoutes(r)
	skill.NewHandler(skill.NewUsecase(skill.NewRepository(s.db))).RegisterRoutes(r)
	setting.NewHandler(setting.NewUsecase(setting.NewRepository(s.db), s.cfg.HistoryDays)).RegisterRoutes(r)
	employee.NewHandler(employee.NewUsecase(employee.NewRepository(s.db))).RegisterRoutes(r)
	file.NewHandler(s.files).RegisterRoutes(r)
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
