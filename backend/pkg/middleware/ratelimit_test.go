package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyedLimiter(t *testing.T) {
	l := NewKeyedLimiter(5, time.Minute)
	for i := range 5 {
		ok, _ := l.Allow("1.2.3.4")
		assert.True(t, ok, "request %d", i+1)
	}
	ok, wait := l.Allow("1.2.3.4")
	assert.False(t, ok)
	assert.Greater(t, wait, time.Duration(0))

	ok, _ = l.Allow("5.6.7.8")
	assert.True(t, ok, "other keys have their own bucket")
}
