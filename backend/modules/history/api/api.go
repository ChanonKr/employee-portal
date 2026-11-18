package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/history/service"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

// API is the HTTP adapter for change history. Routes are registered in modules/server.
type API struct {
	svc Service
}

func New(svc Service) *API {
	return &API{svc: svc}
}

// ListOwn: GET /profile/history?field=&page=&pageSize=
func (h *API) ListOwn(c *gin.Context) {
	page, err := httputil.QueryInt(c, "page")
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	pageSize, err := httputil.QueryInt(c, "pageSize")
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	res, err := h.svc.ListOwn(c.Request.Context(), service.ListInput{
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

// AdminList: GET /admin/employees/:id/history?field=&page=&pageSize= (all time)
func (h *API) AdminList(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	page, err := httputil.QueryInt(c, "page")
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	pageSize, err := httputil.QueryInt(c, "pageSize")
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	res, err := h.svc.ListForEmployee(c.Request.Context(), id, service.ListInput{
		Actor: middleware.ActorFrom(c), Field: c.Query("field"), Page: page, PageSize: pageSize,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, toListResponse(res))
}

func toListResponse(r service.Result) listResponse {
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
	var days *int
	if !r.AllTime {
		days = &r.Days
	}
	return listResponse{
		Days:     days,
		HasOlder: r.HasOlder,
		Items:    items,
		Page:     pageDTO{Page: r.Page.Page, PageSize: r.Page.PageSize, HasNext: r.HasNext},
	}
}
