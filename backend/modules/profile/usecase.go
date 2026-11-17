package profile

import (
	"context"

	"employee-portal/modules/entity"
)

// Repository is implemented in repository.go. The use case owns the interface,
// so it never imports sqlc or pgx.
type Repository interface {
	GetProfile(ctx context.Context, employeeID int64) (entity.Profile, error)
	// WithTx runs fn in one transaction: commit if fn returns nil, otherwise roll back.
	WithTx(ctx context.Context, fn func(tx TxRepository) error) error
}

// TxRepository is only available inside WithTx.
type TxRepository interface {
	GetForUpdate(ctx context.Context, employeeID int64) (entity.Employee, error)
	SaveContact(ctx context.Context, e entity.Employee, f entity.ContactField) (newVersion int32, err error)
	AppendChange(ctx context.Context, employeeID int64, c entity.Change, by entity.Actor) error

	ListEmployeeSkills(ctx context.Context, employeeID int64) ([]entity.Skill, error)
	// GetSkillsByIDs returns the matching skills, active or not, in display order.
	GetSkillsByIDs(ctx context.Context, ids []int64) ([]entity.Skill, error)
	ReplaceSkills(ctx context.Context, employeeID int64, skillIDs []int64) error
	// BumpVersion increments the version for edits stored outside the employees row.
	BumpVersion(ctx context.Context, e entity.Employee) (newVersion int32, err error)
}

type UpdateContactInput struct {
	Actor   entity.Actor
	Field   string
	Value   string
	Version int32 // from If-Match
}

type SetSkillsInput struct {
	Actor    entity.Actor
	SkillIDs []int64
	Version  int32 // from If-Match
}

type Usecase interface {
	GetProfile(ctx context.Context, actor entity.Actor) (entity.Profile, error)
	UpdateContact(ctx context.Context, in UpdateContactInput) (newVersion int32, err error)
	SetSkills(ctx context.Context, in SetSkillsInput) (newVersion int32, err error)
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) GetProfile(ctx context.Context, actor entity.Actor) (entity.Profile, error) {
	return u.repo.GetProfile(ctx, actor.EmployeeID)
}

// UpdateContact: lock row → apply domain rule → save (version + 1) → log if changed.
// The update and its history entry commit together or not at all.
func (u *usecase) UpdateContact(ctx context.Context, in UpdateContactInput) (int32, error) {
	field, err := entity.ParseContactField(in.Field)
	if err != nil {
		return 0, err
	}

	var newVersion int32
	err = u.repo.WithTx(ctx, func(tx TxRepository) error {
		emp, err := tx.GetForUpdate(ctx, in.Actor.EmployeeID)
		if err != nil {
			return err
		}
		change, err := emp.ChangeContact(field, in.Value, in.Version)
		if err != nil {
			return err
		}
		if newVersion, err = tx.SaveContact(ctx, emp, field); err != nil {
			return err
		}
		if change == nil {
			return nil
		}
		return tx.AppendChange(ctx, emp.ID, *change, in.Actor)
	})
	if err != nil {
		return 0, err
	}
	return newVersion, nil
}

// SetSkills replaces the employee's whole skill set in one request:
// lock row → load current + requested skills → apply domain rule → replace → version + 1 → log if changed.
func (u *usecase) SetSkills(ctx context.Context, in SetSkillsInput) (int32, error) {
	// Reject a malformed list before opening a transaction (ChangeSkills checks it again).
	if err := entity.ValidateSkillIDs(in.SkillIDs); err != nil {
		return 0, err
	}

	var newVersion int32
	err := u.repo.WithTx(ctx, func(tx TxRepository) error {
		emp, err := tx.GetForUpdate(ctx, in.Actor.EmployeeID)
		if err != nil {
			return err
		}
		current, err := tx.ListEmployeeSkills(ctx, emp.ID)
		if err != nil {
			return err
		}
		selected, err := tx.GetSkillsByIDs(ctx, in.SkillIDs)
		if err != nil {
			return err
		}
		change, err := emp.ChangeSkills(in.SkillIDs, current, selected, in.Version)
		if err != nil {
			return err
		}
		if err := tx.ReplaceSkills(ctx, emp.ID, in.SkillIDs); err != nil {
			return err
		}
		if newVersion, err = tx.BumpVersion(ctx, emp); err != nil {
			return err
		}
		if change == nil {
			return nil
		}
		return tx.AppendChange(ctx, emp.ID, *change, in.Actor)
	})
	if err != nil {
		return 0, err
	}
	return newVersion, nil
}
