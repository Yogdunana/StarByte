package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID returns a new random UUID as a string.
func GenerateUUID() string {
	return uuid.New().String()
}
