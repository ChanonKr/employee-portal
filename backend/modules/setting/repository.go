package setting

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
	"employee-portal/pkg/database/sqlc"
)

type repository struct {
	q *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{q: sqlc.New(pool)}
}

func (r *repository) Get(ctx context.Context, key string) (string, entity.HistorySettings, error) {
	row, err := r.q.GetSettingDetail(ctx, key)
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
	return database.MapError(r.q.UpsertSetting(ctx, sqlc.UpsertSettingParams{Key: key, Value: value, UpdatedBy: &uid}))
}
