package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetentionCutoff(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	assert.Equal(t, time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC), RetentionCutoff(now, 90))
}
