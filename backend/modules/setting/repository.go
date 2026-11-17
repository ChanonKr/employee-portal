package setting

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Get(ctx context.Context, key string) (string, entity.HistorySettings, error) {
	var row struct {
		Value         string
		UpdatedAt     time.Time
		UpdatedByID   *int64
		UpdatedByCode *string
	}
	err := r.db.WithContext(ctx).Model(&settingModel{}).
		Select("app_settings.value, app_settings.updated_at, users.id AS updated_by_id, users.employee_code AS updated_by_code").
		Joins("LEFT JOIN users ON users.id = app_settings.updated_by").
		Where("app_settings.key = ?", key).
		Take(&row).Error
	if err != nil {
		return "", entity.HistorySettings{}, database.MapError(err)
	}
	s := entity.HistorySettings{UpdatedByID: row.UpdatedByID, UpdatedByCode: row.UpdatedByCode}
	if row.UpdatedByID != nil { // seed rows have no editor; their timestamp is the migration time
		at := row.UpdatedAt
		s.UpdatedAt = &at
	}
	return row.Value, s, nil
}

func (r *repository) Set(ctx context.Context, key, value string, by entity.Actor) error {
	uid := by.UserID
	s := settingModel{Key: key, Value: value, UpdatedBy: &uid}
	// INSERT ... ON CONFLICT (key) DO UPDATE
	return database.MapError(r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_by", "updated_at"}),
	}).Create(&s).Error)
}
