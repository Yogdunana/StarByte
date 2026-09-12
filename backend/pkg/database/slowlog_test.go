package database

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestSlowQueryLogger_LogModeAndTraceFast(t *testing.T) {
	l := newSlowQueryLogger()
	next := l.LogMode(gormlogger.Info)
	assert.NotNil(t, next)
	l.Info(context.Background(), "hi %s", "x")
	l.Warn(context.Background(), "w")
	l.Error(context.Background(), "e")
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)
}

func TestSlowQueryLogger_TraceSlowAndError(t *testing.T) {
	l := newSlowQueryLogger()
	l.Trace(context.Background(), time.Now().Add(-time.Second), func() (string, int64) {
		return "SELECT * FROM users", 10
	}, nil)
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT fail", 0
	}, errors.New("boom"))
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT missing", 0
	}, gorm.ErrRecordNotFound)
}

func TestSlowQueryLogger_TraceSlowWithRequestID(t *testing.T) {
	l := newSlowQueryLogger()
	ctx := logger.WithRequestID(context.Background(), "slow-req-1")
	l.Trace(ctx, time.Now().Add(-time.Second), func() (string, int64) {
		return "SELECT * FROM slow", 3
	}, nil)
}

func TestSlowQueryLogger_SilentSkipsTrace(t *testing.T) {
	l := newSlowQueryLogger().LogMode(gormlogger.Silent).(*slowQueryLogger)
	l.Trace(context.Background(), time.Now().Add(-time.Second), func() (string, int64) {
		t.Fatal("fc should not run")
		return "", 0
	}, nil)
}

func TestSanitizeSQLAndRing(t *testing.T) {
	ResetSlowQueriesForTest()
	t.Cleanup(ResetSlowQueriesForTest)

	assert.Equal(t, "SELECT 1", SanitizeSQL("  SELECT   1  "))
	long := strings.Repeat("x", 500)
	got := SanitizeSQL(long)
	assert.True(t, strings.HasSuffix(got, "…"))
	assert.Equal(t, 401, len([]rune(got)))

	assert.Nil(t, RecentSlowQueries(10))

	l := newSlowQueryLogger()
	l.Trace(context.Background(), time.Now().Add(-time.Second), func() (string, int64) {
		return "SELECT * FROM members WHERE secret='x'", 4
	}, nil)
	items := RecentSlowQueries(5)
	require.Len(t, items, 1)
	assert.Contains(t, items[0].SQL, "SELECT * FROM members")
	assert.GreaterOrEqual(t, items[0].DurationMs, int64(500))
}

func TestSlowQueryRingNewestFirst(t *testing.T) {
	r := &slowQueryRing{items: make([]RecordedSlowQuery, 3)}
	r.push(RecordedSlowQuery{SQL: "a"})
	r.push(RecordedSlowQuery{SQL: "b"})
	r.push(RecordedSlowQuery{SQL: "c"})
	r.push(RecordedSlowQuery{SQL: "d"})
	got := r.recent(3)
	require.Len(t, got, 3)
	assert.Equal(t, []string{"d", "c", "b"}, []string{got[0].SQL, got[1].SQL, got[2].SQL})
	assert.Nil(t, r.recent(0))
}
