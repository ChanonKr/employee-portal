package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/pkg/httputil"
)

// Logger writes one structured line per request.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		attrs := []any{
			"method", c.Request.Method,
			"route", c.FullPath(),
			"status", status,
			"durationMs", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"requestId", c.GetString(httputil.RequestIDKey),
		}
		if a, ok := c.Get(actorKey); ok {
			attrs = append(attrs, "userId", a.(actorValue).UserID)
		}
		log.Log(c.Request.Context(), level, "http request", attrs...)
	}
}
