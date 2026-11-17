package auth

import (
	"context"

	"gorm.io/gorm"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
	"employee-portal/pkg/database/model"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// userRow is a user with the id of their profile (none for accounts without one).
type userRow struct {
	model.User
	EmployeeID *int64
}

func (r *repository) getUser(ctx context.Context, where string, arg any) (entity.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Select("users.*, employees.id AS employee_id").
		Joins("LEFT JOIN employees ON employees.user_id = users.id").
		Where(where, arg).
		Take(&row).Error
	if err != nil {
		return entity.User{}, database.MapError(err)
	}
	return entity.User{
		ID: row.ID, EmployeeCode: row.EmployeeCode, PasswordHash: row.PasswordHash,
		Role: entity.Role(row.Role), IsActive: row.IsActive,
		MustChangePassword: row.MustChangePassword, EmployeeID: row.EmployeeID,
	}, nil
}

func (r *repository) GetUserByCode(ctx context.Context, code string) (entity.User, error) {
	return r.getUser(ctx, "users.employee_code = ?", code)
}

func (r *repository) GetUserByID(ctx context.Context, id int64) (entity.User, error) {
	return r.getUser(ctx, "users.id = ?", id)
}

func (r *repository) CreateAccount(ctx context.Context, a entity.NewAccount, passwordHash string) (int64, int64, error) {
	// New accounts must change the temporary password on first sign-in.
	u := model.User{EmployeeCode: a.EmployeeCode, PasswordHash: passwordHash, Role: string(a.Role), MustChangePassword: true}
	e := model.Employee{FirstName: a.FirstName, LastName: a.LastName}
	err := database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.Create(&u).Error; err != nil {
			return database.MapError(err)
		}
		e.UserID = u.ID
		return database.MapError(tx.Create(&e).Error)
	})
	return u.ID, e.ID, err
}

func (r *repository) SetPassword(ctx context.Context, userID int64, passwordHash string, mustChange bool) error {
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]any{"password_hash": passwordHash, "must_change_password": mustChange})
	if res.Error != nil {
		return database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *repository) UserIDByEmployee(ctx context.Context, employeeID int64) (int64, error) {
	var e model.Employee
	err := r.db.WithContext(ctx).Select("user_id").Take(&e, employeeID).Error
	return e.UserID, database.MapError(err)
}
