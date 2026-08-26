package logger

import (
	"os"
	"sync"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log *zap.Logger
	mu  sync.Mutex
)

func init() {
	// Provide a no-op logger so the helper functions are safe to call before
	// (or without) Init.
	log = zap.NewNop()
}

// Init creates a zap logger that writes JSON-encoded records to both stdout
// and a rotating log file managed by lumberjack.
func Init(cfg *config.LoggerConfig) error {
	level := parseLevel(cfg.Level)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(encoder, fileWriter, level),
	)

	mu.Lock()
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	mu.Unlock()

	return nil
}

func parseLevel(s string) zapcore.Level {
	switch s {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Sync flushes any buffered log entries.
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

// GetLogger returns the underlying *zap.Logger instance.
// Useful for components that need direct access to the zap logger.
func GetLogger() *zap.Logger {
	return log
}

// Info logs a message at the info level.
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Error logs a message at the error level.
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Warn logs a message at the warn level.
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Fatal logs a message at the fatal level and then terminates the process.
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}
