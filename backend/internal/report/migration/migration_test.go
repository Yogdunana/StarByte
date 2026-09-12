package migration

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	upMigrationName   = "000069_report_read_permission.up.sql"
	downMigrationName = "000069_report_read_permission.down.sql"
)

func TestReportReadPermissionMigrationContract(t *testing.T) {
	up := strings.ToLower(readMigration(t, upMigrationName))
	down := strings.ToLower(readMigration(t, downMigrationName))

	assert.Contains(t, up, "insert into permissions")
	assert.Contains(t, up, "'report:read'")
	assert.Contains(t, up, "'report'")
	assert.Contains(t, up, "'read'")
	assert.NotContains(t, up, "role_permissions")
	assert.NotContains(t, up, "report:review")

	assert.Contains(t, down, "delete from permissions")
	assert.Equal(t, 1, strings.Count(down, "'report:read'"))
	assert.NotContains(t, down, "role_permissions")
	assert.NotContains(t, down, "report:review")
}

func TestReportReadPermissionMigrationPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run migration integration test")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	rawDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, rawDB.Close()) })

	up := readMigration(t, upMigrationName)
	down := readMigration(t, downMigrationName)
	rollback := errors.New("rollback report permission migration test")
	err = db.Transaction(func(tx *gorm.DB) error {
		var initialOtherPermissionCount int64
		require.NoError(t, tx.Table("permissions").Where("code = ?", "user:read").Count(&initialOtherPermissionCount).Error)
		require.Greater(t, initialOtherPermissionCount, int64(0))

		assertPermission(t, tx)
		assertNoRolePermissions(t, tx)

		require.NoError(t, tx.Exec(down).Error)
		var reportPermissionCount int64
		require.NoError(t, tx.Table("permissions").Where("code = ?", "report:read").Count(&reportPermissionCount).Error)
		assert.Zero(t, reportPermissionCount)

		var otherPermissionCount int64
		require.NoError(t, tx.Table("permissions").Where("code = ?", "user:read").Count(&otherPermissionCount).Error)
		assert.Equal(t, initialOtherPermissionCount, otherPermissionCount)

		require.NoError(t, tx.Exec(up).Error)
		assertPermission(t, tx)
		assertNoRolePermissions(t, tx)
		return rollback
	})
	require.ErrorIs(t, err, rollback)
}

func assertPermission(t *testing.T, db *gorm.DB) {
	t.Helper()
	var row struct {
		Name        string
		Code        string
		Resource    string
		Action      string
		Description string
		Type        int16
		IsSystem    bool
		Status      int16
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}
	require.NoError(t, db.Table("permissions").Where("code = ?", "report:read").Take(&row).Error)
	assert.Equal(t, "工作汇报查看", row.Name)
	assert.Equal(t, "report:read", row.Code)
	assert.Equal(t, "report", row.Resource)
	assert.Equal(t, "read", row.Action)
	assert.Equal(t, "查看工作汇报列表", row.Description)
	assert.Equal(t, int16(3), row.Type)
	assert.True(t, row.IsSystem)
	assert.Zero(t, row.Status)
	assert.False(t, row.CreatedAt.IsZero())
	assert.False(t, row.UpdatedAt.IsZero())
}

func assertNoRolePermissions(t *testing.T, db *gorm.DB) {
	t.Helper()
	var count int64
	err := db.Table("role_permissions AS rp").
		Joins("JOIN permissions AS p ON p.id = rp.permission_id").
		Where("p.code = ?", "report:read").
		Count(&count).Error
	require.NoError(t, err)
	assert.Zero(t, count)
}

func readMigration(t *testing.T, name string) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	path := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "migrations", name)
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(contents)
}
