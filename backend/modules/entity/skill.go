package entity

import "fmt"

// MaxSkills is the most skills one employee can select.
const MaxSkills = 20

type Skill struct {
	ID       int64
	Name     string
	IsActive bool // inactive skills are hidden from the picker but kept by employees who have them
}

// ValidateSkillIDs checks the request shape: at most MaxSkills, positive, no duplicates.
// An empty list is valid (it clears all skills).
func ValidateSkillIDs(ids []int64) error {
	if len(ids) > MaxSkills {
		return NewValidationError("skillIds", fmt.Sprintf("must contain at most %d skills", MaxSkills))
	}
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return NewValidationError("skillIds", "must contain valid skill ids")
		}
		if _, dup := seen[id]; dup {
			return NewValidationError("skillIds", "must not contain duplicates")
		}
		seen[id] = struct{}{}
	}
	return nil
}

// CheckSkillSelection verifies the requested ids against the skills found in the
// catalog. Every id must exist. An inactive skill is allowed only if the employee
// already has it, so turning a skill off never forces anyone to lose it.
func CheckSkillSelection(requested []int64, found, current []Skill) error {
	byID := make(map[int64]Skill, len(found))
	for _, s := range found {
		byID[s.ID] = s
	}
	owned := make(map[int64]struct{}, len(current))
	for _, s := range current {
		owned[s.ID] = struct{}{}
	}
	for _, id := range requested {
		s, ok := byID[id]
		if !ok {
			return NewValidationError("skillIds", fmt.Sprintf("skill %d does not exist", id))
		}
		if _, has := owned[id]; !s.IsActive && !has {
			return NewValidationError("skillIds", fmt.Sprintf("%s is no longer available", s.Name))
		}
	}
	return nil
}

// SkillsChange returns the history entry for replacing old with next, or nil when
// both contain the same skills (order does not matter).
func SkillsChange(old, next []Skill) *Change {
	if sameSkillSet(old, next) {
		return nil
	}
	return &Change{Field: FieldSkills, Old: skillNames(old), New: skillNames(next)}
}

func sameSkillSet(a, b []Skill) bool {
	if len(a) != len(b) {
		return false
	}
	ids := make(map[int64]struct{}, len(a))
	for _, s := range a {
		ids[s.ID] = struct{}{}
	}
	for _, s := range b {
		if _, ok := ids[s.ID]; !ok {
			return false
		}
	}
	return true
}

// skillNames is never nil, so an empty set is stored as [] rather than null.
func skillNames(skills []Skill) []string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names
}
