package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"employee-portal/modules/history/domain"
	"employee-portal/modules/history/model"
	"employee-portal/modules/history/service"
	"employee-portal/modules/shared"
	"employee-portal/pkg/database"
)

const settingHistoryDays = "history_days"

// Repository implements service.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

var _ service.Repository = (*Repository)(nil)

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) HistoryDays(ctx context.Context) (string, error) {
	var row struct{ Value string }
	err := r.db.WithContext(ctx).Table(model.SettingTable).Select("value").Where("key = ?", settingHistoryDays).Take(&row).Error
	return row.Value, database.MapError(err)
}

// List reads the log newest first. Days 0 = no time window (admin); Field nil = all fields.
func (r *Repository) List(ctx context.Context, q domain.HistoryQuery, limit int) ([]domain.ChangeLog, error) {
	db := r.db.WithContext(ctx).Table(model.ChangeLogTable).
		Select("profile_change_logs.id, profile_change_logs.field, profile_change_logs.old_value, "+
			"profile_change_logs.new_value, profile_change_logs.changed_at, profile_change_logs.changed_by, "+
			"users.role AS changed_by_role").
		Joins("JOIN users ON users.id = profile_change_logs.changed_by").
		Where("profile_change_logs.employee_id = ?", q.EmployeeID)
	if q.Days > 0 {
		db = db.Where("profile_change_logs.changed_at >= now() - make_interval(days => ?)", q.Days)
	}
	if q.Field != nil {
		db = db.Where("profile_change_logs.field = ?", string(*q.Field))
	}

	var rows []model.ChangeLogRow
	err := db.Order("profile_change_logs.changed_at DESC").Order("profile_change_logs.id DESC").
		Limit(limit).Offset(q.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, database.MapError(err)
	}

	logs := make([]domain.ChangeLog, len(rows))
	for i, row := range rows {
		oldV, err := decode(row.OldValue)
		if err != nil {
			return nil, err
		}
		newV, err := decode(row.NewValue)
		if err != nil {
			return nil, err
		}
		logs[i] = domain.ChangeLog{
			ID:        row.ID,
			Change:    domain.Change{Field: domain.FieldName(row.Field), Old: oldV, New: newV},
			ChangedBy: shared.Actor{UserID: row.ChangedBy, Role: shared.Role(row.ChangedByRole)},
			ChangedAt: row.ChangedAt,
		}
	}
	return logs, nil
}

// HasOlder drives the empty state: "No changes yet" vs "No changes in the last N days".
func (r *Repository) HasOlder(ctx context.Context, employeeID int64, days int) (bool, error) {
	db := r.db.WithContext(ctx)
	return exists(db, db.Table(model.ChangeLogTable).
		Where("employee_id = ? AND changed_at < now() - make_interval(days => ?)", employeeID, days))
}

func (r *Repository) EmployeeExists(ctx context.Context, employeeID int64) (bool, error) {
	db := r.db.WithContext(ctx)
	return exists(db, db.Table(model.EmployeeTable).Where("id = ?", employeeID))
}

// exists runs SELECT EXISTS (<sub>), which stops at the first matching row.
func exists(db, sub *gorm.DB) (bool, error) {
	var ok bool
	err := db.Raw("SELECT EXISTS (?)", sub.Select("1")).Scan(&ok).Error
	return ok, database.MapError(err)
}

// decode turns a JSONB value into nil, a string or []any (skills).
func decode(b []byte) (any, error) {
	if b == nil {
		return nil, nil
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("decode change value: %w", err)
	}
	return v, nil
}
