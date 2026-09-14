package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func postgresCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// IsUniqueViolation reports PostgreSQL 23505 / GORM duplicated-key errors.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return postgresCode(err) == "23505"
}

// IsForeignKeyViolation reports PostgreSQL 23503.
func IsForeignKeyViolation(err error) bool {
	return postgresCode(err) == "23503"
}
