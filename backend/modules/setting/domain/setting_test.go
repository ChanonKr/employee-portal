package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"employee-portal/modules/shared"
)

func TestValidateHistoryDays(t *testing.T) {
	assert.NoError(t, ValidateHistoryDays(1))
	assert.NoError(t, ValidateHistoryDays(365))
	assert.ErrorIs(t, ValidateHistoryDays(0), shared.ErrValidation)
	assert.ErrorIs(t, ValidateHistoryDays(400), shared.ErrValidation)
}
