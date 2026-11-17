package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"employee-portal/pkg/httputil"
)

// Recover turns a panic into a 500 envelope and logs the stack trace.
func Recover(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorContext(c.Request.Context(), "panic recovered",
					"panic", fmt.Sprint(r),
					"requestId", c.GetString(httputil.RequestIDKey),
					"stack", string(debug.Stack()))
				httputil.Abort(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong. Please try again.")
			}
		}()
		c.Next()
	}
}
