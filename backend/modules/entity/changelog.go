package entity

import (
	"strconv"
	"strings"
	"time"
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
		return "", NewValidationError("field", "unknown field")
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
	ChangedBy Actor // UserID and Role only
	ChangedAt time.Time
}

// ── history window ("7 days" is how far back we show, not how long we keep)

const (
	MinHistoryDays = 1
	MaxHistoryDays = 365
)

// ParseHistoryDays validates the history_days setting.
func ParseHistoryDays(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < MinHistoryDays || n > MaxHistoryDays {
		return 0, NewValidationError("historyDays", "must be a whole number from 1 to 365")
	}
	return n, nil
}

// ── paging

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PageRequest struct {
	Page     int // 1-based
	PageSize int
}

// NewPageRequest applies defaults for zero values and rejects out-of-range input.
func NewPageRequest(page, pageSize int) (PageRequest, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if page < 1 {
		return PageRequest{}, NewValidationError("page", "must be 1 or more")
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		return PageRequest{}, NewValidationError("pageSize", "must be from 1 to 100")
	}
	return PageRequest{Page: page, PageSize: pageSize}, nil
}

func (p PageRequest) Offset() int { return (p.Page - 1) * p.PageSize }

// HistoryQuery selects change logs for one employee, newest first.
type HistoryQuery struct {
	EmployeeID int64
	Days       int        // 0 = no time window (admin view)
	Field      *FieldName // nil = all fields
	PageRequest
}
