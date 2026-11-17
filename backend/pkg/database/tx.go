package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"employee-portal/pkg/database/sqlc"
)

// WithTx runs fn in a transaction. It commits when fn returns nil and rolls back
// otherwise (including on panic, via the deferred Rollback).
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(q *sqlc.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := fn(sqlc.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
