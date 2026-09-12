package database

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	slowQueryThreshold = 500 * time.Millisecond
	slowQueryRingCap   = 64
	slowQuerySQLMax    = 400
)

// RecordedSlowQuery is a GORM-detected slow statement (no host credentials).
type RecordedSlowQuery struct {
	SQL        string
	DurationMs int64
	Rows       int64
	At         time.Time
}

type slowQueryRing struct {
	mu    sync.Mutex
	items []RecordedSlowQuery
	next  int
	n     int
}

var gormSlowRing = &slowQueryRing{items: make([]RecordedSlowQuery, slowQueryRingCap)}

// slowQueryLogger logs SQL that exceeds slowQueryThreshold.
type slowQueryLogger struct {
	level gormlogger.LogLevel
}

func newSlowQueryLogger() *slowQueryLogger {
	return &slowQueryLogger{level: gormlogger.Warn}
}

func (l *slowQueryLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	cp := *l
	cp.level = level
	return &cp
}

func (l *slowQueryLogger) Info(_ context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Info {
		logger.GetLogger().Sugar().Infof(msg, args...)
	}
}

func (l *slowQueryLogger) Warn(_ context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Warn {
		logger.GetLogger().Sugar().Warnf(msg, args...)
	}
}

func (l *slowQueryLogger) Error(_ context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Error {
		logger.GetLogger().Sugar().Errorf(msg, args...)
	}
}

func (l *slowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= gormlogger.Error {
		logger.Error("sql error",
			zap.Error(err),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("duration", elapsed),
		)
		return
	}
	if elapsed < slowQueryThreshold {
		return
	}
	recordGormSlowQuery(sql, elapsed, rows)
	reqID := logger.RequestIDFrom(ctx)
	logger.Warn("slow query detected",
		zap.String("sql", sql),
		zap.Int64("duration_ms", elapsed.Milliseconds()),
		zap.Int64("rows", rows),
		zap.String("request_id", reqID),
	)
}

func recordGormSlowQuery(sql string, elapsed time.Duration, rows int64) {
	item := RecordedSlowQuery{
		SQL:        SanitizeSQL(sql),
		DurationMs: elapsed.Milliseconds(),
		Rows:       rows,
		At:         time.Now().UTC(),
	}
	gormSlowRing.push(item)
}

func (r *slowQueryRing) push(item RecordedSlowQuery) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) == 0 {
		return
	}
	r.items[r.next] = item
	r.next = (r.next + 1) % len(r.items)
	if r.n < len(r.items) {
		r.n++
	}
}

// RecentSlowQueries returns newest-first GORM slow statements (capped).
func RecentSlowQueries(limit int) []RecordedSlowQuery {
	return gormSlowRing.recent(limit)
}

// ResetSlowQueriesForTest clears the process-wide ring. Tests only.
func ResetSlowQueriesForTest() {
	gormSlowRing.mu.Lock()
	defer gormSlowRing.mu.Unlock()
	gormSlowRing.next = 0
	gormSlowRing.n = 0
}

// RecordSlowQueryForTest appends one statement to the ring. Tests only.
func RecordSlowQueryForTest(sql string, d time.Duration, rows int64) {
	recordGormSlowQuery(sql, d, rows)
}

func (r *slowQueryRing) recent(limit int) []RecordedSlowQuery {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.n == 0 || limit <= 0 {
		return nil
	}
	if limit > r.n {
		limit = r.n
	}
	out := make([]RecordedSlowQuery, 0, limit)
	// newest is just before next
	idx := r.next
	for i := 0; i < limit; i++ {
		idx--
		if idx < 0 {
			idx = len(r.items) - 1
		}
		out = append(out, r.items[idx])
	}
	return out
}

// SanitizeSQL collapses whitespace and truncates so the API never returns
// unbounded SQL text (may still contain application literals).
func SanitizeSQL(sql string) string {
	sql = strings.Join(strings.Fields(sql), " ")
	if sql == "" {
		return ""
	}
	if utf8.RuneCountInString(sql) <= slowQuerySQLMax {
		return sql
	}
	runes := []rune(sql)
	return string(runes[:slowQuerySQLMax]) + "…"
}
