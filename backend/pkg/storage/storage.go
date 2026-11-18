// Package storage keeps uploaded files. Local is the only implementation today;
// an S3/MinIO implementation can satisfy the same methods when the API runs on
// more than one machine.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"

	"employee-portal/modules/shared"
)

// Stored names are generated here (UUID + extension), never taken from the client,
// which rules out path traversal and overwriting another user's file.
var validName = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.(jpg)$`)

type Local struct {
	dir string
}

func NewLocal(dir string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &Local{dir: dir}, nil
}

// Save writes data under a new random name and returns that name.
// It writes to a temp file first so a crash never leaves a half-written photo.
func (l *Local) Save(_ context.Context, data []byte, ext string) (string, error) {
	name := uuid.NewString() + ext
	if !validName.MatchString(name) {
		return "", fmt.Errorf("unsupported extension %q", ext)
	}
	tmp, err := os.CreateTemp(l.dir, ".upload-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck // no-op after rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(l.dir, name)); err != nil {
		return "", err
	}
	return name, nil
}

// Delete removes a stored file. A missing file is not an error.
func (l *Local) Delete(_ context.Context, name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid file name %q", name)
	}
	err := os.Remove(filepath.Join(l.dir, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Open returns the file for reading. Unknown or malformed names are ErrNotFound.
func (l *Local) Open(_ context.Context, name string) (io.ReadSeekCloser, time.Time, error) {
	if !validName.MatchString(name) {
		return nil, time.Time{}, shared.ErrNotFound
	}
	f, err := os.Open(filepath.Join(l.dir, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, time.Time{}, shared.ErrNotFound
	}
	if err != nil {
		return nil, time.Time{}, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, time.Time{}, err
	}
	return f, st.ModTime(), nil
}
