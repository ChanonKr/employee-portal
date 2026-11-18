package skill

import (
	"context"

	"gorm.io/gorm"

	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
	"employee-portal/pkg/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ListActive returns the picker catalog.
func (r *repository) ListActive(ctx context.Context) ([]skilldomain.Skill, error) {
	var rows []skillModel
	if err := r.db.WithContext(ctx).Where("is_active").Scopes(displayOrder).Find(&rows).Error; err != nil {
		return nil, database.MapError(err)
	}
	skills := make([]skilldomain.Skill, len(rows))
	for i, row := range rows {
		skills[i] = row.toEntity()
	}
	return skills, nil
}

func (r *repository) ListAll(ctx context.Context) ([]skilldomain.SkillUsage, error) {
	var rows []struct {
		Skill  skillModel `gorm:"embedded"` // named: GORM skips anonymous unexported types
		UsedBy int64
	}
	err := r.db.WithContext(ctx).Model(&skillModel{}).
		Select("skills.*, count(es.employee_id) AS used_by").
		Joins("LEFT JOIN employee_skills es ON es.skill_id = skills.id").
		Group("skills.id").
		Scopes(displayOrder).
		Find(&rows).Error
	if err != nil {
		return nil, database.MapError(err)
	}
	out := make([]skilldomain.SkillUsage, len(rows))
	for i, row := range rows {
		out[i] = skilldomain.SkillUsage{Skill: row.Skill.toEntity(), SortOrder: row.Skill.SortOrder, UsedBy: row.UsedBy}
	}
	return out, nil
}

// Create puts the new skill at the end of the list.
func (r *repository) Create(ctx context.Context, name string) (int64, error) {
	s := skillModel{Name: name}
	err := database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		var last int32
		if err := tx.Model(&skillModel{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&last).Error; err != nil {
			return database.MapError(err)
		}
		s.SortOrder = last + 1
		return database.MapError(tx.Create(&s).Error)
	})
	return s.ID, err
}

func (r *repository) Update(ctx context.Context, id int64, name *string, isActive *bool) error {
	fields := map[string]any{}
	if name != nil {
		fields["name"] = *name
	}
	if isActive != nil {
		fields["is_active"] = *isActive
	}
	db := r.db.WithContext(ctx).Model(&skillModel{}).Where("id = ?", id)
	if len(fields) == 0 { // nothing to change: only report whether the skill exists
		var n int64
		if err := db.Count(&n).Error; err != nil {
			return database.MapError(err)
		}
		if n == 0 {
			return shared.ErrNotFound
		}
		return nil
	}
	res := db.Updates(fields)
	if res.Error != nil {
		return database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// Reorder applies every position or none: an unknown id rolls back the transaction.
func (r *repository) Reorder(ctx context.Context, items []skilldomain.SkillOrder) error {
	return database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		for _, it := range items {
			res := tx.Model(&skillModel{}).Where("id = ?", it.ID).Update("sort_order", it.SortOrder)
			if res.Error != nil {
				return database.MapError(res.Error)
			}
			if res.RowsAffected == 0 {
				return shared.ErrNotFound
			}
		}
		return nil
	})
}
