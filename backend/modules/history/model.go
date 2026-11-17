package history

import "time"

// Read-only views over profile_change_logs and app_settings
// (see profile/model.go for the rules).

const (
	changeLogTable = "profile_change_logs"
	settingTable   = "app_settings"
	employeeTable  = "employees"
)

// changeLogRow is one history entry with the role of whoever made the change.
type changeLogRow struct {
	ID            int64
	Field         string
	OldValue      []byte
	NewValue      []byte
	ChangedAt     time.Time
	ChangedBy     int64
	ChangedByRole string
}
