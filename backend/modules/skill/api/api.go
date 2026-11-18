package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
	"employee-portal/modules/skill/domain"
	"employee-portal/modules/skill/service"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

// API is the HTTP adapter for the skill catalog. Routes are registered in modules/server.
type API struct {
	svc Service
}

func New(svc Service) *API {
	return &API{svc: svc}
}

// List: GET /skills → active skills in display order.
func (h *API) List(c *gin.Context) {
	skills, err := h.svc.ListActive(c.Request.Context())
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	res := make([]skillResponse, len(skills))
	for i, s := range skills {
		res[i] = skillResponse{ID: s.ID, Name: s.Name}
	}
	c.JSON(http.StatusOK, res)
}

// AdminList: GET /admin/skills → every skill with usage counts.
func (h *API) AdminList(c *gin.Context) {
	skills, err := h.svc.AdminList(c.Request.Context(), middleware.ActorFrom(c))
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	res := make([]adminSkillResponse, len(skills))
	for i, s := range skills {
		res[i] = adminSkillResponse{ID: s.ID, Name: s.Name, IsActive: s.IsActive, SortOrder: s.SortOrder, UsedBy: s.UsedBy}
	}
	c.JSON(http.StatusOK, res)
}

// Create: POST /admin/skills {name} → 201 {id}; 409 if the name exists.
func (h *API) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("name", "is required"))
		return
	}
	id, err := h.svc.Create(c.Request.Context(), middleware.ActorFrom(c), req.Name)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// Update: PATCH /admin/skills/:id {name?, isActive?} → 204.
func (h *API) Update(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("body", "must be JSON with name and/or isActive"))
		return
	}
	err := h.svc.Update(c.Request.Context(), middleware.ActorFrom(c), service.UpdateInput{ID: id, Name: req.Name, IsActive: req.IsActive})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Reorder: PUT /admin/skills/order [{id, sortOrder}] → 204.
func (h *API) Reorder(c *gin.Context) {
	var req []orderItem
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("order", "must be an array of {id, sortOrder}"))
		return
	}
	items := make([]domain.SkillOrder, len(req))
	for i, it := range req {
		if it.SortOrder == nil { // do not rely on validator support for slice elements
			httputil.WriteError(c, shared.NewValidationError("order", "every item needs sortOrder"))
			return
		}
		items[i] = domain.SkillOrder{ID: it.ID, SortOrder: *it.SortOrder}
	}
	if err := h.svc.Reorder(c.Request.Context(), middleware.ActorFrom(c), items); err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
