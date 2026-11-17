package middleware

import (
	"github.com/gin-gonic/gin"

	"employee-portal/modules/entity"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/security"
)

// SessionCookie holds the JWT. It is HttpOnly so scripts cannot read it.
const SessionCookie = "session"

const actorKey = "actor"

type actorValue = entity.Actor

// Auth requires a valid session cookie and stores the Actor in the context.
func Auth(tokens *security.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(SessionCookie)
		if err != nil || raw == "" {
			httputil.WriteError(c, entity.ErrUnauthorized)
			return
		}
		actor, err := tokens.Parse(raw)
		if err != nil {
			httputil.WriteError(c, entity.ErrUnauthorized)
			return
		}
		c.Set(actorKey, actor)
		c.Next()
	}
}

// ActorFrom returns the Actor set by Auth. Only call it on routes behind Auth.
func ActorFrom(c *gin.Context) entity.Actor {
	return c.MustGet(actorKey).(entity.Actor)
}

// RequireProfile rejects accounts that have no employee profile.
func RequireProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ActorFrom(c).HasProfile() {
			httputil.WriteError(c, entity.ErrNotFound)
			return
		}
		c.Next()
	}
}
