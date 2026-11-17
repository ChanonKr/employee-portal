package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmployee_ChangeName(t *testing.T) {
	t.Run("records each changed part", func(t *testing.T) {
		e := Employee{FirstName: "Somchai", LastName: "Jaidi", Version: 2}
		changes, err := e.ChangeName(" Somchai ", "Jaidee", 2)
		require.NoError(t, err)
		assert.Equal(t, []Change{{Field: FieldLastName, Old: "Jaidi", New: "Jaidee"}}, changes)
		assert.Equal(t, "Jaidee", e.LastName)
	})

	t.Run("both parts", func(t *testing.T) {
		e := Employee{FirstName: "A", LastName: "B", Version: 1}
		changes, err := e.ChangeName("C", "D", 1)
		require.NoError(t, err)
		assert.Len(t, changes, 2)
	})

	t.Run("no change returns an empty list", func(t *testing.T) {
		e := Employee{FirstName: "Suda", LastName: "Rakdee", Version: 1}
		changes, err := e.ChangeName("Suda", "  Rakdee", 1)
		require.NoError(t, err)
		assert.Empty(t, changes)
	})

	t.Run("validation and version", func(t *testing.T) {
		e := Employee{FirstName: "A", LastName: "B", Version: 3}
		_, err := e.ChangeName("", "B", 3)
		assert.ErrorIs(t, err, ErrValidation)
		_, err = e.ChangeName("A", strings.Repeat("x", 101), 3)
		assert.ErrorIs(t, err, ErrValidation)
		_, err = e.ChangeName("New", "B", 2)
		assert.ErrorIs(t, err, ErrConflict)
		assert.Equal(t, "A", e.FirstName, "unchanged on error")
	})
}

func TestEmployeeQuery_LikePattern(t *testing.T) {
	assert.Nil(t, EmployeeQuery{Search: "   "}.LikePattern())
	assert.Equal(t, "%som%", *EmployeeQuery{Search: " som "}.LikePattern())
	assert.Equal(t, `%100\%\_\\%`, *EmployeeQuery{Search: `100%_\`}.LikePattern(), "user wildcards match literally")
}
