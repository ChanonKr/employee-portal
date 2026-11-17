package skill

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/pkg/httputil"
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
