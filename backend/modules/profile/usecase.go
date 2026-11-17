package profile

import (
	"context"
	"io"
	"log/slog"

	"employee-portal/modules/entity"
)

// Repository is implemented in repository.go. The use case owns the interface,
// so it never imports GORM or pgx.
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

	SavePhoto(ctx context.Context, e entity.Employee) (newVersion int32, err error)
	SaveName(ctx context.Context, e entity.Employee) (newVersion int32, err error)
}

// PhotoProcessor checks an upload and re-encodes it (pkg/imaging).
type PhotoProcessor interface {
	Normalize(r io.Reader) (jpeg []byte, err error)
}

// FileStorage keeps photo files (pkg/storage).
type FileStorage interface {
	Save(ctx context.Context, data []byte, ext string) (name string, err error)
	Delete(ctx context.Context, name string) error
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

type UpdatePhotoInput struct {
	Actor   entity.Actor
	File    io.Reader
	Version int32 // from If-Match
}

// UpdateNameInput is the admin-only legal name change.
type UpdateNameInput struct {
	Actor      entity.Actor // must be an admin
	EmployeeID int64        // from the URL
	FirstName  string
	LastName   string
	Version    int32 // from If-Match
}

type Usecase interface {
	GetProfile(ctx context.Context, actor entity.Actor) (entity.Profile, error)
	UpdateContact(ctx context.Context, in UpdateContactInput) (newVersion int32, err error)
	SetSkills(ctx context.Context, in SetSkillsInput) (newVersion int32, err error)
	UpdatePhoto(ctx context.Context, in UpdatePhotoInput) (newVersion int32, photoName string, err error)
	UpdateName(ctx context.Context, in UpdateNameInput) (newVersion int32, err error)
}

type Deps struct {
	Repo   Repository
	Photos PhotoProcessor
	Files  FileStorage
	Log    *slog.Logger
}

type usecase struct {
	repo   Repository
	photos PhotoProcessor
	files  FileStorage
	log    *slog.Logger
}

func NewUsecase(d Deps) Usecase {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	return &usecase{repo: d.Repo, photos: d.Photos, files: d.Files, log: d.Log}
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

// UpdatePhoto keeps the file and the database consistent:
//
//  1. check + resize the upload (before touching the DB)
//  2. save the new file
//  3. tx: lock → check version → set photo_path → log
//     failure → delete the new file, so no orphan is left
//     success → delete the old file (best effort: an orphan file is harmless)
func (u *usecase) UpdatePhoto(ctx context.Context, in UpdatePhotoInput) (int32, string, error) {
	jpeg, err := u.photos.Normalize(in.File)
	if err != nil {
		return 0, "", err
	}
	name, err := u.files.Save(ctx, jpeg, ".jpg")
	if err != nil {
		return 0, "", err
	}

	var (
		newVersion int32
		oldName    *string
	)
	err = u.repo.WithTx(ctx, func(tx TxRepository) error {
		emp, err := tx.GetForUpdate(ctx, in.Actor.EmployeeID)
		if err != nil {
			return err
		}
		oldName = emp.PhotoPath
		change, err := emp.ChangePhoto(name, in.Version)
		if err != nil {
			return err
		}
		if newVersion, err = tx.SavePhoto(ctx, emp); err != nil {
			return err
		}
		return tx.AppendChange(ctx, emp.ID, *change, in.Actor)
	})
	if err != nil {
		u.deleteFile(ctx, name, "rollback new photo")
		return 0, "", err
	}
	if oldName != nil {
		u.deleteFile(ctx, *oldName, "remove replaced photo")
	}
	return newVersion, name, nil
}

func (u *usecase) deleteFile(ctx context.Context, name, reason string) {
	// Use a fresh context: the request may already be cancelled.
	if err := u.files.Delete(context.WithoutCancel(ctx), name); err != nil {
		u.log.WarnContext(ctx, "could not delete photo file", "file", name, "reason", reason, "error", err)
	}
}

// UpdateName is the only way to change a name (employees have no route for it).
// Each changed part is logged separately with the admin as changed_by.
func (u *usecase) UpdateName(ctx context.Context, in UpdateNameInput) (int32, error) {
	if err := entity.RequireAdmin(in.Actor); err != nil {
		return 0, err
	}
	var newVersion int32
	err := u.repo.WithTx(ctx, func(tx TxRepository) error {
		emp, err := tx.GetForUpdate(ctx, in.EmployeeID)
		if err != nil {
			return err
		}
		changes, err := emp.ChangeName(in.FirstName, in.LastName, in.Version)
		if err != nil {
			return err
		}
		if newVersion, err = tx.SaveName(ctx, emp); err != nil {
			return err
		}
		for _, c := range changes {
			if err := tx.AppendChange(ctx, emp.ID, c, in.Actor); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return newVersion, nil
}
