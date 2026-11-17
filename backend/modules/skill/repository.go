package skill

import (
	"context"

	"gorm.io/gorm"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
	"employee-portal/pkg/database/model"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ListActive returns the picker catalog.
func (r *repository) ListActive(ctx context.Context) ([]entity.Skill, error) {
	var rows []model.Skill
	if err := r.db.WithContext(ctx).Where("is_active").Scopes(model.SkillOrder).Find(&rows).Error; err != nil {
		return nil, database.MapError(err)
	}
	return model.SkillsToEntity(rows), nil
}

func (r *repository) ListAll(ctx context.Context) ([]entity.SkillUsage, error) {
	var rows []struct {
		model.Skill
		UsedBy int64
	}
	err := r.db.WithContext(ctx).Model(&model.Skill{}).
		Select("skills.*, count(es.employee_id) AS used_by").
		Joins("LEFT JOIN employee_skills es ON es.skill_id = skills.id").
		Group("skills.id").
		Scopes(model.SkillOrder).
		Find(&rows).Error
	if err != nil {
		return nil, database.MapError(err)
	}
	out := make([]entity.SkillUsage, len(rows))
	for i, row := range rows {
		out[i] = entity.SkillUsage{Skill: row.Skill.ToEntity(), SortOrder: row.SortOrder, UsedBy: row.UsedBy}
	}
	return out, nil
}

// Create puts the new skill at the end of the list.
func (r *repository) Create(ctx context.Context, name string) (int64, error) {
	s := model.Skill{Name: name}
	err := database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		var last int32
		if err := tx.Model(&model.Skill{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&last).Error; err != nil {
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
	db := r.db.WithContext(ctx).Model(&model.Skill{}).Where("id = ?", id)
	if len(fields) == 0 { // nothing to change: only report whether the skill exists
		var n int64
		if err := db.Count(&n).Error; err != nil {
			return database.MapError(err)
		}
		if n == 0 {
			return entity.ErrNotFound
		}
		return nil
	}
	res := db.Updates(fields)
	if res.Error != nil {
		return database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Reorder applies every position or none: an unknown id rolls back the transaction.
func (r *repository) Reorder(ctx context.Context, items []entity.SkillOrder) error {
	return database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		for _, it := range items {
			res := tx.Model(&model.Skill{}).Where("id = ?", it.ID).Update("sort_order", it.SortOrder)
			if res.Error != nil {
				return database.MapError(res.Error)
			}
			if res.RowsAffected == 0 {
				return entity.ErrNotFound
			}
		}
		return nil
	})
}
