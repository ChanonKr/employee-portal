package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSameOriginOrCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SameOriginOrCORS(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"}, AllowMethods: []string{"POST"}, MaxAge: time.Hour,
	})))
	r.POST("/x", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	send := func(host, origin string) int {
		req := httptest.NewRequest(http.MethodPost, "/x", nil)
		req.Host = host
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusNoContent, send("portal.example.com", "https://portal.example.com"), "same origin, not in the allowlist")
	assert.Equal(t, http.StatusNoContent, send("localhost:8080", ""), "no Origin (curl, server-to-server)")
	assert.Equal(t, http.StatusNoContent, send("localhost:8080", "http://localhost:5173"), "allowed cross origin")
	assert.Equal(t, http.StatusForbidden, send("localhost:8080", "https://evil.example"), "other cross origin")
}
