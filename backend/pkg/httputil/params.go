package httputil

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
)

// QueryInt returns 0 when the parameter is absent so the service applies its default.
func QueryInt(c *gin.Context, name string) (int, error) {
	s := c.Query(name)
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, shared.NewValidationError(name, "must be a whole number")
	}
	return n, nil
}

// PathID parses a positive numeric path parameter. On failure it writes 404
// (a malformed id cannot exist) and returns false.
func PathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		WriteError(c, shared.ErrNotFound)
		return 0, false
	}
	return id, true
}
