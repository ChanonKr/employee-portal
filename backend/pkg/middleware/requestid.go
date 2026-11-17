package middleware

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"employee-portal/pkg/httputil"
)

const requestIDHeader = "X-Request-ID"

var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID reuses a well-formed incoming X-Request-ID or generates a UUIDv7.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if !safeRequestID.MatchString(id) {
			id = uuid.Must(uuid.NewV7()).String()
		}
		c.Set(httputil.RequestIDKey, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}
