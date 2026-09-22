package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// -mode 默认 seed：不带任何参数运行时行为与改造前完全一致（make seed 走这条路）。
	// 其余模式（show-admin / set-admin-password / set-admin-username）给运维 CLI 用，
	// 见 admin.go。
	mode := flag.String("mode", adminModeSeed, "seed（默认，全量幂等种子）| show-admin | set-admin-password | set-admin-username")
	username := flag.String("username", "admin", "管理员账号名（维护模式的目标账号）")
	value := flag.String("value", "", "新账号名；口令留空则按 环境变量>stdin>自动生成 的顺序取")
	flag.Parse()

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

	if *mode != adminModeSeed {
		// 维护模式：不跑整批种子，只动管理员账号。
		// 出错只写 stderr 并以非零码退出，CLI 据此判断成功与否。
		if err := runAdminMode(database.DB(), cfg, *mode, *username, *value); err != nil {
			fmt.Fprintf(os.Stderr, "starbyte-seed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := SeedAll(database.DB()); err != nil {
		logger.Fatal("seed failed", zap.Error(err))
	}
	logger.Info("seed completed")
}

// SeedAll 幂等写入角色、权限、组织、测试用户和通知模板（整批事务，失败回滚）
func SeedAll(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedRoles(tx); err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}
		if err := seedPermissions(tx); err != nil {
			return fmt.Errorf("seed permissions: %w", err)
		}
		if err := seedDepartments(tx); err != nil {
			return fmt.Errorf("seed departments: %w", err)
		}
		if err := seedPositions(tx); err != nil {
			return fmt.Errorf("seed positions: %w", err)
		}
		if err := seedUsers(tx); err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
		if err := seedMemberProfiles(tx); err != nil {
			return fmt.Errorf("seed member profiles: %w", err)
		}
		if err := seedRolePermissions(tx); err != nil {
			return fmt.Errorf("seed role permissions: %w", err)
		}
		if err := seedTemplates(tx); err != nil {
			return fmt.Errorf("seed templates: %w", err)
		}
		if err := seedInternships(tx); err != nil {
			return fmt.Errorf("seed internships: %w", err)
		}
		if err := seedRuntimeConfigs(tx); err != nil {
			return fmt.Errorf("seed runtime configs: %w", err)
		}
		if err := seedDicts(tx); err != nil {
			return fmt.Errorf("seed dicts: %w", err)
		}
		if err := seedOps(tx); err != nil {
			return fmt.Errorf("seed ops: %w", err)
		}
		if err := seedKnowledge(tx); err != nil {
			return fmt.Errorf("seed knowledge: %w", err)
		}
		return nil
	})
}
