package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePasswordStrength_Valid(t *testing.T) {
	valid := []string{
		"password123",
		"abc12345",
		"Aa1b2c3d",
		"test1234",
	}
	for _, pwd := range valid {
		assert.True(t, ValidatePasswordStrength(pwd), "password should be valid: %s", pwd)
	}
}

func TestValidatePasswordStrength_TooShort(t *testing.T) {
	short := []string{"", "a", "ab1", "1234567"}
	for _, pwd := range short {
		assert.False(t, ValidatePasswordStrength(pwd), "password should be rejected (too short): %s", pwd)
	}
}

func TestValidatePasswordStrength_NoDigits(t *testing.T) {
	assert.False(t, ValidatePasswordStrength("abcdefgh"))
	assert.False(t, ValidatePasswordStrength("onlyletters"))
}

func TestValidatePasswordStrength_NoLetters(t *testing.T) {
	assert.False(t, ValidatePasswordStrength("12345678"))
	assert.False(t, ValidatePasswordStrength("00000000"))
}

func TestHashPassword_AndCheckPassword(t *testing.T) {
	password := "testpass123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Correct password
	assert.True(t, CheckPassword(password, hash))
	// Wrong password
	assert.False(t, CheckPassword("wrongpass", hash))
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "samepassword123"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)
	// bcrypt generates different hashes due to random salt
	assert.NotEqual(t, hash1, hash2)
}
