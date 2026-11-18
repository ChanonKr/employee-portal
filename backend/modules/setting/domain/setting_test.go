package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"employee-portal/modules/shared"
)

func TestValidateHistoryDays(t *testing.T) {
	assert.NoError(t, ValidateHistoryDays(1, 90))
	assert.NoError(t, ValidateHistoryDays(90, 90))
	assert.ErrorIs(t, ValidateHistoryDays(0, 90), shared.ErrValidation)
	assert.ErrorIs(t, ValidateHistoryDays(91, 90), shared.ErrValidation, "longer than the retention")
}
