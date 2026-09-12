package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreToRejectsConninfoDBName(t *testing.T) {
	e := newPGEngine(config.DatabaseConfig{Host: "postgres", User: "starbyte", DBName: "starbyte"}, "false", "false")
	err := e.RestoreTo(context.Background(), bytes.NewReader([]byte("dump")), config.DatabaseConfig{
		Host: "postgres", User: "starbyte", DBName: "host=attacker.example dbname=stolen",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能是 DSN 或 conninfo")

	err = e.RestoreTo(context.Background(), bytes.NewReader([]byte("dump")), config.DatabaseConfig{
		Host: "postgres", User: "starbyte", DBName: "dbname=starbyte",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能是 DSN 或 conninfo")
}
