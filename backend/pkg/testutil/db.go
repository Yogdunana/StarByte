package testutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func postgresCandidates() []string {
	return []string{
		os.Getenv("TEST_DATABASE_URL"),
		os.Getenv("STATS_TEST_DSN"),
		"host=localhost user=starbyte password=starbyte dbname=starbyte_test port=5432 sslmode=disable",
		"host=localhost user=starbyte password=starbyte dbname=starbyte_dev port=5432 sslmode=disable",
		"host=localhost user=postgres password=postgres dbname=starbyte_test port=5432 sslmode=disable",
	}
}

func connectPostgres() (*gorm.DB, error) {
	var last error
	for _, dsn := range postgresCandidates() {
		if dsn == "" {
			continue
		}
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			last = err
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			last = err
			continue
		}
		if err := sqlDB.Ping(); err != nil {
			last = err
			_ = sqlDB.Close()
			continue
		}
		return db, nil
	}
	if last == nil {
		last = fmt.Errorf("no PostgreSQL DSN candidates")
	}
	return nil, last
}

// OpenPostgres connects to a test PostgreSQL.
// Skips when none is reachable, unless TEST_DATABASE_REQUIRED=1 (then the test fails).
func OpenPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := connectPostgres()
	if err != nil {
		if os.Getenv("TEST_DATABASE_REQUIRED") == "1" {
			require.NoError(t, err, "PostgreSQL is required (TEST_DATABASE_REQUIRED=1)")
		}
		t.Skipf("skipping DB test: no PostgreSQL (%v)", err)
	}
	return db
}

// MustOpenPostgres fails the test when PostgreSQL is unreachable.
func MustOpenPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := connectPostgres()
	require.NoError(t, err, "PostgreSQL is required")
	return db
}
