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

	// signedIn: valid session. ready: also not using a temporary password.
	// /auth/* only needs signedIn, so a user can still see who they are and change it.
	signedIn := middleware.Auth(s.tokens)
	ready := middleware.RequirePasswordChanged()

	// ── auth
	authH := auth.NewHandler(
		auth.NewUsecase(auth.NewRepository(s.db), s.tokens),
		auth.CookieOptions{Secure: s.cfg.IsProduction()},
	)
	loginPerMinute := s.cfg.LoginRateLimit
	if loginPerMinute < 1 {
		loginPerMinute = 5 // zero value (e.g. a Config built in tests)
	}
	loginLimiter := middleware.NewKeyedLimiter(loginPerMinute, time.Minute)

	a := api.Group("/auth")
	a.POST("/login", middleware.RateLimit(loginLimiter, middleware.ByIP), authH.Login)
	a.POST("/logout", authH.Logout)
	a.GET("/me", signedIn, authH.Me)
	a.POST("/change-password", signedIn, middleware.RateLimit(loginLimiter, middleware.ByUser), authH.ChangePassword)

	// ── profile (employee id always comes from the token; the URL has no id)
	profileH := profile.NewHandler(profile.NewUsecase(profile.Deps{
		Repo:   profile.NewRepository(s.db),
		Photos: imaging.NewProcessor(s.cfg.UploadMaxBytes),
		Files:  s.files,
		Log:    s.log,
	}))

	p := api.Group("/profile", signedIn, ready, middleware.RequireProfile())
	p.GET("", profileH.GetProfile)
	p.PUT("/contact/:field", profileH.UpdateContact)
	p.PUT("/skills", profileH.SetSkills)
	// No route changes the name: F2 is enforced by the server, not only the UI.

	uploadLimiter := middleware.NewKeyedLimiter(10, time.Minute)
	upload := s.engine.Group("/api/v1/profile",
		middleware.BodyLimit(s.cfg.UploadMaxBytes+multipartOverhead),
		signedIn, ready, middleware.RequireProfile(),
		middleware.RateLimit(uploadLimiter, middleware.ByUser))
	upload.PUT("/photo", profileH.UpdatePhoto)

	// ── files
	fileH := file.NewHandler(s.files)
	api.GET("/files/photos/:name", signedIn, ready, fileH.Photo)

	// ── history (read-only; writes happen inside the profile transactions)
	historyH := history.NewHandler(history.NewUsecase(history.NewRepository(s.db), s.cfg.HistoryDays))
	p.GET("/history", historyH.ListOwn)

	// ── skill catalog (picker)
	skillH := skill.NewHandler(skill.NewUsecase(skill.NewRepository(s.db)))
	api.GET("/skills", signedIn, ready, skillH.List)

	// ── admin (router checks the role; every admin use case checks it again)
	settingH := setting.NewHandler(setting.NewUsecase(setting.NewRepository(s.db), s.cfg.HistoryDays))

	admin := api.Group("/admin", signedIn, ready, middleware.RequireRole(entity.RoleAdmin))
	admin.GET("/skills", skillH.AdminList)
	admin.POST("/skills", skillH.Create)
	admin.PATCH("/skills/:id", skillH.Update)
	admin.PUT("/skills/order", skillH.Reorder)
	admin.GET("/settings", settingH.Get)
	admin.PUT("/settings/history-days", settingH.SetHistoryDays)

	employeeH := employee.NewHandler(employee.NewUsecase(employee.NewRepository(s.db)))
	admin.GET("/stats", employeeH.Stats)
	admin.GET("/employees", employeeH.List)
	admin.GET("/employees/:id", employeeH.Get)
	admin.GET("/employees/:id/history", historyH.AdminList)
	admin.PUT("/employees/:id/name", profileH.AdminUpdateName)
	admin.POST("/employees", authH.AdminCreateAccount)
	admin.POST("/employees/:id/reset-password", authH.AdminResetPassword)
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
