package history

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
	"employee-portal/pkg/database/sqlc"
)

const settingHistoryDays = "history_days"

type repository struct {
	q *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{q: sqlc.New(pool)}
}

func (r *repository) HistoryDays(ctx context.Context) (string, error) {
	v, err := r.q.GetSetting(ctx, settingHistoryDays)
	return v, database.MapError(err)
}

func (r *repository) List(ctx context.Context, q entity.HistoryQuery, limit int) ([]entity.ChangeLog, error) {
	params := sqlc.ListHistoryParams{
		EmployeeID: q.EmployeeID,
		PageOffset: int32(q.Offset()),
		PageLimit:  int32(limit),
	}
	if q.Days > 0 {
		d := int32(q.Days)
		params.Days = &d
	}
	if q.Field != nil {
		f := string(*q.Field)
		params.Field = &f
	}

	rows, err := r.q.ListHistory(ctx, params)
	if err != nil {
		return nil, database.MapError(err)
	}
	logs := make([]entity.ChangeLog, len(rows))
	for i, row := range rows {
		oldV, err := decode(row.OldValue)
		if err != nil {
			return nil, err
		}
		newV, err := decode(row.NewValue)
		if err != nil {
			return nil, err
		}
		logs[i] = entity.ChangeLog{
			ID:        row.ID,
			Change:    entity.Change{Field: entity.FieldName(row.Field), Old: oldV, New: newV},
			ChangedBy: entity.Actor{UserID: row.ChangedBy, Role: entity.Role(row.ChangedByRole)},
			ChangedAt: row.ChangedAt,
		}
	}
	return logs, nil
}

func (r *repository) HasOlder(ctx context.Context, employeeID int64, days int) (bool, error) {
	ok, err := r.q.HasOlderHistory(ctx, sqlc.HasOlderHistoryParams{EmployeeID: employeeID, Days: int32(days)})
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
