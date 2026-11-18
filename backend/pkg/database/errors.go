package database

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"employee-portal/modules/shared"
)

// PostgreSQL error codes, see https://www.postgresql.org/docs/17/errcodes-appendix.html
const (
	codeUniqueViolation     = "23505"
	codeCheckViolation      = "23514"
	codeForeignKeyViolation = "23503"
)

// MapError turns driver errors into domain errors so use cases never see GORM or pgx types.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case codeUniqueViolation:
			return fmt.Errorf("%w: %s", shared.ErrDuplicate, pgErr.ConstraintName)
		case codeCheckViolation:
			return shared.NewValidationError(pgErr.ConstraintName, "violates a data rule")
		case codeForeignKeyViolation:
			return shared.NewValidationError(pgErr.ConstraintName, "references a missing record")
		}
	}
	return fmt.Errorf("database: %w", err)
}
