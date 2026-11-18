package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmployeeQuery_LikePattern(t *testing.T) {
	assert.Nil(t, EmployeeQuery{Search: "   "}.LikePattern())
	assert.Equal(t, "%som%", *EmployeeQuery{Search: " som "}.LikePattern())
	assert.Equal(t, `%100\%\_\\%`, *EmployeeQuery{Search: `100%_\`}.LikePattern(), "user wildcards match literally")
}
