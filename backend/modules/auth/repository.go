package auth

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

func (r *repository) CreateAccount(ctx context.Context, a entity.NewAccount, passwordHash string) (int64, int64, error) {
	var userID, employeeID int64
	err := database.WithTx(ctx, r.pool, func(q *sqlc.Queries) error {
		var err error
		userID, err = q.CreateUser(ctx, sqlc.CreateUserParams{
			EmployeeCode: a.EmployeeCode, PasswordHash: passwordHash, Role: string(a.Role),
		})
		if err != nil {
			return database.MapError(err)
		}
		employeeID, err = q.CreateEmployee(ctx, sqlc.CreateEmployeeParams{
			UserID: userID, FirstName: a.FirstName, LastName: a.LastName,
		})
		return database.MapError(err)
	})
	return userID, employeeID, err
}

func (r *repository) SetPassword(ctx context.Context, userID int64, passwordHash string, mustChange bool) error {
	n, err := r.q.SetPassword(ctx, sqlc.SetPasswordParams{ID: userID, PasswordHash: passwordHash, MustChangePassword: mustChange})
	if err != nil {
		return database.MapError(err)
	}
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *repository) UserIDByEmployee(ctx context.Context, employeeID int64) (int64, error) {
	id, err := r.q.GetUserIDByEmployee(ctx, employeeID)
	return id, database.MapError(err)
}
