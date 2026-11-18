// Package domain holds the change history rules: which fields are logged, what a
// change looks like and the history window. No framework imports.
package domain

import (
	"strconv"
	"strings"
	"time"

	"employee-portal/modules/shared"
)

// FieldName is the profile field recorded in the change history.
type FieldName string

const (
	FieldFirstName    FieldName = "first_name"
	FieldLastName     FieldName = "last_name"
	FieldMobilePhone  FieldName = "mobile_phone"
	FieldContactEmail FieldName = "contact_email"
	FieldAddress      FieldName = "address"
	FieldPhoto        FieldName = "photo"
	FieldSkills       FieldName = "skills"
)

// Same list as the CHECK constraint on profile_change_logs.field.
var fieldNames = map[FieldName]struct{}{
	FieldFirstName: {}, FieldLastName: {}, FieldMobilePhone: {}, FieldContactEmail: {},
	FieldAddress: {}, FieldPhoto: {}, FieldSkills: {},
}

// ParseFieldName validates a history filter such as ?field=mobile_phone.
func ParseFieldName(s string) (FieldName, error) {
	f := FieldName(s)
	if _, ok := fieldNames[f]; !ok {
		return "", shared.NewValidationError("field", "unknown field")
	}
	return f, nil
}

// Change is one field edit. Old and New are nil, a string, or []string (skills).
type Change struct {
	Field FieldName
	Old   any
	New   any
}

// ChangeLog is a recorded Change.
type ChangeLog struct {
	ID int64
	Change
	ChangedBy shared.Actor // UserID and Role only
	ChangedAt time.Time
}

// ── history window and retention
//
// Two different numbers:
//   - retention (HISTORY_RETENTION_DAYS, default 90): how long changes are kept
//     before the retention job deletes them. Set by operators, not in the UI.
//   - window (app_settings.history_days, default 7): how far back employees see
//     on their profile. Set by admins in the UI; never more than retention.

const (
	DefaultHistoryDays = 7 // window when the admin setting is missing or invalid
	MinHistoryDays     = 1
	MaxRetentionDays   = 3650 // upper bound for HISTORY_RETENTION_DAYS
)

// ParseHistoryDays validates the stored history_days setting (a whole number ≥ 1).
// The upper bound is the retention, applied by HistoryWindow.
func ParseHistoryDays(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < MinHistoryDays {
		return 0, shared.NewValidationError("historyDays", "must be a whole number of at least 1")
	}
	return n, nil
}

// HistoryWindow is the window employees see: the stored setting when it is valid,
// otherwise DefaultHistoryDays, and never more days than are kept.
// raw is "" when the setting does not exist.
func HistoryWindow(raw string, retentionDays int) int {
	days, err := ParseHistoryDays(raw)
	if err != nil {
		days = DefaultHistoryDays
	}
	return min(days, retentionDays)
}

// HistoryQuery selects change logs for one employee, newest first.
type HistoryQuery struct {
	EmployeeID int64
	Days       int        // 0 = no time window (admin view)
	Field      *FieldName // nil = all fields
	shared.PageRequest
}
