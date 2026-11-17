package file

import "employee-portal/pkg/router"

func (h *Handler) RegisterRoutes(r router.Groups) {
	r.API.GET("/files/photos/:name", r.SignedIn, r.Ready, h.Photo)
}
