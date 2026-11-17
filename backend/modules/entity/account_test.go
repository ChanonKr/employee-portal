package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEmployeeCode(t *testing.T) {
	c, err := NormalizeEmployeeCode(" emp-00042 ")
	require.NoError(t, err)
	assert.Equal(t, "EMP-00042", c)
	for _, bad := range []string{"", "AB", "EMP 1", "EMP_1", strings.Repeat("A", 21), "ปปป"} {
		_, err := NormalizeEmployeeCode(bad)
		assert.ErrorIs(t, err, ErrValidation, bad)
	}
}

func TestValidateNewPassword(t *testing.T) {
	assert.NoError(t, ValidateNewPassword("p", "abcd1234"))
	assert.NoError(t, ValidateNewPassword("p", "รหัสผ่าน12"), "non-Latin letters count")
	for _, bad := range []string{"abc123", "abcdefgh", "12345678", strings.Repeat("a1", 37)} {
		assert.ErrorIs(t, ValidateNewPassword("p", bad), ErrValidation, bad)
	}
}

func TestNormalizeNewAccount(t *testing.T) {
	a, err := NormalizeNewAccount("emp-9", " New ", "Person", "")
	require.NoError(t, err)
	assert.Equal(t, NewAccount{EmployeeCode: "EMP-9", FirstName: "New", LastName: "Person", Role: RoleEmployee}, a)

	_, err = NormalizeNewAccount("EMP-9", "A", "B", "root")
	assert.ErrorIs(t, err, ErrValidation)
	_, err = NormalizeNewAccount("EMP-9", "", "B", "admin")
	assert.ErrorIs(t, err, ErrValidation)
}
