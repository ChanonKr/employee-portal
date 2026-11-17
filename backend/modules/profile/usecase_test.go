package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/entity"
)

// fakeRepo is an in-memory Repository. WithTx works on a copy and only keeps it
// when fn succeeds, mimicking commit/rollback.
type fakeRepo struct {
	emp       entity.Employee
	skills    []int64        // employee's skill ids
	catalog   []entity.Skill // all skills, in display order
	logs      []entity.Change
	failSave  error
	failLog   error
	committed bool
}

func (f *fakeRepo) GetProfile(context.Context, int64) (entity.Profile, error) {
	return entity.Profile{Employee: f.emp}, nil
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
	emp    entity.Employee
	skills []int64
	logs   []entity.Change
}

func (t *fakeTx) skillsWhere(keep func(id int64) bool) []entity.Skill {
	out := []entity.Skill{}
	for _, s := range t.repo.catalog {
		if keep(s.ID) {
			out = append(out, s)
		}
	}
	return out
}

func (t *fakeTx) ListEmployeeSkills(context.Context, int64) ([]entity.Skill, error) {
	return t.skillsWhere(func(id int64) bool { return contains(t.skills, id) }), nil
}

func (t *fakeTx) GetSkillsByIDs(_ context.Context, ids []int64) ([]entity.Skill, error) {
	return t.skillsWhere(func(id int64) bool { return contains(ids, id) }), nil
}

func (t *fakeTx) ReplaceSkills(_ context.Context, _ int64, ids []int64) error {
	t.skills = append([]int64(nil), ids...)
	return nil
}

func (t *fakeTx) BumpVersion(_ context.Context, e entity.Employee) (int32, error) {
	t.emp.Version = e.Version + 1
	return t.emp.Version, nil
}

func contains(ids []int64, id int64) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func (t *fakeTx) GetForUpdate(_ context.Context, id int64) (entity.Employee, error) {
	if id != t.emp.ID {
		return entity.Employee{}, entity.ErrNotFound
	}
	return t.emp, nil
}

func (t *fakeTx) SaveContact(_ context.Context, e entity.Employee, _ entity.ContactField) (int32, error) {
	if t.repo.failSave != nil {
		return 0, t.repo.failSave
	}
	e.Version++
	t.emp = e
	return e.Version, nil
}

func (t *fakeTx) AppendChange(_ context.Context, _ int64, c entity.Change, _ entity.Actor) error {
	if t.repo.failLog != nil {
		return t.repo.failLog
	}
	t.logs = append(t.logs, c)
	return nil
}

func ptr(s string) *string { return &s }

func newRepo() *fakeRepo {
	return &fakeRepo{
		emp:    entity.Employee{ID: 10, Version: 3, MobilePhone: ptr("0811111111")},
		skills: []int64{1, 3},
		catalog: []entity.Skill{
			{ID: 1, Name: "Go", IsActive: true},
			{ID: 2, Name: "React", IsActive: true},
			{ID: 3, Name: "Flutter", IsActive: false},
			{ID: 4, Name: "Rust", IsActive: false},
		},
	}
}

var actor = entity.Actor{UserID: 1, EmployeeID: 10, Role: entity.RoleEmployee}

func TestUpdateContact(t *testing.T) {
	ctx := context.Background()

	t.Run("updates, bumps version and logs one change", func(t *testing.T) {
		repo := newRepo()
		v, err := NewUsecase(repo).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "081-234-5678", Version: 3,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Equal(t, "0812345678", *repo.emp.MobilePhone)
		require.Len(t, repo.logs, 1)
		assert.Equal(t, entity.Change{Field: entity.FieldMobilePhone, Old: "0811111111", New: "0812345678"}, repo.logs[0])
	})

	t.Run("same value bumps version without a log", func(t *testing.T) {
		repo := newRepo()
		v, err := NewUsecase(repo).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "0811111111", Version: 3,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Empty(t, repo.logs)
	})

	t.Run("stale version returns conflict and changes nothing", func(t *testing.T) {
		repo := newRepo()
		_, err := NewUsecase(repo).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "0812345678", Version: 2,
		})
		assert.ErrorIs(t, err, entity.ErrConflict)
		assert.False(t, repo.committed)
		assert.Equal(t, "0811111111", *repo.emp.MobilePhone)
	})

	t.Run("unknown field is a validation error before any DB work", func(t *testing.T) {
		repo := newRepo()
		_, err := NewUsecase(repo).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "first_name", Value: "Hacker", Version: 3,
		})
		assert.ErrorIs(t, err, entity.ErrValidation)
		assert.False(t, repo.committed)
	})

	t.Run("failed log insert rolls back the update", func(t *testing.T) {
		repo := newRepo()
		repo.failLog = errors.New("insert failed")
		_, err := NewUsecase(repo).UpdateContact(ctx, UpdateContactInput{
			Actor: actor, Field: "mobile", Value: "0812345678", Version: 3,
		})
		assert.Error(t, err)
		assert.EqualValues(t, 3, repo.emp.Version)
		assert.Equal(t, "0811111111", *repo.emp.MobilePhone)
	})

	t.Run("missing employee is not found", func(t *testing.T) {
		repo := newRepo()
		_, err := NewUsecase(repo).UpdateContact(ctx, UpdateContactInput{
			Actor: entity.Actor{UserID: 2, EmployeeID: 99}, Field: "email", Value: "a@example.com", Version: 1,
		})
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestSetSkills(t *testing.T) {
	ctx := context.Background()
	set := func(repo *fakeRepo, ids []int64, version int32) (int32, error) {
		return NewUsecase(repo).SetSkills(ctx, SetSkillsInput{Actor: actor, SkillIDs: ids, Version: version})
	}

	t.Run("replaces the set, bumps version and logs names", func(t *testing.T) {
		repo := newRepo() // has Go + Flutter(inactive)
		v, err := set(repo, []int64{1, 2}, 3)
		require.NoError(t, err)
		assert.EqualValues(t, 4, v)
		assert.Equal(t, []int64{1, 2}, repo.skills)
		require.Len(t, repo.logs, 1)
		assert.Equal(t, entity.Change{Field: entity.FieldSkills,
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
		assert.ErrorIs(t, err, entity.ErrValidation)
		assert.False(t, repo.committed)
		assert.Equal(t, []int64{1, 3}, repo.skills)
	})

	t.Run("unknown skill id is rejected", func(t *testing.T) {
		_, err := set(newRepo(), []int64{1, 999}, 3)
		assert.ErrorIs(t, err, entity.ErrValidation)
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
		assert.ErrorIs(t, err, entity.ErrConflict)
		assert.Equal(t, []int64{1, 3}, repo.skills)
	})

	t.Run("duplicate ids are rejected before any DB work", func(t *testing.T) {
		repo := newRepo()
		_, err := set(repo, []int64{1, 1}, 3)
		assert.ErrorIs(t, err, entity.ErrValidation)
		assert.False(t, repo.committed)
	})
}
