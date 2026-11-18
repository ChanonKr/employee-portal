//go:build integration

package integration

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/profile"
	"employee-portal/modules/shared"
)

func testPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.NRGBA{B: 255, A: 255})
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestUpdatePhoto_ReplacesFileAndLogs(t *testing.T) {
	ctx := context.Background()
	userID, empID, _ := newEmployee(t, "0811111111")
	uc, dir := newProfileUsecaseWithDir(t)
	actor := shared.Actor{UserID: userID, EmployeeID: empID}

	v, first, err := uc.UpdatePhoto(ctx, profile.UpdatePhotoInput{Actor: actor, File: bytes.NewReader(testPNG(t, 1200, 800)), Version: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 2, v)
	assert.Equal(t, []string{first}, filesIn(t, dir))

	cfg, format, err := image.DecodeConfig(mustOpen(t, filepath.Join(dir, first)))
	require.NoError(t, err)
	assert.Equal(t, "jpeg", format)
	assert.Equal(t, 512, cfg.Width)

	_, second, err := uc.UpdatePhoto(ctx, profile.UpdatePhotoInput{Actor: actor, File: bytes.NewReader(testPNG(t, 64, 64)), Version: 2})
	require.NoError(t, err)
	assert.Equal(t, []string{second}, filesIn(t, dir), "old file removed after commit")

	var path string
	require.NoError(t, ownerPool.QueryRow(ctx, `SELECT photo_path FROM employees WHERE id = $1`, empID).Scan(&path))
	assert.Equal(t, second, path)
	assert.Equal(t, 2, countLogs(t, empID))
}

func TestUpdatePhoto_FailureLeavesNoOrphanFile(t *testing.T) {
	ctx := context.Background()
	_, empID, _ := newEmployee(t, "0811111111")
	uc, dir := newProfileUsecaseWithDir(t)

	// Stale version → conflict after the file was saved.
	_, _, err := uc.UpdatePhoto(ctx, profile.UpdatePhotoInput{
		Actor: shared.Actor{UserID: 1, EmployeeID: empID}, File: bytes.NewReader(testPNG(t, 10, 10)), Version: 99,
	})
	assert.ErrorIs(t, err, shared.ErrConflict)
	assert.Empty(t, filesIn(t, dir))

	// Log insert fails (unknown changed_by) → DB rolls back and the file is deleted.
	_, _, err = uc.UpdatePhoto(ctx, profile.UpdatePhotoInput{
		Actor: shared.Actor{UserID: 999_999_999, EmployeeID: empID}, File: bytes.NewReader(testPNG(t, 10, 10)), Version: 1,
	})
	require.Error(t, err)
	assert.Empty(t, filesIn(t, dir))

	var path *string
	require.NoError(t, ownerPool.QueryRow(ctx, `SELECT photo_path FROM employees WHERE id = $1`, empID).Scan(&path))
	assert.Nil(t, path)
}

func TestUpdatePhoto_RejectsNonImage(t *testing.T) {
	_, empID, _ := newEmployee(t, "0811111111")
	uc, dir := newProfileUsecaseWithDir(t)
	_, _, err := uc.UpdatePhoto(context.Background(), profile.UpdatePhotoInput{
		Actor: shared.Actor{UserID: 1, EmployeeID: empID}, File: bytes.NewReader([]byte("not an image, renamed to .jpg")), Version: 1,
	})
	assert.ErrorIs(t, err, shared.ErrPhotoFormat)
	assert.Empty(t, filesIn(t, dir))
}

func mustOpen(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	return f
}
