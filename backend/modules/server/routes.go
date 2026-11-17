package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/auth"
	"employee-portal/modules/file"
	"employee-portal/modules/history"
	"employee-portal/modules/profile"
	"employee-portal/modules/skill"
	"employee-portal/pkg/imaging"
	"employee-portal/pkg/middleware"
)

// multipartOverhead leaves room for multipart headers around the file itself.
const multipartOverhead = 64 << 10

// routes wires each module: repository → usecase → handler → route group.
// This is the only place that knows about every layer.
func (s *Server) routes() {
	// JSON endpoints get a small body limit. The photo upload is registered on its
	// own group below with a limit sized for the file.
	api := s.engine.Group("/api/v1", middleware.BodyLimit(jsonBodyLimit))
	api.GET("/healthz", s.health)

	requireAuth := middleware.Auth(s.tokens)

	// ── auth
	authH := auth.NewHandler(
		auth.NewUsecase(auth.NewRepository(s.pool), s.tokens),
		auth.CookieOptions{Secure: s.cfg.IsProduction()},
	)
	loginLimiter := middleware.NewKeyedLimiter(5, time.Minute)

	a := api.Group("/auth")
	a.POST("/login", middleware.RateLimit(loginLimiter, middleware.ByIP), authH.Login)
	a.POST("/logout", authH.Logout)
	a.GET("/me", requireAuth, authH.Me)

	// ── profile (employee id always comes from the token; the URL has no id)
	profileH := profile.NewHandler(profile.NewUsecase(profile.Deps{
		Repo:   profile.NewRepository(s.pool),
		Photos: imaging.NewProcessor(s.cfg.UploadMaxBytes),
		Files:  s.files,
		Log:    s.log,
	}))

	p := api.Group("/profile", requireAuth, middleware.RequireProfile())
	p.GET("", profileH.GetProfile)
	p.PUT("/contact/:field", profileH.UpdateContact)
	p.PUT("/skills", profileH.SetSkills)
	// No route changes the name: F2 is enforced by the server, not only the UI.

	uploadLimiter := middleware.NewKeyedLimiter(10, time.Minute)
	upload := s.engine.Group("/api/v1/profile",
		middleware.BodyLimit(s.cfg.UploadMaxBytes+multipartOverhead),
		requireAuth, middleware.RequireProfile(),
		middleware.RateLimit(uploadLimiter, middleware.ByUser))
	upload.PUT("/photo", profileH.UpdatePhoto)

	// ── files
	fileH := file.NewHandler(s.files)
	api.GET("/files/photos/:name", requireAuth, fileH.Photo)

	// ── history (read-only; writes happen inside the profile transactions)
	historyH := history.NewHandler(history.NewUsecase(history.NewRepository(s.pool), s.cfg.HistoryDays))
	p.GET("/history", historyH.ListOwn)

	// ── skill catalog (picker)
	skillH := skill.NewHandler(skill.NewUsecase(skill.NewRepository(s.pool)))
	api.GET("/skills", requireAuth, skillH.List)
}

func (s *Server) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		s.log.WarnContext(ctx, "health check: database unreachable", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "db": "down"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "ok"})
}
