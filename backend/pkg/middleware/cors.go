package middleware

import (
	"net/url"

	"github.com/gin-gonic/gin"
)

// SameOriginOrCORS skips the CORS check for same-origin requests. Browsers send
// an Origin header on same-origin POST/PUT too; without this, the SPA served by
// nginx (or the Vite proxy) on any host not listed in CORS_ORIGINS would get 403.
// Cross-origin requests still go through cors (the allowlist).
func SameOriginOrCORS(cors gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		if u, err := url.Parse(origin); err == nil && u.Host != "" && u.Host == c.Request.Host {
			c.Next()
			return
		}
		cors(c)
	}
}
