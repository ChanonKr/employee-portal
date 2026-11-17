package setting

import "employee-portal/pkg/router"

func (h *Handler) RegisterRoutes(r router.Groups) {
	r.Admin.GET("/settings", h.Get)
	r.Admin.PUT("/settings/history-days", h.SetHistoryDays)
}
