package httputil

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"employee-portal/modules/entity"
)

// RequestIDKey is the gin.Context key set by the request-id middleware.
const RequestIDKey = "requestId"

type ErrorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields"`
	RequestID string            `json:"requestId"`
}

type errorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// WriteError maps a domain error to an HTTP status and the standard envelope.
func WriteError(c *gin.Context, err error) {
	status, body := classify(err)
	body.RequestID = c.GetString(RequestIDKey)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(c.Request.Context(), "request failed",
			"error", err, "requestId", body.RequestID, "path", c.FullPath())
	}
	c.AbortWithStatusJSON(status, errorEnvelope{Error: body})
}

// Abort writes an envelope for errors that originate in the HTTP layer.
func Abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorEnvelope{Error: ErrorBody{
		Code: code, Message: message, RequestID: c.GetString(RequestIDKey),
	}})
}

func classify(err error) (int, ErrorBody) {
	var ve *entity.ValidationError
	switch {
	case errors.As(err, &ve):
		return http.StatusBadRequest, ErrorBody{Code: "VALIDATION_ERROR", Message: "Some fields are invalid.", Fields: ve.Fields}
	case errors.Is(err, ErrMissingIfMatch):
		return http.StatusPreconditionRequired, ErrorBody{Code: "PRECONDITION_REQUIRED", Message: "If-Match header with the current version is required."}
	case errors.Is(err, ErrInvalidIfMatch):
		return http.StatusBadRequest, ErrorBody{Code: "VALIDATION_ERROR", Message: "If-Match header is invalid.", Fields: map[string]string{"If-Match": "must be a quoted version number"}}
	case errors.Is(err, entity.ErrUnauthorized):
		return http.StatusUnauthorized, ErrorBody{Code: "UNAUTHORIZED", Message: "Please sign in."}
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, ErrorBody{Code: "FORBIDDEN", Message: "You do not have access to this resource."}
	case errors.Is(err, entity.ErrNotFound):
		return http.StatusNotFound, ErrorBody{Code: "NOT_FOUND", Message: "Not found."}
	case errors.Is(err, entity.ErrConflict):
		return http.StatusConflict, ErrorBody{Code: "VERSION_CONFLICT", Message: "Profile was updated elsewhere. Please reload."}
	default:
		return http.StatusInternalServerError, ErrorBody{Code: "INTERNAL_ERROR", Message: "Something went wrong. Please try again."}
	}
}
