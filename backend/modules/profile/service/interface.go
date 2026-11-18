package service

import (
	"context"
	"io"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/profile/domain"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

// Repository is the port this service needs; repo.Repository implements it.
// The service owns the interface, so it never imports GORM or pgx.
type Repository interface {
	GetProfile(ctx context.Context, employeeID int64) (domain.Profile, error)
	// WithTx runs fn in one transaction: commit if fn returns nil, otherwise roll back.
	WithTx(ctx context.Context, fn func(tx TxRepository) error) error
}

// TxRepository is only available inside WithTx.
type TxRepository interface {
	GetForUpdate(ctx context.Context, employeeID int64) (domain.Employee, error)
	SaveContact(ctx context.Context, e domain.Employee, f domain.ContactField) (newVersion int32, err error)
	AppendChange(ctx context.Context, employeeID int64, c historydomain.Change, by shared.Actor) error

	ListEmployeeSkills(ctx context.Context, employeeID int64) ([]skilldomain.Skill, error)
	// GetSkillsByIDs returns the matching skills, active or not, in display order.
	GetSkillsByIDs(ctx context.Context, ids []int64) ([]skilldomain.Skill, error)
	ReplaceSkills(ctx context.Context, employeeID int64, skillIDs []int64) error
	// BumpVersion increments the version for edits stored outside the employees row.
	BumpVersion(ctx context.Context, e domain.Employee) (newVersion int32, err error)

	SavePhoto(ctx context.Context, e domain.Employee) (newVersion int32, err error)
	SaveName(ctx context.Context, e domain.Employee) (newVersion int32, err error)
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
