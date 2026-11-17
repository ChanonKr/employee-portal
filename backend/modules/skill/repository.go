package skill

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
	"employee-portal/pkg/database/sqlc"
)

type repository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool, q: sqlc.New(pool)}
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

func (r *repository) ListAll(ctx context.Context) ([]entity.SkillUsage, error) {
	rows, err := r.q.AdminListSkills(ctx)
	if err != nil {
		return nil, database.MapError(err)
	}
	out := make([]entity.SkillUsage, len(rows))
	for i, row := range rows {
		out[i] = entity.SkillUsage{
			Skill:     entity.Skill{ID: row.ID, Name: row.Name, IsActive: row.IsActive},
			SortOrder: row.SortOrder,
			UsedBy:    row.UsedBy,
		}
	}
	return out, nil
}

func (r *repository) Create(ctx context.Context, name string) (int64, error) {
	id, err := r.q.CreateSkill(ctx, name)
	return id, database.MapError(err)
}

func (r *repository) Update(ctx context.Context, id int64, name *string, isActive *bool) error {
	n, err := r.q.UpdateSkill(ctx, sqlc.UpdateSkillParams{ID: id, Name: name, IsActive: isActive})
	if err != nil {
		return database.MapError(err)
	}
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *repository) Reorder(ctx context.Context, items []entity.SkillOrder) error {
	p := sqlc.ReorderSkillsParams{
		Ids:        make([]int64, len(items)),
		SortOrders: make([]int32, len(items)),
	}
	for i, it := range items {
		p.Ids[i], p.SortOrders[i] = it.ID, it.SortOrder
	}
	// In a transaction so an unknown id rolls back the rows that did match:
	// the new order applies completely or not at all.
	return database.WithTx(ctx, r.pool, func(q *sqlc.Queries) error {
		n, err := q.ReorderSkills(ctx, p)
		if err != nil {
			return database.MapError(err)
		}
		if n != int64(len(items)) {
			return entity.ErrNotFound
		}
		return nil
	})
}
