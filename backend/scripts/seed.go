package main

import (
	"fmt"
	"os"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		if _, err := os.Stat("configs/config.yaml"); err == nil {
			cfgPath = "configs/config.yaml"
		} else {
			cfgPath = "backend/configs/config.yaml"
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	if err := logger.Init(&cfg.Logger); err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("init database failed", zap.Error(err))
	}
	defer func() { _ = database.Close() }()

	if err := SeedAll(database.DB()); err != nil {
		logger.Fatal("seed failed", zap.Error(err))
	}
	logger.Info("seed completed")
}

// SeedAll 幂等写入角色、权限、组织、测试用户和通知模板
func SeedAll(db *gorm.DB) error {
	if err := seedRoles(db); err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}
	if err := seedPermissions(db); err != nil {
		return fmt.Errorf("seed permissions: %w", err)
	}
	if err := seedDepartments(db); err != nil {
		return fmt.Errorf("seed departments: %w", err)
	}
	if err := seedPositions(db); err != nil {
		return fmt.Errorf("seed positions: %w", err)
	}
	if err := seedUsers(db); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedRolePermissions(db); err != nil {
		return fmt.Errorf("seed role permissions: %w", err)
	}
	if err := seedTemplates(db); err != nil {
		return fmt.Errorf("seed templates: %w", err)
	}
	return nil
}
