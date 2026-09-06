package database

import (
	"context"
	"errors"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const slowQueryThreshold = 500 * time.Millisecond

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
	reqID, _ := ctx.Value("request_id").(string)
	logger.Warn("slow query detected",
		zap.String("sql", sql),
		zap.Int64("duration_ms", elapsed.Milliseconds()),
		zap.Int64("rows", rows),
		zap.String("request_id", reqID),
	)
}
