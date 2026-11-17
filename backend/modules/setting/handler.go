package setting

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/entity"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

type Handler struct {
	uc Usecase
}

func NewHandler(uc Usecase) *Handler {
	return &Handler{uc: uc}
}

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

// Get: GET /admin/settings
func (h *Handler) Get(c *gin.Context) {
	s, err := h.uc.GetHistory(c.Request.Context(), middleware.ActorFrom(c))
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	res := settingsResponse{HistoryDays: s.Days, UpdatedAt: s.UpdatedAt}
	if s.UpdatedByID != nil && s.UpdatedByCode != nil {
		res.UpdatedBy = &updatedByDTO{ID: *s.UpdatedByID, EmployeeCode: *s.UpdatedByCode}
	}
	c.JSON(http.StatusOK, res)
}

// SetHistoryDays: PUT /admin/settings/history-days {value} (1–365) → 204
func (h *Handler) SetHistoryDays(c *gin.Context) {
	var req historyDaysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, entity.NewValidationError("value", "must be a whole number from 1 to 365"))
		return
	}
	if err := h.uc.SetHistoryDays(c.Request.Context(), middleware.ActorFrom(c), *req.Value); err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
