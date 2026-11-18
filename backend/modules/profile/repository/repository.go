package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/profile/domain"
	"employee-portal/modules/profile/model"
	"employee-portal/modules/profile/service"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
	"employee-portal/pkg/database"
)

// contactColumn maps a domain field to its column. It is the only source of column
// names used in updates, so user input never reaches SQL.
var contactColumn = map[domain.ContactField]string{
	domain.ContactMobile:  "mobile_phone",
	domain.ContactEmail:   "contact_email",
	domain.ContactAddress: "address",
}

// Repository implements service.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

var _ service.Repository = (*Repository)(nil)

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetProfile(ctx context.Context, employeeID int64) (domain.Profile, error) {
	db := r.db.WithContext(ctx)
	var row struct {
		Employee     model.Employee `gorm:"embedded"` // named: GORM skips anonymous unexported types
		EmployeeCode string
	}
	err := db.Model(&model.Employee{}).
		Select("employees.*, users.employee_code").
		Joins("JOIN users ON users.id = employees.user_id").
		Where("employees.id = ?", employeeID).
		Take(&row).Error
	if err != nil {
		return domain.Profile{}, database.MapError(err)
	}
	var skills []model.Skill
	if err := db.Scopes(model.SkillsOf(employeeID)).Find(&skills).Error; err != nil {
		return domain.Profile{}, database.MapError(err)
	}
	return domain.Profile{
		Employee:     row.Employee.ToDomain(),
		EmployeeCode: row.EmployeeCode,
		Skills:       model.ToSkills(skills),
	}, nil
}

func (r *Repository) WithTx(ctx context.Context, fn func(service.TxRepository) error) error {
	return database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		return fn(&txRepository{db: tx})
	})
}

// txRepository runs on the transaction's *gorm.DB (already bound to the context).
type txRepository struct {
	db *gorm.DB
}

func (t *txRepository) GetForUpdate(ctx context.Context, employeeID int64) (domain.Employee, error) {
	var e model.Employee
	// SELECT ... FOR UPDATE: concurrent edits of one profile are serialized.
	err := t.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Take(&e, employeeID).Error
	if err != nil {
		return domain.Employee{}, database.MapError(err)
	}
	return e.ToDomain(), nil
}

// update writes fields only if the version is still e.Version and bumps it.
// No row matched = someone else changed it first (cannot happen under FOR UPDATE,
// but report a conflict rather than "not found" if it ever does).
func (t *txRepository) update(ctx context.Context, e domain.Employee, fields map[string]any) (int32, error) {
	fields["version"] = gorm.Expr("version + 1")
	fields["updated_at"] = gorm.Expr("now()")
	res := t.db.WithContext(ctx).Model(&model.Employee{}).
		Where("id = ? AND version = ?", e.ID, e.Version).
		Updates(fields)
	if res.Error != nil {
		return 0, database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return 0, shared.ErrConflict
	}
	return e.Version + 1, nil
}

func (t *txRepository) SaveContact(ctx context.Context, e domain.Employee, f domain.ContactField) (int32, error) {
	col, ok := contactColumn[f]
	if !ok {
		return 0, fmt.Errorf("no column for contact field %q", f)
	}
	return t.update(ctx, e, map[string]any{col: e.Contact(f)})
}

func (t *txRepository) SavePhoto(ctx context.Context, e domain.Employee) (int32, error) {
	return t.update(ctx, e, map[string]any{"photo_path": e.PhotoPath})
}

func (t *txRepository) SaveName(ctx context.Context, e domain.Employee) (int32, error) {
	return t.update(ctx, e, map[string]any{"first_name": e.FirstName, "last_name": e.LastName})
}

func (t *txRepository) BumpVersion(ctx context.Context, e domain.Employee) (int32, error) {
	return t.update(ctx, e, map[string]any{})
}

func (t *txRepository) AppendChange(ctx context.Context, employeeID int64, c historydomain.Change, by shared.Actor) error {
	oldJSON, err := jsonOrNull(c.Old)
	if err != nil {
		return err
	}
	newJSON, err := jsonOrNull(c.New)
	if err != nil {
		return err
	}
	return database.MapError(t.db.WithContext(ctx).Create(&model.ChangeLog{
		EmployeeID: employeeID,
		Field:      string(c.Field),
		OldValue:   oldJSON,
		NewValue:   newJSON,
		ChangedBy:  by.UserID,
	}).Error)
}

func (t *txRepository) ListEmployeeSkills(ctx context.Context, employeeID int64) ([]skilldomain.Skill, error) {
	var rows []model.Skill
	if err := t.db.WithContext(ctx).Scopes(model.SkillsOf(employeeID)).Find(&rows).Error; err != nil {
		return nil, database.MapError(err)
	}
	return model.ToSkills(rows), nil
}

func (t *txRepository) GetSkillsByIDs(ctx context.Context, ids []int64) ([]skilldomain.Skill, error) {
	if len(ids) == 0 {
		return []skilldomain.Skill{}, nil
	}
	// Includes inactive skills so the service can tell "unknown" from "inactive".
	var rows []model.Skill
	if err := t.db.WithContext(ctx).Where("id IN ?", ids).Scopes(model.SkillOrder).Find(&rows).Error; err != nil {
		return nil, database.MapError(err)
	}
	return model.ToSkills(rows), nil
}

func (t *txRepository) ReplaceSkills(ctx context.Context, employeeID int64, skillIDs []int64) error {
	db := t.db.WithContext(ctx)
	if err := db.Where("employee_id = ?", employeeID).Delete(&model.EmployeeSkill{}).Error; err != nil {
		return database.MapError(err)
	}
	if len(skillIDs) == 0 {
		return nil
	}
	rows := make([]model.EmployeeSkill, len(skillIDs))
	for i, id := range skillIDs {
		rows[i] = model.EmployeeSkill{EmployeeID: employeeID, SkillID: id}
	}
	return database.MapError(db.Create(&rows).Error)
}

// jsonOrNull returns nil (SQL NULL) for a nil value, otherwise its JSON encoding.
func jsonOrNull(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode change value: %w", err)
	}
	return b, nil
}
