package repository

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	"employee-portal/modules/history/domain"
	"employee-portal/modules/history/service"
	"employee-portal/pkg/database"
)

// RetentionRepository purges old history. Its *gorm.DB must connect as the
// "retention" role (SELECT, DELETE on profile_change_logs; SELECT on employees):
// the API role may only append to the change log.
type RetentionRepository struct {
	db *gorm.DB
}

var _ service.RetentionRepository = (*RetentionRepository)(nil)

func NewRetention(db *gorm.DB) *RetentionRepository {
	return &RetentionRepository{db: db}
}

// retentionLockKey serializes purges when several API instances run the job.
const retentionLockKey = "employee-portal:history-retention"

// purgeSQL deletes every change older than the cutoff except, per employee, the
// newest @keep photo changes among them. It returns the photo file names the
// deleted photo changes referred to (candidates for deletion).
const purgeSQL = `
WITH kept_photos AS (
  SELECT id FROM (
    SELECT id, row_number() OVER (PARTITION BY employee_id ORDER BY changed_at DESC, id DESC) AS rn
    FROM profile_change_logs
    WHERE field = 'photo' AND changed_at < @cutoff
  ) ranked
  WHERE rn <= @keep
)
DELETE FROM profile_change_logs
WHERE changed_at < @cutoff AND id NOT IN (SELECT id FROM kept_photos)
RETURNING field, old_value #>> '{}' AS old_name, new_value #>> '{}' AS new_name`

// inUseSQL keeps a file that is still a current photo or the new value of a
// remaining photo change (see domain/retention.go).
const inUseSQL = `
SELECT photo_path AS name FROM employees WHERE photo_path IN @names
UNION
SELECT new_value #>> '{}' FROM profile_change_logs WHERE field = 'photo' AND new_value #>> '{}' IN @names`

func (r *RetentionRepository) Purge(ctx context.Context, cutoff time.Time, keepPhotos int) (domain.PurgeResult, error) {
	var res domain.PurgeResult
	err := database.WithTx(ctx, r.db, func(tx *gorm.DB) error {
		// Released at commit/rollback; another instance gets false and skips.
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(hashtext(?))", retentionLockKey).Scan(&res.Ran).Error; err != nil {
			return database.MapError(err)
		}
		if !res.Ran {
			return nil
		}

		var deleted []struct {
			Field   string
			OldName sql.NullString
			NewName sql.NullString
		}
		err := tx.Raw(purgeSQL, map[string]any{"cutoff": cutoff, "keep": keepPhotos}).Scan(&deleted).Error
		if err != nil {
			return database.MapError(err)
		}
		res.DeletedChanges = int64(len(deleted))

		seen := map[string]struct{}{}
		var candidates []string
		for _, d := range deleted {
			if d.Field != string(domain.FieldPhoto) {
				continue
			}
			for _, n := range []sql.NullString{d.OldName, d.NewName} {
				if _, dup := seen[n.String]; n.Valid && !dup {
					seen[n.String] = struct{}{}
					candidates = append(candidates, n.String)
				}
			}
		}
		if len(candidates) == 0 {
			return nil
		}

		var inUse []string
		if err := tx.Raw(inUseSQL, map[string]any{"names": candidates}).Scan(&inUse).Error; err != nil {
			return database.MapError(err)
		}
		used := make(map[string]struct{}, len(inUse))
		for _, n := range inUse {
			used[n] = struct{}{}
		}
		for _, n := range candidates {
			if _, ok := used[n]; !ok {
				res.UnusedPhotos = append(res.UnusedPhotos, n)
			}
		}
		return nil
	})
	if err != nil {
		return domain.PurgeResult{}, err
	}
	return res, nil
}
