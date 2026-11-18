// Package model holds the read-only row shapes for profile_change_logs and
// app_settings (see modules/profile/model for the rules); only repo uses them.
package model

import "time"

const (
	ChangeLogTable = "profile_change_logs"
	SettingTable   = "app_settings"
	EmployeeTable  = "employees"
)

// ChangeLogRow is one history entry with the role of whoever made the change.
type ChangeLogRow struct {
	ID            int64
	Field         string
	OldValue      []byte
	NewValue      []byte
	ChangedAt     time.Time
	ChangedBy     int64
	ChangedByRole string
}
