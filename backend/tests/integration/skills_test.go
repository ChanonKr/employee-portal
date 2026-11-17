//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/entity"
	"employee-portal/modules/profile"
	"employee-portal/modules/skill"
)

// newSkill adds a uniquely named skill so tests do not depend on the seed catalog.
func newSkill(t *testing.T, active bool) (int64, string) {
	t.Helper()
	name := fmt.Sprintf("S%d", time.Now().UnixNano()%1_000_000_000)
	var id int64
	require.NoError(t, ownerPool.QueryRow(context.Background(),
		`INSERT INTO skills (name, is_active, sort_order) VALUES ($1, $2, 1000) RETURNING id`,
		name, active).Scan(&id))
	return id, name
}

func employeeSkillIDs(t *testing.T, empID int64) []int64 {
	t.Helper()
	rows, err := ownerPool.Query(context.Background(),
		`SELECT skill_id FROM employee_skills WHERE employee_id = $1 ORDER BY skill_id`, empID)
	require.NoError(t, err)
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	return ids
}

func TestSetSkills_ReplacesSetAndLogsNames(t *testing.T) {
	ctx := context.Background()
	userID, empID, _ := newEmployee(t, "0811111111")
	a, aName := newSkill(t, true)
	b, bName := newSkill(t, true)
	uc := newProfileUsecase(t)
	actor := entity.Actor{UserID: userID, EmployeeID: empID}

	v, err := uc.SetSkills(ctx, profile.SetSkillsInput{Actor: actor, SkillIDs: []int64{a, b}, Version: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 2, v)
	assert.ElementsMatch(t, []int64{a, b}, employeeSkillIDs(t, empID))

	var oldV, newV string
	require.NoError(t, ownerPool.QueryRow(ctx, `
		SELECT old_value::text, new_value::text FROM profile_change_logs
		WHERE employee_id = $1 AND field = 'skills'`, empID).Scan(&oldV, &newV))
	assert.Equal(t, `[]`, oldV)
	assert.JSONEq(t, fmt.Sprintf(`[%q, %q]`, aName, bName), newV)

	// Same set in another order: version moves, no new log.
	v, err = uc.SetSkills(ctx, profile.SetSkillsInput{Actor: actor, SkillIDs: []int64{b, a}, Version: 2})
	require.NoError(t, err)
	assert.EqualValues(t, 3, v)
	assert.Equal(t, 1, countLogs(t, empID))

	// Clear all.
	_, err = uc.SetSkills(ctx, profile.SetSkillsInput{Actor: actor, SkillIDs: []int64{}, Version: 3})
	require.NoError(t, err)
	assert.Empty(t, employeeSkillIDs(t, empID))
}

func TestSetSkills_InactiveSkill(t *testing.T) {
	ctx := context.Background()
	userID, empID, _ := newEmployee(t, "0811111111")
	active, _ := newSkill(t, true)
	retired, _ := newSkill(t, false)
	uc := newProfileUsecase(t)
	actor := entity.Actor{UserID: userID, EmployeeID: empID}

	// Cannot newly pick a retired skill.
	_, err := uc.SetSkills(ctx, profile.SetSkillsInput{Actor: actor, SkillIDs: []int64{active, retired}, Version: 1})
	assert.ErrorIs(t, err, entity.ErrValidation)
	assert.Empty(t, employeeSkillIDs(t, empID))

	// An employee who already has it (picked before it was retired) keeps it.
	_, err = ownerPool.Exec(ctx, `INSERT INTO employee_skills (employee_id, skill_id) VALUES ($1, $2)`, empID, retired)
	require.NoError(t, err)
	_, err = uc.SetSkills(ctx, profile.SetSkillsInput{Actor: actor, SkillIDs: []int64{active, retired}, Version: 1})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{active, retired}, employeeSkillIDs(t, empID))

	// It is still hidden from the picker.
	catalog, err := skill.NewUsecase(skill.NewRepository(appDB)).ListActive(ctx)
	require.NoError(t, err)
	for _, s := range catalog {
		assert.NotEqual(t, retired, s.ID)
	}
}

func TestSetSkills_LogFailureRollsBackSkills(t *testing.T) {
	ctx := context.Background()
	_, empID, _ := newEmployee(t, "0811111111")
	a, _ := newSkill(t, true)
	uc := newProfileUsecase(t)

	ghost := entity.Actor{UserID: 999_999_999, EmployeeID: empID} // log insert fails (FK)
	_, err := uc.SetSkills(ctx, profile.SetSkillsInput{Actor: ghost, SkillIDs: []int64{a}, Version: 1})
	require.Error(t, err)

	assert.Empty(t, employeeSkillIDs(t, empID), "skills must roll back with the failed log")
	var version int32
	require.NoError(t, ownerPool.QueryRow(ctx, `SELECT version FROM employees WHERE id = $1`, empID).Scan(&version))
	assert.EqualValues(t, 1, version)
}
