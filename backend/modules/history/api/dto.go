package api

import (
	"time"

	"employee-portal/modules/shared"
)

// Request and response bodies (JSON).

type changedByDTO struct {
	ID   int64       `json:"id"`
	Role shared.Role `json:"role"`
}

type itemDTO struct {
	ID        int64        `json:"id"`
	Field     string       `json:"field"`
	OldValue  any          `json:"oldValue"`
	NewValue  any          `json:"newValue"`
	ChangedAt time.Time    `json:"changedAt"`
	ChangedBy changedByDTO `json:"changedBy"`
}

type pageDTO struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasNext  bool `json:"hasNext"`
}

type listResponse struct {
	Days     *int      `json:"days"` // null = no window (admin view)
	HasOlder bool      `json:"hasOlder"`
	Items    []itemDTO `json:"items"`
	Page     pageDTO   `json:"page"`
}
