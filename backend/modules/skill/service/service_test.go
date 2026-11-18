package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/shared"
	"employee-portal/modules/skill/domain"
)

type fakeRepo struct {
	created   []string
	updated   *UpdateInput
	reordered []domain.SkillOrder
}

func (f *fakeRepo) ListActive(context.Context) ([]domain.Skill, error)   { return nil, nil }
func (f *fakeRepo) ListAll(context.Context) ([]domain.SkillUsage, error) { return nil, nil }
func (f *fakeRepo) Create(_ context.Context, name string) (int64, error) {
	f.created = append(f.created, name)
	return 42, nil
}
func (f *fakeRepo) Update(_ context.Context, id int64, name *string, active *bool) error {
	f.updated = &UpdateInput{ID: id, Name: name, IsActive: active}
	return nil
}
func (f *fakeRepo) Reorder(_ context.Context, items []domain.SkillOrder) error {
	f.reordered = items
	return nil
}

var (
	admin    = shared.Actor{UserID: 1, Role: shared.RoleAdmin}
	employee = shared.Actor{UserID: 2, Role: shared.RoleEmployee}
)

func strp(s string) *string { return &s }
func boolp(b bool) *bool    { return &b }

func TestAdminUseCasesRejectEmployees(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{}
	svc := New(repo)

	_, err := svc.AdminList(ctx, employee)
	assert.ErrorIs(t, err, shared.ErrForbidden)
	_, err = svc.Create(ctx, employee, "Rust")
	assert.ErrorIs(t, err, shared.ErrForbidden)
	assert.ErrorIs(t, svc.Update(ctx, employee, UpdateInput{ID: 1, IsActive: boolp(false)}), shared.ErrForbidden)
	assert.ErrorIs(t, svc.Reorder(ctx, employee, []domain.SkillOrder{{ID: 1}}), shared.ErrForbidden)

	assert.Empty(t, repo.created)
	assert.Nil(t, repo.updated)
	assert.Nil(t, repo.reordered)
}

func TestCreate_NormalizesName(t *testing.T) {
	repo := &fakeRepo{}
	id, err := New(repo).Create(context.Background(), admin, "  Machine   Learning ")
	require.NoError(t, err)
	assert.EqualValues(t, 42, id)
	assert.Equal(t, []string{"Machine Learning"}, repo.created)

	_, err = New(repo).Create(context.Background(), admin, "   ")
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{}
	svc := New(repo)

	require.NoError(t, svc.Update(ctx, admin, UpdateInput{ID: 3, Name: strp(" Go "), IsActive: boolp(false)}))
	assert.Equal(t, "Go", *repo.updated.Name)
	assert.False(t, *repo.updated.IsActive)

	assert.ErrorIs(t, svc.Update(ctx, admin, UpdateInput{ID: 3}), shared.ErrValidation, "nothing to update")
	assert.ErrorIs(t, svc.Update(ctx, admin, UpdateInput{ID: 3, Name: strp("")}), shared.ErrValidation)
}
