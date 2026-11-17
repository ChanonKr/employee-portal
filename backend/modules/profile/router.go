package profile

import (
	"time"

	"employee-portal/pkg/middleware"
	"employee-portal/pkg/router"
)

// multipartOverhead leaves room for multipart headers around the file itself.
const multipartOverhead = 64 << 10

// RegisterRoutes mounts /profile/* (the employee id always comes from the token;
// the URL has no id) and the admin rename.
func (h *Handler) RegisterRoutes(r router.Groups, uploadMaxBytes int64) {
	p := r.API.Group("/profile", r.SignedIn, r.Ready, middleware.RequireProfile())
	p.GET("", h.GetProfile)
	p.PUT("/contact/:field", h.UpdateContact)
	p.PUT("/skills", h.SetSkills)
	// No route changes the name: F2 is enforced by the server, not only the UI.

	// The upload has its own group: a body limit sized for the file, not for JSON.
	uploadLimiter := middleware.NewKeyedLimiter(10, time.Minute)
	upload := r.Root.Group("/api/v1/profile",
		middleware.BodyLimit(uploadMaxBytes+multipartOverhead),
		r.SignedIn, r.Ready, middleware.RequireProfile(),
		middleware.RateLimit(uploadLimiter, middleware.ByUser))
	upload.PUT("/photo", h.UpdatePhoto)

	r.Admin.PUT("/employees/:id/name", h.AdminUpdateName)
}
