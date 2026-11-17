package skill

import "employee-portal/pkg/router"

// RegisterRoutes mounts the picker catalog and the admin skill management.
func (h *Handler) RegisterRoutes(r router.Groups) {
	r.API.GET("/skills", r.SignedIn, r.Ready, h.List)

	r.Admin.GET("/skills", h.AdminList)
	r.Admin.POST("/skills", h.Create)
	r.Admin.PATCH("/skills/:id", h.Update)
	r.Admin.PUT("/skills/order", h.Reorder)
}
