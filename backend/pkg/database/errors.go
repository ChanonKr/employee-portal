package database

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"employee-portal/modules/entity"
)

// PostgreSQL error codes, see https://www.postgresql.org/docs/17/errcodes-appendix.html
const (
	codeUniqueViolation     = "23505"
	codeCheckViolation      = "23514"
	codeForeignKeyViolation = "23503"
)

// MapError turns driver errors into domain errors so use cases never see pgx types.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case codeUniqueViolation:
			return fmt.Errorf("%w: %s", entity.ErrDuplicate, pgErr.ConstraintName)
		case codeCheckViolation:
			return entity.NewValidationError(pgErr.ConstraintName, "violates a data rule")
		case codeForeignKeyViolation:
			return entity.NewValidationError(pgErr.ConstraintName, "references a missing record")
		}
	}
	return fmt.Errorf("database: %w", err)
}
