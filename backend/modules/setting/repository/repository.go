package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"employee-portal/modules/setting/domain"
	"employee-portal/modules/setting/model"
	"employee-portal/modules/setting/service"
	"employee-portal/modules/shared"
	"employee-portal/pkg/database"
)

// Repository implements service.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

var _ service.Repository = (*Repository)(nil)

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context, key string) (string, domain.HistorySettings, error) {
	var row struct {
		Value         string
		UpdatedAt     time.Time
		UpdatedByID   *int64
		UpdatedByCode *string
	}
	err := r.db.WithContext(ctx).Model(&model.Setting{}).
		Select("app_settings.value, app_settings.updated_at, users.id AS updated_by_id, users.employee_code AS updated_by_code").
		Joins("LEFT JOIN users ON users.id = app_settings.updated_by").
		Where("app_settings.key = ?", key).
		Take(&row).Error
	if err != nil {
		return "", domain.HistorySettings{}, database.MapError(err)
	}
	s := domain.HistorySettings{UpdatedByID: row.UpdatedByID, UpdatedByCode: row.UpdatedByCode}
	if row.UpdatedByID != nil { // seed rows have no editor; their timestamp is the migration time
		at := row.UpdatedAt
		s.UpdatedAt = &at
	}
	return row.Value, s, nil
}

func (r *Repository) Set(ctx context.Context, key, value string, by shared.Actor) error {
	uid := by.UserID
	s := model.Setting{Key: key, Value: value, UpdatedBy: &uid}
	// INSERT ... ON CONFLICT (key) DO UPDATE
	return database.MapError(r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_by", "updated_at"}),
	}).Create(&s).Error)
}
