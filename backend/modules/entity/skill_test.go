package entity

import (
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
