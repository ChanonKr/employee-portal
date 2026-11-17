package employee

import (
	"net/http"
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

type summaryDTO struct {
	ID           int64       `json:"id"`
	EmployeeCode string      `json:"employeeCode"`
	FirstName    string      `json:"firstName"`
	LastName     string      `json:"lastName"`
	PhotoURL     *string     `json:"photoUrl"`
	Role         entity.Role `json:"role"`
	IsActive     bool        `json:"isActive"`
}

type listResponse struct {
	Items []summaryDTO `json:"items"`
	Page  struct {
		Page     int  `json:"page"`
		PageSize int  `json:"pageSize"`
		HasNext  bool `json:"hasNext"`
	} `json:"page"`
	Total int64 `json:"total"`
}

type skillDTO struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

type detailResponse struct {
	ID           int64       `json:"id"`
	EmployeeCode string      `json:"employeeCode"`
	FirstName    string      `json:"firstName"`
	LastName     string      `json:"lastName"`
	PhotoURL     *string     `json:"photoUrl"`
	Contact      contactDTO  `json:"contact"`
	Skills       []skillDTO  `json:"skills"`
	Version      int32       `json:"version"`
	Role         entity.Role `json:"role"`
	IsActive     bool        `json:"isActive"`
}

type contactDTO struct {
	MobilePhone *string `json:"mobilePhone"`
	Email       *string `json:"email"`
	Address     *string `json:"address"`
}

type topSkillDTO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	UsedBy int64  `json:"usedBy"`
}

type statsResponse struct {
	Employees        int64         `json:"employees"`
	CompleteProfiles int64         `json:"completeProfiles"`
	TopSkills        []topSkillDTO `json:"topSkills"`
	RecentChanges    []recentDTO   `json:"recentChanges"`
}

type recentDTO struct {
	ID            int64       `json:"id"`
	EmployeeID    int64       `json:"employeeId"`
	EmployeeCode  string      `json:"employeeCode"`
	Name          string      `json:"name"`
	Field         string      `json:"field"`
	ChangedAt     time.Time   `json:"changedAt"`
	ChangedByRole entity.Role `json:"changedByRole"`
}

// List: GET /admin/employees?q=&page=&pageSize=
func (h *Handler) List(c *gin.Context) {
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
	res, err := h.uc.List(c.Request.Context(), middleware.ActorFrom(c), ListInput{
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
func (h *Handler) Get(c *gin.Context) {
	id, ok := httputil.PathID(c, "id")
	if !ok {
		return
	}
	d, err := h.uc.Get(c.Request.Context(), middleware.ActorFrom(c), id)
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
func (h *Handler) Stats(c *gin.Context) {
	s, err := h.uc.Stats(c.Request.Context(), middleware.ActorFrom(c))
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
