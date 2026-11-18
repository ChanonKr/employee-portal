package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/employee/service"
	"employee-portal/pkg/httputil"
	"employee-portal/pkg/middleware"
)

// API is the HTTP adapter for the admin directory. Routes are registered in modules/server.
type API struct {
	svc Service
}

func New(svc Service) *API {
	return &API{svc: svc}
}

// List: GET /admin/employees?q=&page=&pageSize=
func (h *API) List(c *gin.Context) {
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
	res, err := h.svc.List(c.Request.Context(), middleware.ActorFrom(c), service.ListInput{
		Search: c.Query("q"), Page: page, PageSize: pageSize,
	})
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	out := listResponse{Items: make([]summaryDTO, len(res.Items)), Total: res.Total}
	for i, e := range res.Items {
		out.Items[i] = summaryDTO{
			ID: e.ID, EmployeeCode: e.EmployeeCode, FirstName: e.FirstName, LastName: e.LastName,
			PhotoURL: httputil.OptionalPhotoURL(e.PhotoPath), Role: e.Role, IsActive: e.IsActive,
		}
	}
	out.Page.Page, out.Page.PageSize = res.Page.Page, res.Page.PageSize
	out.Page.HasNext = int64(res.Page.Offset()+len(res.Items)) < res.Total
	c.JSON(http.StatusOK, out)
}

// Get: GET /admin/employees/:id (ETag = version, for the name edit's If-Match)
func (h *API) Get(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	d, err := h.svc.Get(c.Request.Context(), middleware.ActorFrom(c), id)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	res := detailResponse{
		ID: d.ID, EmployeeCode: d.EmployeeCode, FirstName: d.FirstName, LastName: d.LastName,
		PhotoURL: httputil.OptionalPhotoURL(d.PhotoPath),
		Contact:  contactDTO{MobilePhone: d.MobilePhone, Email: d.ContactEmail, Address: d.Address},
		Skills:   make([]skillDTO, len(d.Skills)),
		Version:  d.Version, Role: d.Role, IsActive: d.IsActive,
	}
	for i, s := range d.Skills {
		res.Skills[i] = skillDTO{ID: s.ID, Name: s.Name, IsActive: s.IsActive}
	}
	c.Header("ETag", httputil.ETag(d.Version))
	c.JSON(http.StatusOK, res)
}

// Stats: GET /admin/stats
func (h *API) Stats(c *gin.Context) {
	s, err := h.svc.Stats(c.Request.Context(), middleware.ActorFrom(c))
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	res := statsResponse{Employees: s.Employees, CompleteProfiles: s.CompleteProfiles}
	res.TopSkills = make([]topSkillDTO, len(s.TopSkills))
	for i, t := range s.TopSkills {
		res.TopSkills[i] = topSkillDTO{ID: t.ID, Name: t.Name, UsedBy: t.UsedBy}
	}
	res.RecentChanges = make([]recentDTO, len(s.RecentChanges))
	for i, r := range s.RecentChanges {
		res.RecentChanges[i] = recentDTO{
			ID: r.ID, EmployeeID: r.EmployeeID, EmployeeCode: r.EmployeeCode,
			Name: r.FirstName + " " + r.LastName, Field: string(r.Field),
			ChangedAt: r.ChangedAt.UTC(), ChangedByRole: r.ChangedByRole,
		}
	}
	c.JSON(http.StatusOK, res)
}
