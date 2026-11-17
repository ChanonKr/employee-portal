package skill

import (
	"gorm.io/gorm"

	"employee-portal/modules/entity"
)

// GORM struct for the skill catalog (see profile/model.go for the rules).
type skillModel struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	IsActive  bool `gorm:"default:true"`
	SortOrder int32
}

func (skillModel) TableName() string { return "skills" }

func (s skillModel) toEntity() entity.Skill {
	return entity.Skill{ID: s.ID, Name: s.Name, IsActive: s.IsActive}
}

func displayOrder(db *gorm.DB) *gorm.DB {
	return db.Order("skills.sort_order").Order("skills.name")
}
