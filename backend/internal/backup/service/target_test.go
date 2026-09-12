package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePostgresTarget(t *testing.T) {
	urlCfg, err := ParsePostgresTarget("postgres://alice:s3cret@db.example:6543/starbyte_drill?sslmode=require")
	require.NoError(t, err)
	assert.Equal(t, "db.example", urlCfg.Host)
	assert.Equal(t, 6543, urlCfg.Port)
	assert.Equal(t, "alice", urlCfg.User)
	assert.Equal(t, "s3cret", urlCfg.Password)
	assert.Equal(t, "starbyte_drill", urlCfg.DBName)
	assert.Equal(t, "require", urlCfg.SSLMode)

	kv, err := ParsePostgresTarget("host=10.0.0.8 port=5432 user=ops dbname=drill sslmode=disable")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.8", kv.Host)
	assert.Equal(t, "drill", kv.DBName)
}

func TestResolveDrillTargetRejectsLiveDB(t *testing.T) {
	live := config.DatabaseConfig{Host: "postgres", Port: 5432, User: "starbyte", DBName: "starbyte"}
	_, err := ResolveDrillTarget(live, &dto.DrillRequest{TargetDBName: "starbyte"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能是当前应用库")

	got, err := ResolveDrillTarget(live, &dto.DrillRequest{TargetDBName: "starbyte_drill"})
	require.NoError(t, err)
	assert.Equal(t, "starbyte_drill", got.DBName)
	assert.Equal(t, "postgres", got.Host)
	assert.Equal(t, "starbyte", got.User)
}

func TestSameDatabaseTreatsLocalhostAliases(t *testing.T) {
	a := config.DatabaseConfig{Host: "127.0.0.1", Port: 0, DBName: "starbyte"}
	b := config.DatabaseConfig{Host: "localhost", Port: 5432, DBName: "starbyte"}
	assert.True(t, SameDatabase(a, b))
	b.DBName = "other"
	assert.False(t, SameDatabase(a, b))
}
