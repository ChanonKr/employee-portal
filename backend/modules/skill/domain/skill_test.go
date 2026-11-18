package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"employee-portal/modules/shared"
)

func TestValidateSkillIDs(t *testing.T) {
	assert.NoError(t, ValidateSkillIDs(nil))
	assert.NoError(t, ValidateSkillIDs([]int64{1, 2, 3}))

	tooMany := make([]int64, MaxSkills+1)
	for i := range tooMany {
		tooMany[i] = int64(i + 1)
	}
	for name, ids := range map[string][]int64{
		"too many":    tooMany,
		"duplicate":   {1, 2, 1},
		"zero id":     {0},
		"negative id": {-5},
	} {
		assert.ErrorIs(t, ValidateSkillIDs(ids), shared.ErrValidation, name)
	}
}

func TestNormalizeSkillName(t *testing.T) {
	got, err := NormalizeSkillName("  Machine   Learning ")
	assert.NoError(t, err)
	assert.Equal(t, "Machine Learning", got)

	for _, in := range []string{"", "   ", strings.Repeat("x", 51)} {
		_, err := NormalizeSkillName(in)
		assert.ErrorIs(t, err, shared.ErrValidation, in)
	}
	_, err = NormalizeSkillName(strings.Repeat("ก", 50))
	assert.NoError(t, err, "length counts characters, not bytes")
}

func TestValidateSkillOrder(t *testing.T) {
	assert.NoError(t, ValidateSkillOrder([]SkillOrder{{1, 0}, {2, 1}}))
	assert.ErrorIs(t, ValidateSkillOrder(nil), shared.ErrValidation)
	assert.ErrorIs(t, ValidateSkillOrder([]SkillOrder{{1, 0}, {1, 1}}), shared.ErrValidation)
	assert.ErrorIs(t, ValidateSkillOrder([]SkillOrder{{1, -1}}), shared.ErrValidation)
	assert.ErrorIs(t, ValidateSkillOrder([]SkillOrder{{0, 1}}), shared.ErrValidation)
}
