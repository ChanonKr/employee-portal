package middleware

import (
	"github.com/gin-gonic/gin"

	"employee-portal/modules/entity"
	"employee-portal/pkg/httputil"
)

// RequireRole must run after Auth. Use cases check the role again (defense in depth).
func RequireRole(role entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ActorFrom(c).Role != role {
			httputil.WriteError(c, entity.ErrForbidden)
			return
		}
		c.Next()
	}
}
