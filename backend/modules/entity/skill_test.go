package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	goSkill      = Skill{ID: 1, Name: "Go", IsActive: true}
	reactSkill   = Skill{ID: 2, Name: "React", IsActive: true}
	flutterSkill = Skill{ID: 3, Name: "Flutter", IsActive: false} // turned off by admin
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
		assert.ErrorIs(t, ValidateSkillIDs(ids), ErrValidation, name)
	}
}

func TestNormalizeSkillName(t *testing.T) {
	got, err := NormalizeSkillName("  Machine   Learning ")
	assert.NoError(t, err)
	assert.Equal(t, "Machine Learning", got)

	for _, in := range []string{"", "   ", strings.Repeat("x", 51)} {
		_, err := NormalizeSkillName(in)
		assert.ErrorIs(t, err, ErrValidation, in)
	}
	_, err = NormalizeSkillName(strings.Repeat("ก", 50))
	assert.NoError(t, err, "length counts characters, not bytes")
}

func TestValidateSkillOrder(t *testing.T) {
	assert.NoError(t, ValidateSkillOrder([]SkillOrder{{1, 0}, {2, 1}}))
	assert.ErrorIs(t, ValidateSkillOrder(nil), ErrValidation)
	assert.ErrorIs(t, ValidateSkillOrder([]SkillOrder{{1, 0}, {1, 1}}), ErrValidation)
	assert.ErrorIs(t, ValidateSkillOrder([]SkillOrder{{1, -1}}), ErrValidation)
	assert.ErrorIs(t, ValidateSkillOrder([]SkillOrder{{0, 1}}), ErrValidation)
}

func TestRequireAdminAndHistoryDays(t *testing.T) {
	assert.NoError(t, RequireAdmin(Actor{Role: RoleAdmin}))
	assert.ErrorIs(t, RequireAdmin(Actor{Role: RoleEmployee}), ErrForbidden)

	assert.NoError(t, ValidateHistoryDays(1))
	assert.NoError(t, ValidateHistoryDays(365))
	assert.ErrorIs(t, ValidateHistoryDays(0), ErrValidation)
	assert.ErrorIs(t, ValidateHistoryDays(400), ErrValidation)
}
