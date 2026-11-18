package middleware

import (
	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
	"employee-portal/pkg/httputil"
)

// RequireRole must run after Auth. Use cases check the role again (defense in depth).
func RequireRole(role shared.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ActorFrom(c).Role != role {
			httputil.WriteError(c, shared.ErrForbidden)
			return
		}
		c.Next()
	}
}
