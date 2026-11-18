// Package domain holds the app settings rules. No framework imports.
package domain

import (
	"time"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
)

// HistorySettings is the admin view of the history window.
type HistorySettings struct {
	Days          int
	UpdatedByID   *int64 // nil: never changed in the UI (seed or env fallback)
	UpdatedByCode *string
	UpdatedAt     *time.Time
}

// ValidateHistoryDays checks a value entered by an admin.
func ValidateHistoryDays(days int) error {
	if days < historydomain.MinHistoryDays || days > historydomain.MaxHistoryDays {
		return shared.NewValidationError("value", "must be from 1 to 365 days")
	}
	return nil
}
