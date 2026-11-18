package profile

import (
	"context"
	"io"

	historydomain "employee-portal/modules/history/domain"
	profiledomain "employee-portal/modules/profile/domain"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

// The ports this module defines: what the handler calls (Usecase) and what the
// use case needs (Repository and friends). Implementations live in usecase.go
// and repository.go; tests use fakes.

// Repository is implemented in repository.go. The use case owns the interface,
// so it never imports GORM or pgx.
type Repository interface {
	GetProfile(ctx context.Context, employeeID int64) (profiledomain.Profile, error)
	// WithTx runs fn in one transaction: commit if fn returns nil, otherwise roll back.
	WithTx(ctx context.Context, fn func(tx TxRepository) error) error
}

// TxRepository is only available inside WithTx.
type TxRepository interface {
	GetForUpdate(ctx context.Context, employeeID int64) (profiledomain.Employee, error)
	SaveContact(ctx context.Context, e profiledomain.Employee, f profiledomain.ContactField) (newVersion int32, err error)
	AppendChange(ctx context.Context, employeeID int64, c historydomain.Change, by shared.Actor) error

	ListEmployeeSkills(ctx context.Context, employeeID int64) ([]skilldomain.Skill, error)
	// GetSkillsByIDs returns the matching skills, active or not, in display order.
	GetSkillsByIDs(ctx context.Context, ids []int64) ([]skilldomain.Skill, error)
	ReplaceSkills(ctx context.Context, employeeID int64, skillIDs []int64) error
	// BumpVersion increments the version for edits stored outside the employees row.
	BumpVersion(ctx context.Context, e profiledomain.Employee) (newVersion int32, err error)

	SavePhoto(ctx context.Context, e profiledomain.Employee) (newVersion int32, err error)
	SaveName(ctx context.Context, e profiledomain.Employee) (newVersion int32, err error)
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

type Usecase interface {
	GetProfile(ctx context.Context, actor shared.Actor) (profiledomain.Profile, error)
	UpdateContact(ctx context.Context, in UpdateContactInput) (newVersion int32, err error)
	SetSkills(ctx context.Context, in SetSkillsInput) (newVersion int32, err error)
	UpdatePhoto(ctx context.Context, in UpdatePhotoInput) (newVersion int32, photoName string, err error)
	UpdateName(ctx context.Context, in UpdateNameInput) (newVersion int32, err error)
}
