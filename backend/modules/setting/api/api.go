package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

// API is the HTTP adapter for settings. Routes are registered in modules/server.
type API struct {
	svc Service
}

func New(svc Service) *API {
	return &API{svc: svc}
}

// Get: GET /admin/settings
func (h *API) Get(c *gin.Context) {
	s, err := h.svc.GetHistory(c.Request.Context(), middleware.ActorFrom(c))
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
func (h *API) SetHistoryDays(c *gin.Context) {
	var req historyDaysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("value", "must be a whole number from 1 to 365"))
		return
	}
	if err := h.svc.SetHistoryDays(c.Request.Context(), middleware.ActorFrom(c), *req.Value); err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
