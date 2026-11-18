package middleware

import (
	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/security"
)

// SessionCookie holds the JWT. It is HttpOnly so scripts cannot read it.
const SessionCookie = "session"

const actorKey = "actor"

type actorValue = shared.Actor

// Auth requires a valid session cookie and stores the Actor in the context.
func Auth(tokens *security.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(SessionCookie)
		if err != nil || raw == "" {
			httputil.WriteError(c, shared.ErrUnauthorized)
			return
		}
		actor, err := tokens.Parse(raw)
		if err != nil {
			httputil.WriteError(c, shared.ErrUnauthorized)
			return
		}
		c.Set(actorKey, actor)
		c.Next()
	}
}

// RequirePasswordChanged blocks accounts that still use a temporary password.
// Every route except /auth/* uses it.
func RequirePasswordChanged() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ActorFrom(c).MustChangePassword {
			httputil.WriteError(c, shared.ErrPasswordChangeRequired)
			return
		}
		c.Next()
	}
}

// ActorFrom returns the Actor set by Auth. Only call it on routes behind Auth.
func ActorFrom(c *gin.Context) shared.Actor {
	return c.MustGet(actorKey).(shared.Actor)
}

// RequireProfile rejects accounts that have no employee profile.
func RequireProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ActorFrom(c).HasProfile() {
			httputil.WriteError(c, shared.ErrNotFound)
			return
		}
		c.Next()
	}
}
