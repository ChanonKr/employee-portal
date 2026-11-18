package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/shared"
)

func TestLocal_SaveOpenDelete(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := NewLocal(dir)
	require.NoError(t, err)

	name, err := s.Save(ctx, []byte("jpeg-bytes"), ".jpg")
	require.NoError(t, err)
	assert.Regexp(t, validName, name)

	f, _, err := s.Open(ctx, name)
	require.NoError(t, err)
	got, _ := io.ReadAll(f)
	f.Close()
	assert.Equal(t, "jpeg-bytes", string(got))

	entries, _ := os.ReadDir(dir)
	assert.Len(t, entries, 1, "no temp files left behind")

	require.NoError(t, s.Delete(ctx, name))
	_, err = os.Stat(filepath.Join(dir, name))
	assert.True(t, os.IsNotExist(err))
	assert.NoError(t, s.Delete(ctx, name), "deleting twice is fine")
}

func TestLocal_RejectsUnsafeNames(t *testing.T) {
	ctx := context.Background()
	s, err := NewLocal(t.TempDir())
	require.NoError(t, err)

	for _, name := range []string{"../../etc/passwd", "a.jpg", "", "x/../y.jpg",
		"9f2c0000-0000-0000-0000-000000000000.png"} {
		_, _, err := s.Open(ctx, name)
		assert.ErrorIs(t, err, shared.ErrNotFound, name)
		assert.Error(t, s.Delete(ctx, name), name)
	}

	_, err = s.Save(ctx, []byte("x"), ".exe")
	assert.Error(t, err)
}
