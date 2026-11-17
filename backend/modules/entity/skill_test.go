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

func TestCheckSkillSelection(t *testing.T) {
	found := []Skill{goSkill, reactSkill, flutterSkill}

	t.Run("active skills are allowed", func(t *testing.T) {
		assert.NoError(t, CheckSkillSelection([]int64{1, 2}, found, nil))
	})
	t.Run("unknown id is rejected", func(t *testing.T) {
		assert.ErrorIs(t, CheckSkillSelection([]int64{1, 99}, found, nil), ErrValidation)
	})
	t.Run("inactive skill cannot be newly added", func(t *testing.T) {
		err := CheckSkillSelection([]int64{3}, found, []Skill{goSkill})
		assert.ErrorIs(t, err, ErrValidation)
		assert.Contains(t, err.Error(), "Flutter")
	})
	t.Run("inactive skill already owned can be kept", func(t *testing.T) {
		assert.NoError(t, CheckSkillSelection([]int64{1, 3}, found, []Skill{goSkill, flutterSkill}))
	})
}

func TestSkillsChange(t *testing.T) {
	t.Run("same set in a different order is no change", func(t *testing.T) {
		assert.Nil(t, SkillsChange([]Skill{goSkill, reactSkill}, []Skill{reactSkill, goSkill}))
	})
	t.Run("records names before and after", func(t *testing.T) {
		c := SkillsChange([]Skill{goSkill, flutterSkill}, []Skill{goSkill, reactSkill})
		assert.Equal(t, &Change{Field: FieldSkills, Old: []string{"Go", "Flutter"}, New: []string{"Go", "React"}}, c)
	})
	t.Run("empty side is an empty list, not nil", func(t *testing.T) {
		c := SkillsChange(nil, []Skill{goSkill})
		assert.Equal(t, []string{}, c.Old)
		c = SkillsChange([]Skill{goSkill}, nil)
		assert.Equal(t, []string{}, c.New)
	})
}
