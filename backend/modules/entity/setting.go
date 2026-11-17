package entity

import "time"

// HistorySettings is the admin view of the history window.
type HistorySettings struct {
	Days          int
	UpdatedByID   *int64 // nil: never changed in the UI (seed or env fallback)
	UpdatedByCode *string
	UpdatedAt     *time.Time
}

// ValidateHistoryDays checks a value entered by an admin.
func ValidateHistoryDays(days int) error {
	if days < MinHistoryDays || days > MaxHistoryDays {
		return NewValidationError("value", "must be from 1 to 365 days")
	}
	return nil
}

// RequireAdmin is the use-case-level role check (the router checks it too).
func RequireAdmin(a Actor) error {
	if !a.IsAdmin() {
		return ErrForbidden
	}
	return nil
}
