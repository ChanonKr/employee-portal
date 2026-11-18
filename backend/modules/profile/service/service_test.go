package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/profile/domain"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

// fakeRepo is an in-memory Repository. WithTx works on a copy and only keeps it
// when fn succeeds, mimicking commit/rollback.
type fakeRepo struct {
	emp       domain.Employee
	skills    []int64             // employee's skill ids
	catalog   []skilldomain.Skill // all skills, in display order
	logs      []historydomain.Change
	failSave  error
	failLog   error
	committed bool
}

func (f *fakeRepo) GetProfile(context.Context, int64) (domain.Profile, error) {
	return domain.Profile{Employee: f.emp}, nil
}

func (f *fakeRepo) WithTx(_ context.Context, fn func(TxRepository) error) error {
	tx := &fakeTx{repo: f, emp: f.emp, skills: append([]int64(nil), f.skills...)}
	if err := fn(tx); err != nil {
		return err // rollback: nothing copied back
	}
	f.emp = tx.emp
	f.skills = tx.skills
	f.logs = append(f.logs, tx.logs...)
	f.committed = true
	return nil
}

type fakeTx struct {
	repo   *fakeRepo
	emp    domain.Employee
	skills []int64
	logs   []historydomain.Change
}

func (t *fakeTx) skillsWhere(keep func(id int64) bool) []skilldomain.Skill {
	out := []skilldomain.Skill{}
	for _, s := range t.repo.catalog {
		if keep(s.ID) {
			out = append(out, s)
		}
	}
	return out
}

func (t *fakeTx) ListEmployeeSkills(context.Context, int64) ([]skilldomain.Skill, error) {
	return t.skillsWhere(func(id int64) bool { return contains(t.skills, id) }), nil
}

func (t *fakeTx) GetSkillsByIDs(_ context.Context, ids []int64) ([]skilldomain.Skill, error) {
	return t.skillsWhere(func(id int64) bool { return contains(ids, id) }), nil
}

func (t *fakeTx) ReplaceSkills(_ context.Context, _ int64, ids []int64) error {
	t.skills = append([]int64(nil), ids...)
	return nil
}

func (t *fakeTx) BumpVersion(_ context.Context, e domain.Employee) (int32, error) {
	t.emp.Version = e.Version + 1
	return t.emp.Version, nil
}

func (t *fakeTx) SaveName(_ context.Context, e domain.Employee) (int32, error) {
	e.Version++
	t.emp = e
	return e.Version, nil
}

func (t *fakeTx) SavePhoto(_ context.Context, e domain.Employee) (int32, error) {
	if t.repo.failSave != nil {
		return 0, t.repo.failSave
	}
	e.Version++
	t.emp = e
	return e.Version, nil
}

// fakePhotos accepts any input except the literal "bad".
type fakePhotos struct{}

func (fakePhotos) Normalize(r io.Reader) ([]byte, error) {
	b, _ := io.ReadAll(r)
	if string(b) == "bad" {
		return nil, shared.ErrPhotoFormat
	}
	return []byte("jpeg"), nil
}

// fakeFiles records which files exist.
type fakeFiles struct {
	n     int
	files map[string]bool
}

func (f *fakeFiles) Save(context.Context, []byte, string) (string, error) {
	f.n++
	name := fmt.Sprintf("new-%d.jpg", f.n)
	f.files[name] = true
	return name, nil
}

func (f *fakeFiles) Delete(_ context.Context, name string) error {
	delete(f.files, name)
	return nil
}

func contains(ids []int64, id int64) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func (t *fakeTx) GetForUpdate(_ context.Context, id int64) (domain.Employee, error) {
	if id != t.emp.ID {
		return domain.Employee{}, shared.ErrNotFound
	}
	return t.emp, nil
}

func (t *fakeTx) SaveContact(_ context.Context, e domain.Employee, _ domain.ContactField) (int32, error) {
	if t.repo.failSave != nil {
		return 0, t.repo.failSave
	}
	e.Version++
	t.emp = e
	return e.Version, nil
}

func (t *fakeTx) AppendChange(_ context.Context, _ int64, c historydomain.Change, _ shared.Actor) error {
	if t.repo.failLog != nil {
		return t.repo.failLog
	}
	t.logs = append(t.logs, c)
	return nil
}

func ptr(s string) *string { return &s }

func newRepo() *fakeRepo {
	return &fakeRepo{
		emp:    domain.Employee{ID: 10, Version: 3, MobilePhone: ptr("0811111111")},
		skills: []int64{1, 3},
		catalog: []skilldomain.Skill{
			{ID: 1, Name: "Go", IsActive: true},
			{ID: 2, Name: "React", IsActive: true},
			{ID: 3, Name: "Flutter", IsActive: false},
			{ID: 4, Name: "Rust", IsActive: false},
		},
	}
}

var actor = shared.Actor{UserID: 1, EmployeeID: 10, Role: shared.RoleEmployee}

func TestUpdateContact(t *testing.T) {
	ctx := context.Background()

	t.Run("updates, bumps version and logs one change", func(t *testing.T) {
		repo := newRepo()
		v, err := New(Deps{Repo: repo}).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "081-234-5678", Version: 3,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Equal(t, "0812345678", *repo.emp.MobilePhone)
		require.Len(t, repo.logs, 1)
		assert.Equal(t, historydomain.Change{Field: historydomain.FieldMobilePhone, Old: "0811111111", New: "0812345678"}, repo.logs[0])
	})

	t.Run("same value bumps version without a log", func(t *testing.T) {
		repo := newRepo()
		v, err := New(Deps{Repo: repo}).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "0811111111", Version: 3,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Empty(t, repo.logs)
	})

	t.Run("stale version returns conflict and changes nothing", func(t *testing.T) {
		repo := newRepo()
		_, err := New(Deps{Repo: repo}).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "0812345678", Version: 2,
		})
		assert.ErrorIs(t, err, shared.ErrConflict)
		assert.False(t, repo.committed)
		assert.Equal(t, "0811111111", *repo.emp.MobilePhone)
	})

	t.Run("unknown field is a validation error before any DB work", func(t *testing.T) {
		repo := newRepo()
		_, err := New(Deps{Repo: repo}).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "first_name", Value: "Hacker", Version: 3,
		})
		assert.ErrorIs(t, err, shared.ErrValidation)
		assert.False(t, repo.committed)
	})

	t.Run("failed log insert rolls back the update", func(t *testing.T) {
		repo := newRepo()
		repo.failLog = errors.New("insert failed")
		_, err := New(Deps{Repo: repo}).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "0812345678", Version: 3,
		})
		assert.Error(t, err)
		assert.EqualValues(t, 3, repo.emp.Version)
		assert.Equal(t, "0811111111", *repo.emp.MobilePhone)
	})

	t.Run("missing employee is not found", func(t *testing.T) {
		repo := newRepo()
		_, err := New(Deps{Repo: repo}).UpdateContact(ctx, UpdateContactInput{
			Actor: shared.Actor{UserID: 2, EmployeeID: 99}, Field: "email", Value: "a@example.com", Version: 1,
		})
		assert.ErrorIs(t, err, shared.ErrNotFound)
	})
}

func TestSetSkills(t *testing.T) {
	ctx := context.Background()
	set := func(repo *fakeRepo, ids []int64, version int32) (int32, error) {
		return New(Deps{Repo: repo}).SetSkills(ctx, SetSkillsInput{Actor: actor, SkillIDs: ids, Version: version})
	}

	t.Run("replaces the set, bumps version and logs names", func(t *testing.T) {
		repo := newRepo() // has Go + Flutter(inactive)
		v, err := set(repo, []int64{1, 2}, 3)
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Equal(t, []int64{1, 2}, repo.skills)
		require.Len(t, repo.logs, 1)
		assert.Equal(t, historydomain.Change{Field: historydomain.FieldSkills,
			Old: []string{"Go", "Flutter"}, New: []string{"Go", "React"}}, repo.logs[0])
	})

	t.Run("same set bumps version without a log", func(t *testing.T) {
		repo := newRepo()
		v, err := set(repo, []int64{3, 1}, 3)
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Empty(t, repo.logs)
	})

	t.Run("keeps an inactive skill the employee already has", func(t *testing.T) {
		repo := newRepo()
		_, err := set(repo, []int64{2, 3}, 3)
		assert.NoError(t, err)
	})

	t.Run("cannot add an inactive skill", func(t *testing.T) {
		repo := newRepo()
		_, err := set(repo, []int64{1, 4}, 3)
		assert.ErrorIs(t, err, shared.ErrValidation)
		assert.False(t, repo.committed)
		assert.Equal(t, []int64{1, 3}, repo.skills)
	})

	t.Run("unknown skill id is rejected", func(t *testing.T) {
		_, err := set(newRepo(), []int64{1, 999}, 3)
		assert.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("empty list clears all skills", func(t *testing.T) {
		repo := newRepo()
		_, err := set(repo, []int64{}, 3)
		require.NoError(t, err)
		assert.Empty(t, repo.skills)
		require.Len(t, repo.logs, 1)
		assert.Equal(t, []string{}, repo.logs[0].New)
	})

	t.Run("stale version is a conflict", func(t *testing.T) {
		repo := newRepo()
		_, err := set(repo, []int64{1}, 2)
		assert.ErrorIs(t, err, shared.ErrConflict)
		assert.Equal(t, []int64{1, 3}, repo.skills)
	})

	t.Run("duplicate ids are rejected before any DB work", func(t *testing.T) {
		repo := newRepo()
		_, err := set(repo, []int64{1, 1}, 3)
		assert.ErrorIs(t, err, shared.ErrValidation)
		assert.False(t, repo.committed)
	})
}

func TestUpdatePhoto(t *testing.T) {
	ctx := context.Background()
	setup := func(oldPhoto *string) (*fakeRepo, *fakeFiles, *Service) {
		repo := newRepo()
		repo.emp.PhotoPath = oldPhoto
		files := &fakeFiles{files: map[string]bool{}}
		if oldPhoto != nil {
			files.files[*oldPhoto] = true
		}
		return repo, files, New(Deps{Repo: repo, Photos: fakePhotos{}, Files: files})
	}
	upload := func(svc *Service, body string, version int32) (int32, string, error) {
		return svc.UpdatePhoto(ctx, UpdatePhotoInput{Actor: actor, File: strings.NewReader(body), Version: version})
	}

	t.Run("saves new file, logs, deletes the replaced file", func(t *testing.T) {
		repo, files, svc := setup(ptr("old.jpg"))
		v, name, err := upload(svc, "png", 3)
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Equal(t, "new-1.jpg", name)
		assert.Equal(t, "new-1.jpg", *repo.emp.PhotoPath)
		assert.Equal(t, map[string]bool{"new-1.jpg": true}, files.files)
		require.Len(t, repo.logs, 1)
		assert.Equal(t, historydomain.Change{Field: historydomain.FieldPhoto, Old: "old.jpg", New: "new-1.jpg"}, repo.logs[0])
	})

	t.Run("first photo has nil old value", func(t *testing.T) {
		repo, _, svc := setup(nil)
		_, _, err := upload(svc, "png", 3)
		require.NoError(t, err)
		assert.Nil(t, repo.logs[0].Old)
	})

	t.Run("stale version deletes the new file and keeps the old one", func(t *testing.T) {
		repo, files, svc := setup(ptr("old.jpg"))
		_, _, err := upload(svc, "png", 2)
		assert.ErrorIs(t, err, shared.ErrConflict)
		assert.Equal(t, map[string]bool{"old.jpg": true}, files.files, "no orphan file")
		assert.Equal(t, "old.jpg", *repo.emp.PhotoPath)
	})

	t.Run("failed log insert deletes the new file", func(t *testing.T) {
		repo, files, svc := setup(ptr("old.jpg"))
		repo.failLog = errors.New("insert failed")
		_, _, err := upload(svc, "png", 3)
		assert.Error(t, err)
		assert.Equal(t, map[string]bool{"old.jpg": true}, files.files)
	})

	t.Run("invalid image never reaches storage or the DB", func(t *testing.T) {
		repo, files, svc := setup(nil)
		_, _, err := upload(svc, "bad", 3)
		assert.ErrorIs(t, err, shared.ErrPhotoFormat)
		assert.Zero(t, files.n)
		assert.False(t, repo.committed)
	})
}

func TestUpdateName(t *testing.T) {
	ctx := context.Background()
	admin := shared.Actor{UserID: 99, Role: shared.RoleAdmin}

	t.Run("admin renames, each part logged", func(t *testing.T) {
		repo := newRepo()
		repo.emp.FirstName, repo.emp.LastName = "Somchai", "Jaidi"
		v, err := New(Deps{Repo: repo}).UpdateName(ctx, UpdateNameInput{
			Actor: admin, EmployeeID: 10, FirstName: "Somchai", LastName: "Jaidee", Version: 3,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Equal(t, "Jaidee", repo.emp.LastName)
		assert.Equal(t, []historydomain.Change{{Field: historydomain.FieldLastName, Old: "Jaidi", New: "Jaidee"}}, repo.logs)
	})

	t.Run("employees cannot rename, even themselves", func(t *testing.T) {
		repo := newRepo()
		_, err := New(Deps{Repo: repo}).UpdateName(ctx, UpdateNameInput{
			Actor: actor, EmployeeID: actor.EmployeeID, FirstName: "X", LastName: "Y", Version: 3,
		})
		assert.ErrorIs(t, err, shared.ErrForbidden)
		assert.False(t, repo.committed)
	})
}
