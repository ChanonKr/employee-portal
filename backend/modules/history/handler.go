package history

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/entity"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

type Handler struct {
	uc Usecase
}

func NewHandler(uc Usecase) *Handler {
	return &Handler{uc: uc}
}

type changedByDTO struct {
	ID   int64       `json:"id"`
	Role entity.Role `json:"role"`
}

type itemDTO struct {
	ID        int64        `json:"id"`
	Field     string       `json:"field"`
	OldValue  any          `json:"oldValue"`
	NewValue  any          `json:"newValue"`
	ChangedAt time.Time    `json:"changedAt"`
	ChangedBy changedByDTO `json:"changedBy"`
}

type pageDTO struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasNext  bool `json:"hasNext"`
}

type listResponse struct {
	Days     int       `json:"days"`
	HasOlder bool      `json:"hasOlder"`
	Items    []itemDTO `json:"items"`
	Page     pageDTO   `json:"page"`
}

// ListOwn: GET /profile/history?field=&page=&pageSize=
func (h *Handler) ListOwn(c *gin.Context) {
	page, err := queryInt(c, "page")
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	pageSize, err := queryInt(c, "pageSize")
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	res, err := h.uc.ListOwn(c.Request.Context(), ListInput{
		Actor:    middleware.ActorFrom(c),
		Field:    c.Query("field"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, toListResponse(res))
}

// queryInt returns 0 when the parameter is absent so the use case applies its default.
func queryInt(c *gin.Context, name string) (int, error) {
	s := c.Query(name)
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, entity.NewValidationError(name, "must be a whole number")
	}
	return n, nil
}

func toListResponse(r Result) listResponse {
	items := make([]itemDTO, len(r.Items))
	for i, l := range r.Items {
		items[i] = itemDTO{
			ID:        l.ID,
			Field:     string(l.Field),
			OldValue:  l.Old,
			NewValue:  l.New,
			ChangedAt: l.ChangedAt.UTC(),
			ChangedBy: changedByDTO{ID: l.ChangedBy.UserID, Role: l.ChangedBy.Role},
		}
	}
	return listResponse{
		Days:     r.Days,
		HasOlder: r.HasOlder,
		Items:    items,
		Page:     pageDTO{Page: r.Page.Page, PageSize: r.Page.PageSize, HasNext: r.HasNext},
	}
}
