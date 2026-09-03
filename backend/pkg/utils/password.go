package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the bcrypt cost factor. Issue #17 requires cost=12.
const bcryptCost = 12

// HashPassword returns a bcrypt hash of the given plain-text password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword reports whether the plain-text password matches the bcrypt hash.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength checks that a password meets the minimum strength
// requirements: at least 8 characters, containing both letters and numbers.
func ValidatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for _, ch := range password {
		switch {
		case (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'):
			hasLetter = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
