package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/profile/domain"
	"employee-portal/modules/profile/service"
	"employee-portal/modules/shared"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

// API is the HTTP adapter for profile. Routes are registered in modules/server.
type API struct {
	svc Service
}

func New(svc Service) *API {
	return &API{svc: svc}
}

// GetProfile: GET /profile
func (h *API) GetProfile(c *gin.Context) {
	p, err := h.svc.GetProfile(c.Request.Context(), middleware.ActorFrom(c))
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Header("ETag", httputil.ETag(p.Version))
	c.JSON(http.StatusOK, toProfileResponse(p))
}

// UpdateContact: PUT /profile/contact/:field  (field = mobile | email | address)
func (h *API) UpdateContact(c *gin.Context) {
	version, err := httputil.IfMatchVersion(c)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	var req updateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("value", "is required"))
		return
	}

	v, err := h.svc.UpdateContact(c.Request.Context(), service.UpdateContactInput{
		Actor:   middleware.ActorFrom(c),
		Field:   c.Param("field"),
		Value:   *req.Value,
		Version: version,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Header("ETag", httputil.ETag(v))
	c.JSON(http.StatusOK, versionResponse{Version: v})
}

// UpdatePhoto: PUT /profile/photo  multipart field "file". The route sets the body limit.
func (h *API) UpdatePhoto(c *gin.Context) {
	version, err := httputil.IfMatchVersion(c)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			httputil.WriteError(c, shared.ErrPhotoTooLarge)
			return
		}
		httputil.WriteError(c, shared.NewValidationError("file", "is required"))
		return
	}
	defer file.Close()

	v, name, err := h.svc.UpdatePhoto(c.Request.Context(), service.UpdatePhotoInput{
		Actor:   middleware.ActorFrom(c),
		File:    file,
		Version: version,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Header("ETag", httputil.ETag(v))
	c.JSON(http.StatusOK, photoResponse{Version: v, PhotoURL: httputil.PhotoURL(name)})
}

// SetSkills: PUT /profile/skills  body {skillIds: number[]}
func (h *API) SetSkills(c *gin.Context) {
	version, err := httputil.IfMatchVersion(c)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	var req setSkillsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("skillIds", "must be an array of skill ids"))
		return
	}

	v, err := h.svc.SetSkills(c.Request.Context(), service.SetSkillsInput{
		Actor:    middleware.ActorFrom(c),
		SkillIDs: req.SkillIDs,
		Version:  version,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Header("ETag", httputil.ETag(v))
	c.JSON(http.StatusOK, versionResponse{Version: v})
}

// AdminUpdateName: PUT /admin/employees/:id/name {firstName, lastName} + If-Match
func (h *API) AdminUpdateName(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	version, err := httputil.IfMatchVersion(c)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	var req updateNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("body", "must be JSON with firstName and lastName"))
		return
	}
	v, err := h.svc.UpdateName(c.Request.Context(), service.UpdateNameInput{
		Actor: middleware.ActorFrom(c), EmployeeID: id,
		FirstName: req.FirstName, LastName: req.LastName, Version: version,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Header("ETag", httputil.ETag(v))
	c.JSON(http.StatusOK, versionResponse{Version: v})
}

func toProfileResponse(p domain.Profile) profileResponse {
	res := profileResponse{
		ID:           p.ID,
		EmployeeCode: p.EmployeeCode,
		FirstName:    p.FirstName,
		LastName:     p.LastName,
		Contact:      contactDTO{MobilePhone: p.MobilePhone, Email: p.ContactEmail, Address: p.Address},
		Skills:       make([]skillDTO, len(p.Skills)),
		Version:      p.Version,
	}
	res.PhotoURL = httputil.OptionalPhotoURL(p.PhotoPath)
	for i, s := range p.Skills {
		res.Skills[i] = skillDTO{ID: s.ID, Name: s.Name}
	}
	return res
}
