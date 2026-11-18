package shared

import (
	"errors"
	"sort"
	"strings"
)

// Domain errors describe what went wrong, not how it is shown to the client.
// pkg/httputil maps them to HTTP status codes and error codes.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")       // stale version (optimistic locking)
	ErrDuplicate    = errors.New("already exists") // unique value taken (e.g. skill name)
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrValidation   = errors.New("validation failed")
)

// ValidationError carries per-field messages. errors.Is(err, ErrValidation) is true.
type ValidationError struct {
	Fields map[string]string
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Fields: map[string]string{field: message}}
}

func (e *ValidationError) Error() string {
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+e.Fields[k])
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

// ErrPasswordChangeRequired blocks every API except /auth/* until the user
// replaces a temporary password.
var ErrPasswordChangeRequired = errors.New("password change required")

// Photo rules shared with pkg/imaging, which checks and normalizes the image.
const MaxPhotoSide = 512 // stored photos fit in 512×512

var (
	ErrPhotoTooLarge = errors.New("photo is too large")
	ErrPhotoFormat   = errors.New("photo must be a JPEG, PNG or WebP image")
)
