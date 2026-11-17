package history

import (
	"employee-portal/pkg/middleware"
	"employee-portal/pkg/router"
)

// RegisterRoutes mounts the read-only history routes. Writes happen inside the
// profile transactions.
func (h *Handler) RegisterRoutes(r router.Groups) {
	r.API.GET("/profile/history", r.SignedIn, r.Ready, middleware.RequireProfile(), h.ListOwn)
	r.Admin.GET("/employees/:id/history", h.AdminList)
}
