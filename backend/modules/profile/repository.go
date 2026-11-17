package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
	"employee-portal/pkg/database/sqlc"
)

// contactColumn maps a domain field to its column. It is the only source of column
// names passed to UpdateContactField, so user input never reaches SQL.
var contactColumn = map[entity.ContactField]string{
	entity.ContactMobile:  "mobile_phone",
	entity.ContactEmail:   "contact_email",
	entity.ContactAddress: "address",
}

type repository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool, q: sqlc.New(pool)}
}

func (r *repository) GetProfile(ctx context.Context, employeeID int64) (entity.Profile, error) {
	row, err := r.q.GetProfile(ctx, employeeID)
	if err != nil {
		return entity.Profile{}, database.MapError(err)
	}
	skills, err := r.q.ListEmployeeSkills(ctx, employeeID)
	if err != nil {
		return entity.Profile{}, database.MapError(err)
	}
	p := entity.Profile{
		Employee:     toEmployee(row.Employee),
		EmployeeCode: row.EmployeeCode,
		Skills:       make([]entity.Skill, len(skills)),
	}
	for i, s := range skills {
		p.Skills[i] = entity.Skill{ID: s.ID, Name: s.Name}
	}
	return p, nil
}

func (r *repository) WithTx(ctx context.Context, fn func(TxRepository) error) error {
	return database.WithTx(ctx, r.pool, func(q *sqlc.Queries) error {
		return fn(&txRepository{q: q})
	})
}

type txRepository struct {
	q *sqlc.Queries
}

func (t *txRepository) GetForUpdate(ctx context.Context, employeeID int64) (entity.Employee, error) {
	row, err := t.q.GetProfileForUpdate(ctx, employeeID)
	if err != nil {
		return entity.Employee{}, database.MapError(err)
	}
	return toEmployee(row), nil
}

func (t *txRepository) SaveContact(ctx context.Context, e entity.Employee, f entity.ContactField) (int32, error) {
	col, ok := contactColumn[f]
	if !ok {
		return 0, fmt.Errorf("no column for contact field %q", f)
	}
	v, err := t.q.UpdateContactField(ctx, sqlc.UpdateContactFieldParams{
		ID: e.ID, Field: col, Value: e.Contact(f), Version: e.Version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		// Version changed between lock and update. Cannot happen under FOR UPDATE,
		// but report it as a conflict rather than "not found" if it ever does.
		return 0, entity.ErrConflict
	}
	if err != nil {
		return 0, database.MapError(err)
	}
	return v, nil
}

func (t *txRepository) AppendChange(ctx context.Context, employeeID int64, c entity.Change, by entity.Actor) error {
	oldJSON, err := jsonOrNull(c.Old)
	if err != nil {
		return err
	}
	newJSON, err := jsonOrNull(c.New)
	if err != nil {
		return err
	}
	return database.MapError(t.q.InsertChangeLog(ctx, sqlc.InsertChangeLogParams{
		EmployeeID: employeeID,
		Field:      string(c.Field),
		OldValue:   oldJSON,
		NewValue:   newJSON,
		ChangedBy:  by.UserID,
	}))
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

func toEmployee(e sqlc.Employee) entity.Employee {
	return entity.Employee{
		ID: e.ID, UserID: e.UserID, FirstName: e.FirstName, LastName: e.LastName,
		PhotoPath: e.PhotoPath, MobilePhone: e.MobilePhone, ContactEmail: e.ContactEmail,
		Address: e.Address, Version: e.Version,
	}
}
