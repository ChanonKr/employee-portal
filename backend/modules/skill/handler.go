package skill

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

type Handler struct {
	uc Usecase
}

func NewHandler(uc Usecase) *Handler {
	return &Handler{uc: uc}
}

type skillResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type adminSkillResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"isActive"`
	SortOrder int32  `json:"sortOrder"`
	UsedBy    int64  `json:"usedBy"`
}

type createRequest struct {
	Name string `json:"name" binding:"required"`
}

type updateRequest struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"isActive"`
}

type orderItem struct {
	ID        int64  `json:"id" binding:"required"`
	SortOrder *int32 `json:"sortOrder" binding:"required"`
}

// List: GET /skills → active skills in display order.
func (h *Handler) List(c *gin.Context) {
	skills, err := h.uc.ListActive(c.Request.Context())
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
func (h *Handler) AdminList(c *gin.Context) {
	skills, err := h.uc.AdminList(c.Request.Context(), middleware.ActorFrom(c))
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
func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("name", "is required"))
		return
	}
	id, err := h.uc.Create(c.Request.Context(), middleware.ActorFrom(c), req.Name)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// Update: PATCH /admin/skills/:id {name?, isActive?} → 204.
func (h *Handler) Update(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("body", "must be JSON with name and/or isActive"))
		return
	}
	err := h.uc.Update(c.Request.Context(), middleware.ActorFrom(c), UpdateInput{ID: id, Name: req.Name, IsActive: req.IsActive})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Reorder: PUT /admin/skills/order [{id, sortOrder}] → 204.
func (h *Handler) Reorder(c *gin.Context) {
	var req []orderItem
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, shared.NewValidationError("order", "must be an array of {id, sortOrder}"))
		return
	}
	items := make([]skilldomain.SkillOrder, len(req))
	for i, it := range req {
		if it.SortOrder == nil { // do not rely on validator support for slice elements
			httputil.WriteError(c, shared.NewValidationError("order", "every item needs sortOrder"))
			return
		}
		items[i] = skilldomain.SkillOrder{ID: it.ID, SortOrder: *it.SortOrder}
	}
	if err := h.uc.Reorder(c.Request.Context(), middleware.ActorFrom(c), items); err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
