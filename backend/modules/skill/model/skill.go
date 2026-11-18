// Package model holds the GORM structs for the skill catalog. They mirror the SQL
// migrations (GORM never migrates the schema) and are used only by repo.
package model

import (
	"gorm.io/gorm"

	"employee-portal/modules/skill/domain"
)

type Skill struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	IsActive  bool `gorm:"default:true"`
	SortOrder int32
}

func (Skill) TableName() string { return "skills" }

func (s Skill) ToDomain() domain.Skill {
	return domain.Skill{ID: s.ID, Name: s.Name, IsActive: s.IsActive}
}

func DisplayOrder(db *gorm.DB) *gorm.DB {
	return db.Order("skills.sort_order").Order("skills.name")
}
