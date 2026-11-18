package repository

import (
	"context"

	"gorm.io/gorm"

	"employee-portal/modules/shared"
	"employee-portal/modules/skill/domain"
	"employee-portal/modules/skill/model"
	"employee-portal/modules/skill/service"
	"employee-portal/pkg/database"
)

// Repository stores the skill catalog with GORM.
type Repository struct {
	db *gorm.DB
}

var _ service.Repository = (*Repository)(nil)

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListActive returns the picker catalog.
func (r *Repository) ListActive(ctx context.Context) ([]domain.Skill, error) {
	var rows []model.Skill
	if err := r.db.WithContext(ctx).Where("is_active").Scopes(model.DisplayOrder).Find(&rows).Error; err != nil {
		return nil, database.MapError(err)
	}
	skills := make([]domain.Skill, len(rows))
	for i, row := range rows {
		skills[i] = row.ToDomain()
	}
	return skills, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]domain.SkillUsage, error) {
	var rows []struct {
		Skill  model.Skill `gorm:"embedded"` // named: GORM skips anonymous unexported types
		UsedBy int64
	}
	err := r.db.WithContext(ctx).Model(&model.Skill{}).
		Select("skills.*, count(es.employee_id) AS used_by").
		Joins("LEFT JOIN employee_skills es ON es.skill_id = skills.id").
		Group("skills.id").
		Scopes(model.DisplayOrder).
		Find(&rows).Error
	if err != nil {
		return nil, database.MapError(err)
	}
	out := make([]domain.SkillUsage, len(rows))
	for i, row := range rows {
		out[i] = domain.SkillUsage{Skill: row.Skill.ToDomain(), SortOrder: row.Skill.SortOrder, UsedBy: row.UsedBy}
	}
	return out, nil
}

// Create puts the new skill at the end of the list.
func (r *Repository) Create(ctx context.Context, name string) (int64, error) {
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

func (r *Repository) Update(ctx context.Context, id int64, name *string, isActive *bool) error {
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
func (r *Repository) Reorder(ctx context.Context, items []domain.SkillOrder) error {
	return database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		for _, it := range items {
			res := tx.Model(&model.Skill{}).Where("id = ?", it.ID).Update("sort_order", it.SortOrder)
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
