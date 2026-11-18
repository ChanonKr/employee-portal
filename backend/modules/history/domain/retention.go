package domain

import "time"

// Retention rules, applied by the retention job (history/service.Retention):
//
//   - changes older than the retention (HISTORY_RETENTION_DAYS) are deleted;
//   - except photo changes: per employee, the KeepPhotosBeyondRetention newest
//     photo changes older than the retention are kept, so the last few photos
//     before the window can still be seen in the history;
//   - a photo file is deleted once no remaining photo change has it as its new
//     value and no employee uses it as the current photo.
//
// The file a kept change replaced (its old value) may therefore be gone; the UI
// shows a placeholder for it.

// KeepPhotosBeyondRetention is how many photos older than the retention are kept per employee.
const KeepPhotosBeyondRetention = 3

// RetentionCutoff is the moment before which changes are purged.
func RetentionCutoff(now time.Time, retentionDays int) time.Time {
	return now.AddDate(0, 0, -retentionDays)
}

// PurgeResult is what one retention run did.
type PurgeResult struct {
	Ran            bool     // false when another instance was already purging
	DeletedChanges int64    // change log rows deleted
	UnusedPhotos   []string // photo files nothing refers to any more (to delete after commit)
}
