package database

import (
	"context"

	"gorm.io/gorm"
)

// WithTx runs fn in a transaction. It commits when fn returns nil and rolls back
// otherwise (including on panic).
func WithTx(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}
