package database

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestIsUniqueViolation(t *testing.T) {
	assert.False(t, IsUniqueViolation(nil))
	assert.True(t, IsUniqueViolation(gorm.ErrDuplicatedKey))
	assert.True(t, IsUniqueViolation(fmt.Errorf("wrap: %w", gorm.ErrDuplicatedKey)))
	assert.True(t, IsUniqueViolation(&pgconn.PgError{Code: "23505"}))
	assert.False(t, IsUniqueViolation(&pgconn.PgError{Code: "23503"}))
}

func TestIsForeignKeyViolation(t *testing.T) {
	assert.False(t, IsForeignKeyViolation(nil))
	assert.True(t, IsForeignKeyViolation(&pgconn.PgError{Code: "23503"}))
	assert.False(t, IsForeignKeyViolation(&pgconn.PgError{Code: "23505"}))
}
