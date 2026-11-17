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
	logs      []entity.Change
	failSave  error
	failLog   error
	committed bool
}

func (f *fakeRepo) GetProfile(context.Context, int64) (entity.Profile, error) {
	return entity.Profile{Employee: f.emp}, nil
}

func (f *fakeRepo) WithTx(_ context.Context, fn func(TxRepository) error) error {
	tx := &fakeTx{repo: f, emp: f.emp}
	if err := fn(tx); err != nil {
		return err // rollback: nothing copied back
	}
	f.emp = tx.emp
	f.logs = append(f.logs, tx.logs...)
	f.committed = true
	return nil
}

type fakeTx struct {
	repo *fakeRepo
	emp  entity.Employee
	logs []entity.Change
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
	return &fakeRepo{emp: entity.Employee{ID: 10, Version: 3, MobilePhone: ptr("0811111111")}}
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
