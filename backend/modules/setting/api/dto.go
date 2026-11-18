package api

import (
	"time"
)

// Request and response bodies (JSON).

type updatedByDTO struct {
	ID           int64  `json:"id"`
	EmployeeCode string `json:"employeeCode"`
}

type settingsResponse struct {
	HistoryDays int           `json:"historyDays"`
	UpdatedBy   *updatedByDTO `json:"updatedBy"`
	UpdatedAt   *time.Time    `json:"updatedAt"`
}

type historyDaysRequest struct {
	Value *int `json:"value" binding:"required"`
}
