package service

import (
	"context"
	"log/slog"
	"time"

	"employee-portal/modules/history/domain"
)

// RetentionRepository deletes old history. It runs as a separate database role:
// the API role may only append to the change log.
type RetentionRepository interface {
	// Purge deletes changes older than cutoff, keeping the keepPhotos newest photo
	// changes per employee among them, in one transaction. It reports Ran=false and
	// does nothing when another instance is already purging.
	Purge(ctx context.Context, cutoff time.Time, keepPhotos int) (domain.PurgeResult, error)
}

// FileDeleter removes stored photo files (pkg/storage). A missing file is not an error.
type FileDeleter interface {
	Delete(ctx context.Context, name string) error
}

// Retention enforces HISTORY_RETENTION_DAYS (see domain/retention.go for the rules).
type Retention struct {
	repo  RetentionRepository
	files FileDeleter
	days  int
	log   *slog.Logger
	now   func() time.Time
}

func NewRetention(repo RetentionRepository, files FileDeleter, retentionDays int, log *slog.Logger) *Retention {
	if log == nil {
		log = slog.Default()
	}
	return &Retention{repo: repo, files: files, days: retentionDays, log: log, now: time.Now}
}

// Run purges once: rows first (one transaction), then the photo files nothing uses
// any more. A file that cannot be deleted is logged and left behind; it is harmless.
func (r *Retention) Run(ctx context.Context) (domain.PurgeResult, error) {
	res, err := r.repo.Purge(ctx, domain.RetentionCutoff(r.now(), r.days), domain.KeepPhotosBeyondRetention)
	if err != nil || !res.Ran {
		return res, err
	}
	for _, name := range res.UnusedPhotos {
		if err := r.files.Delete(ctx, name); err != nil {
			r.log.WarnContext(ctx, "retention: could not delete photo file", "file", name, "error", err)
		}
	}
	return res, nil
}
