package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/history/domain"
)

type fakeRetentionRepo struct {
	res        domain.PurgeResult
	err        error
	gotCutoff  time.Time
	gotKeepNew int
}

func (f *fakeRetentionRepo) Purge(_ context.Context, cutoff time.Time, keep int) (domain.PurgeResult, error) {
	f.gotCutoff, f.gotKeepNew = cutoff, keep
	return f.res, f.err
}

type fakeDeleter struct {
	deleted []string
	failOn  string
}

func (f *fakeDeleter) Delete(_ context.Context, name string) error {
	if name == f.failOn {
		return errors.New("disk error")
	}
	f.deleted = append(f.deleted, name)
	return nil
}

func newTestRetention(repo RetentionRepository, files FileDeleter) *Retention {
	r := NewRetention(repo, files, 90, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r.now = func() time.Time { return time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC) }
	return r
}

func TestRetention_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("purges with the cutoff and the photo allowance, then deletes unused files", func(t *testing.T) {
		repo := &fakeRetentionRepo{res: domain.PurgeResult{Ran: true, DeletedChanges: 4, UnusedPhotos: []string{"a.jpg", "b.jpg", "c.jpg"}}}
		files := &fakeDeleter{failOn: "b.jpg"}
		res, err := newTestRetention(repo, files).Run(ctx)
		require.NoError(t, err, "a file that cannot be deleted does not fail the run")
		assert.Equal(t, time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), repo.gotCutoff)
		assert.Equal(t, domain.KeepPhotosBeyondRetention, repo.gotKeepNew)
		assert.EqualValues(t, 4, res.DeletedChanges)
		assert.Equal(t, []string{"a.jpg", "c.jpg"}, files.deleted)
	})

	t.Run("another instance is purging: nothing is deleted", func(t *testing.T) {
		files := &fakeDeleter{}
		res, err := newTestRetention(&fakeRetentionRepo{res: domain.PurgeResult{Ran: false, UnusedPhotos: []string{"a.jpg"}}}, files).Run(ctx)
		require.NoError(t, err)
		assert.False(t, res.Ran)
		assert.Empty(t, files.deleted)
	})

	t.Run("a database error deletes no files", func(t *testing.T) {
		files := &fakeDeleter{}
		_, err := newTestRetention(&fakeRetentionRepo{err: errors.New("db down"), res: domain.PurgeResult{Ran: true, UnusedPhotos: []string{"a.jpg"}}}, files).Run(ctx)
		assert.Error(t, err)
		assert.Empty(t, files.deleted)
	})
}
