package server

import (
	"context"
	"time"

	"gorm.io/gorm"

	historyrepo "employee-portal/modules/history/repo"
	historyservice "employee-portal/modules/history/service"
)

// RunRetention deletes history older than HISTORY_RETENTION_DAYS (and the photo
// files nothing uses any more) right away and then every RETENTION_INTERVAL,
// until ctx is cancelled. db must connect as the "retention" role. Several API
// instances may run it: a database lock lets only one purge at a time.
func (s *Server) RunRetention(ctx context.Context, db *gorm.DB) {
	job := historyservice.NewRetention(historyrepo.NewRetention(db), s.files, s.cfg.HistoryRetentionDays, s.log)
	run := func() {
		start := time.Now()
		res, err := job.Run(ctx)
		switch {
		case err != nil:
			s.log.ErrorContext(ctx, "retention: purge failed", "error", err)
		case !res.Ran:
			s.log.InfoContext(ctx, "retention: another instance is purging, skipped")
		default:
			s.log.InfoContext(ctx, "retention: purged",
				"retentionDays", s.cfg.HistoryRetentionDays,
				"deletedChanges", res.DeletedChanges,
				"deletedPhotos", len(res.UnusedPhotos),
				"took", time.Since(start).String())
		}
	}

	run()
	t := time.NewTicker(s.cfg.RetentionInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			run()
		}
	}
}
