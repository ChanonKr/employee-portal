package model

import (
	"gorm.io/gorm"

	"employee-portal/modules/entity"
)

func (e Employee) ToEntity() entity.Employee {
	return entity.Employee{
		ID: e.ID, UserID: e.UserID, FirstName: e.FirstName, LastName: e.LastName,
		PhotoPath: e.PhotoPath, MobilePhone: e.MobilePhone, ContactEmail: e.ContactEmail,
		Address: e.Address, Version: e.Version,
	}
}

func (s Skill) ToEntity() entity.Skill {
	return entity.Skill{ID: s.ID, Name: s.Name, IsActive: s.IsActive}
}

func SkillsToEntity(rows []Skill) []entity.Skill {
	out := make([]entity.Skill, len(rows))
	for i, r := range rows {
		out[i] = r.ToEntity()
	}
	return out
}

// SkillOrder sorts skills the way the catalog shows them.
func SkillOrder(db *gorm.DB) *gorm.DB {
	return db.Order("skills.sort_order").Order("skills.name")
}

// SkillsOf limits a skills query to one employee's selection, in display order.
func SkillsOf(employeeID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN employee_skills es ON es.skill_id = skills.id").
			Where("es.employee_id = ?", employeeID).
			Scopes(SkillOrder)
	}
}
