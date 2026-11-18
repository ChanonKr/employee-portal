package repo

import (
	"context"

	"gorm.io/gorm"

	"employee-portal/modules/auth/domain"
	"employee-portal/modules/auth/model"
	"employee-portal/modules/auth/service"
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

// userRow is a user with the id of their profile (none for accounts without one).
// The model is a named, embedded field: GORM skips anonymous fields of unexported types.
type userRow struct {
	User       model.User `gorm:"embedded"`
	EmployeeID *int64
}

func (r *Repository) getUser(ctx context.Context, where string, arg any) (domain.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Select("users.*, employees.id AS employee_id").
		Joins("LEFT JOIN employees ON employees.user_id = users.id").
		Where(where, arg).
		Take(&row).Error
	if err != nil {
		return domain.User{}, database.MapError(err)
	}
	u := row.User
	return domain.User{
		ID: u.ID, EmployeeCode: u.EmployeeCode, PasswordHash: u.PasswordHash,
		Role: shared.Role(u.Role), IsActive: u.IsActive,
		MustChangePassword: u.MustChangePassword, EmployeeID: row.EmployeeID,
	}, nil
}

func (r *Repository) GetUserByCode(ctx context.Context, code string) (domain.User, error) {
	return r.getUser(ctx, "users.employee_code = ?", code)
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	return r.getUser(ctx, "users.id = ?", id)
}

func (r *Repository) CreateAccount(ctx context.Context, a domain.NewAccount, passwordHash string) (int64, int64, error) {
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

func (r *Repository) SetPassword(ctx context.Context, userID int64, passwordHash string, mustChange bool) error {
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]any{"password_hash": passwordHash, "must_change_password": mustChange})
	if res.Error != nil {
		return database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) UserIDByEmployee(ctx context.Context, employeeID int64) (int64, error) {
	var e model.Employee
	err := r.db.WithContext(ctx).Select("user_id").Take(&e, employeeID).Error
	return e.UserID, database.MapError(err)
}
