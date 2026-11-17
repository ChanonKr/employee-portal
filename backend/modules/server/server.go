package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"employee-portal/configs"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
	"employee-portal/pkg/security"
	"employee-portal/pkg/storage"
)

const (
	jsonBodyLimit  = 64 << 10 // 64 KB for JSON endpoints
	requestTimeout = 15 * time.Second
)

type Server struct {
	cfg    configs.Config
	log    *slog.Logger
	pool   *pgxpool.Pool
	tokens *security.TokenManager
	files  *storage.Local
	engine *gin.Engine
}

func New(cfg configs.Config, log *slog.Logger, pool *pgxpool.Pool) (*Server, error) {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	if err := engine.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("trusted proxies: %w", err)
	}

	engine.Use(
		middleware.RequestID(),
		middleware.Recover(log),
		middleware.Logger(log),
		middleware.SecureHeaders(),
		middleware.SameOriginOrCORS(cors.New(cors.Config{
			AllowOrigins:     cfg.CORSOrigins,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
			AllowHeaders:     []string{"Content-Type", "If-Match", "X-Request-ID"},
			ExposeHeaders:    []string{"ETag", "X-Request-ID", "Retry-After"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		})),
		middleware.Timeout(requestTimeout),
	)
	engine.NoRoute(func(c *gin.Context) {
		httputil.Abort(c, http.StatusNotFound, "NOT_FOUND", "Not found.")
	})
	engine.NoMethod(func(c *gin.Context) {
		httputil.Abort(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	})

	files, err := storage.NewLocal(cfg.UploadDir)
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:    cfg,
		log:    log,
		pool:   pool,
		tokens: security.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL),
		files:  files,
		engine: engine,
	}
	s.routes()
	return s, nil
}

// Handler exposes the router for tests (httptest).
func (s *Server) Handler() http.Handler { return s.engine }

// Run serves until ctx is cancelled, then drains in-flight requests.
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.Port),
		Handler:           s.engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("api listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	s.log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
