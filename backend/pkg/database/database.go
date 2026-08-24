package database

import (
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

// Init opens a GORM connection to PostgreSQL using the provided configuration
// and configures the underlying connection pool.
func Init(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	if cfg.MaxOpen > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	}
	// A reasonable default connection max lifetime so stale connections are
	// recycled even when the config does not specify one explicitly.
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

// Close releases the underlying database connection pool.
func Close() error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DB returns the global *gorm.DB instance. It returns nil if Init has not been
// called successfully.
func DB() *gorm.DB {
	return db
}
