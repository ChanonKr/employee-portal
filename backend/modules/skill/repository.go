package skill

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

func (r *repository) ListActive(ctx context.Context) ([]entity.Skill, error) {
	rows, err := r.q.ListActiveSkills(ctx)
	if err != nil {
		return nil, database.MapError(err)
	}
	skills := make([]entity.Skill, len(rows))
	for i, row := range rows {
		skills[i] = entity.Skill{ID: row.ID, Name: row.Name, IsActive: row.IsActive}
	}
	return skills, nil
}
