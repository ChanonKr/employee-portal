// Package domain holds the app settings rules. No framework imports.
package domain

import (
	"fmt"
	"time"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
)

// HistorySettings is the admin view of the history window.
type HistorySettings struct {
	Days          int    // the window in effect
	MaxDays       int    // retention (HISTORY_RETENTION_DAYS): the window cannot be longer
	UpdatedByID   *int64 // nil: never changed in the UI (seed or default)
	UpdatedByCode *string
	UpdatedAt     *time.Time
}

// ValidateHistoryDays checks a window entered by an admin: from 1 to the retention.
func ValidateHistoryDays(days, retentionDays int) error {
	if days < historydomain.MinHistoryDays || days > retentionDays {
		return shared.NewValidationError("value", fmt.Sprintf("must be from 1 to %d days", retentionDays))
	}
	return nil
}
