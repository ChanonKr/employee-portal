// Package domain holds the skill catalog rules. No framework imports.
package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"employee-portal/modules/shared"
)

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
		return shared.NewValidationError("skillIds", fmt.Sprintf("must contain at most %d skills", MaxSkills))
	}
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return shared.NewValidationError("skillIds", "must contain valid skill ids")
		}
		if _, dup := seen[id]; dup {
			return shared.NewValidationError("skillIds", "must not contain duplicates")
		}
		seen[id] = struct{}{}
	}
	return nil
}

// CheckSelection verifies the requested ids against the skills found in the
// catalog. Every id must exist. An inactive skill is allowed only if the employee
// already has it, so turning a skill off never forces anyone to lose it.
func CheckSelection(requested []int64, found, current []Skill) error {
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
			return shared.NewValidationError("skillIds", fmt.Sprintf("skill %d does not exist", id))
		}
		if _, has := owned[id]; !s.IsActive && !has {
			return shared.NewValidationError("skillIds", fmt.Sprintf("%s is no longer available", s.Name))
		}
	}
	return nil
}

// SameSet reports whether a and b contain the same skills (order does not matter).
func SameSet(a, b []Skill) bool {
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

// Names is never nil, so an empty set is stored as [] rather than null.
func Names(skills []Skill) []string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names
}

const maxSkillNameLength = 50

// NormalizeSkillName trims and collapses inner whitespace, then checks the length.
// Uniqueness (case-insensitive) is enforced by the database.
func NormalizeSkillName(raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	n := utf8.RuneCountInString(name)
	if n == 0 {
		return "", shared.NewValidationError("name", "is required")
	}
	if n > maxSkillNameLength {
		return "", shared.NewValidationError("name", "must be at most 50 characters")
	}
	return name, nil
}

// SkillUsage is a catalog row for the admin screen.
type SkillUsage struct {
	Skill
	SortOrder int32
	UsedBy    int64 // employees who have this skill
}

// SkillOrder sets one skill's position in the picker.
type SkillOrder struct {
	ID        int64
	SortOrder int32
}

func ValidateSkillOrder(items []SkillOrder) error {
	if len(items) == 0 {
		return shared.NewValidationError("order", "must not be empty")
	}
	seen := make(map[int64]struct{}, len(items))
	for _, it := range items {
		if it.ID <= 0 || it.SortOrder < 0 {
			return shared.NewValidationError("order", "ids must be positive and sortOrder 0 or more")
		}
		if _, dup := seen[it.ID]; dup {
			return shared.NewValidationError("order", "must not contain duplicate ids")
		}
		seen[it.ID] = struct{}{}
	}
	return nil
}
