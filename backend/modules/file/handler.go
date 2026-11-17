// Package file serves stored photos. It has no use case layer: it only streams a
// file by name, and pkg/storage already validates the name.
package file

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"employee-portal/pkg/httputil"
)

// Reader is implemented by pkg/storage.
type Reader interface {
	Open(ctx context.Context, name string) (io.ReadSeekCloser, time.Time, error)
}

type Handler struct {
	files Reader
}

func NewHandler(files Reader) *Handler {
	return &Handler{files: files}
}

// Photo: GET /files/photos/:name (signed-in users only).
func (h *Handler) Photo(c *gin.Context) {
	name := c.Param("name")
	f, modTime, err := h.files.Open(c.Request.Context(), name)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	defer f.Close()

	// Names are unique per upload, so the content never changes: cache privately.
	c.Header("Cache-Control", "private, max-age=31536000, immutable")
	c.Header("Content-Type", "image/jpeg")
	http.ServeContent(c.Writer, c.Request, name, modTime, f)
}
