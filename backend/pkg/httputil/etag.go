package httputil

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	ErrMissingIfMatch = errors.New("missing If-Match header")
	ErrInvalidIfMatch = errors.New("invalid If-Match header")
)

// ETag formats a profile version as a strong ETag: "7".
func ETag(version int32) string {
	return `"` + strconv.FormatInt(int64(version), 10) + `"`
}

// IfMatchVersion reads the expected profile version from the If-Match header.
func IfMatchVersion(c *gin.Context) (int32, error) {
	return ParseIfMatch(c.GetHeader("If-Match"))
}

func ParseIfMatch(h string) (int32, error) {
	h = strings.TrimSpace(h)
	if h == "" {
		return 0, ErrMissingIfMatch
	}
	if len(h) < 3 || h[0] != '"' || h[len(h)-1] != '"' {
		return 0, ErrInvalidIfMatch
	}
	v, err := strconv.ParseInt(h[1:len(h)-1], 10, 32)
	if err != nil || v < 1 {
		return 0, ErrInvalidIfMatch
	}
	return int32(v), nil
}
