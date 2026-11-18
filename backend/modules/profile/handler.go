package profile

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	profiledomain "employee-portal/modules/profile/domain"
	"employee-portal/modules/shared"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

type Handler struct {
	uc Usecase
}

func NewHandler(uc Usecase) *Handler {
	return &Handler{uc: uc}
}

type contactDTO struct {
	MobilePhone *string `json:"mobilePhone"`
	Email       *string `json:"email"`
	Address     *string `json:"address"`
}

type skillDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type profileResponse struct {
	ID           int64      `json:"id"`
	EmployeeCode string     `json:"employeeCode"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	PhotoURL     *string    `json:"photoUrl"`
	Contact      contactDTO `json:"contact"`
	Skills       []skillDTO `json:"skills"`
	Version      int32      `json:"version"`
}

type updateContactRequest struct {
	Value *string `json:"value" binding:"required"`
}

type setSkillsRequest struct {
	SkillIDs []int64 `json:"skillIds" binding:"required"` // [] clears all skills; missing is an error
}

type versionResponse struct {
	Version int32 `json:"version"`
}

// GetProfile: GET /profile
func (h *Handler) GetProfile(c *gin.Context) {
	p, err := h.uc.GetProfile(c.Request.Context(), middleware.ActorFrom(c))
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Header("ETag", httputil.ETag(p.Version))
	c.JSON(http.StatusOK, toProfileResponse(p))
}

// UpdateContact: PUT /profile/contact/:field  (field = mobile | email | address)
func (h *Handler) UpdateContact(c *gin.Context) {
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

	v, err := h.uc.UpdateContact(c.Request.Context(), UpdateContactInput{
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

type photoResponse struct {
	Version  int32  `json:"version"`
	PhotoURL string `json:"photoUrl"`
}

// UpdatePhoto: PUT /profile/photo  multipart field "file". The route sets the body limit.
func (h *Handler) UpdatePhoto(c *gin.Context) {
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

	v, name, err := h.uc.UpdatePhoto(c.Request.Context(), UpdatePhotoInput{
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
func (h *Handler) SetSkills(c *gin.Context) {
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

	v, err := h.uc.SetSkills(c.Request.Context(), SetSkillsInput{
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

type updateNameRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// AdminUpdateName: PUT /admin/employees/:id/name {firstName, lastName} + If-Match
func (h *Handler) AdminUpdateName(c *gin.Context) {
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
	v, err := h.uc.UpdateName(c.Request.Context(), UpdateNameInput{
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

func toProfileResponse(p profiledomain.Profile) profileResponse {
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
