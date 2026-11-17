package auth

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

func (r *repository) GetUserByCode(ctx context.Context, code string) (entity.User, error) {
	row, err := r.q.GetUserByCode(ctx, code)
	if err != nil {
		return entity.User{}, database.MapError(err)
	}
	return entity.User{
		ID: row.ID, EmployeeCode: row.EmployeeCode, PasswordHash: row.PasswordHash,
		Role: entity.Role(row.Role), IsActive: row.IsActive,
		MustChangePassword: row.MustChangePassword, EmployeeID: row.EmployeeID,
	}, nil
}

func (r *repository) GetUserByID(ctx context.Context, id int64) (entity.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return entity.User{}, database.MapError(err)
	}
	return entity.User{
		ID: row.ID, EmployeeCode: row.EmployeeCode, PasswordHash: row.PasswordHash,
		Role: entity.Role(row.Role), IsActive: row.IsActive,
		MustChangePassword: row.MustChangePassword, EmployeeID: row.EmployeeID,
	}, nil
}
