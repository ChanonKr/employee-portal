package employee

import "employee-portal/pkg/router"

// RegisterRoutes mounts the admin directory and dashboard.
func (h *Handler) RegisterRoutes(r router.Groups) {
	r.Admin.GET("/stats", h.Stats)
	r.Admin.GET("/employees", h.List)
	r.Admin.GET("/employees/:id", h.Get)
}
