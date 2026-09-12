package main

import (
	"context"
	"fmt"
	"os"

	backupcli "github.com/Yogdunana/StarByte/backend/internal/backup/cli"
	backupRepo "github.com/Yogdunana/StarByte/backend/internal/backup/repo"
	backupService "github.com/Yogdunana/StarByte/backend/internal/backup/service"
	notifRepo "github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	notifService "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"go.uber.org/zap"
)

func runBackupCLI(cfg *config.Config, args []string) error {
	if err := database.Init(&cfg.Database); err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	defer func() { _ = database.Close() }()

	var objectStore storage.ObjectStorage
	if store, err := storage.NewMinIO(cfg.MinIO); err != nil {
		logger.Warn("backup CLI MinIO unavailable, local path may still work", zap.Error(err))
	} else {
		objectStore = store
	}
	backupMinioCfg := cfg.MinIO
	if bucket := cfg.Backup.Bucket; bucket != "" {
		backupMinioCfg.Bucket = bucket
	}
	if cfg.Backup.Bucket == "" {
		cfg.Backup.Bucket = cfg.MinIO.Bucket
	}
	backupStore := objectStore
	if backupMinioCfg.Bucket != cfg.MinIO.Bucket {
		if dedicated, err := storage.NewMinIO(backupMinioCfg); err == nil {
			backupStore = dedicated
		}
	}

	notifR := notifRepo.NewNotificationRepo(database.DB())
	tplRepo := notifRepo.NewTemplateRepo(database.DB())
	channels := notifService.NewChannelRegistry()
	channels.Register(notifService.NewInAppChannel(notifR))
	channels.Register(notifService.NewEmailChannelFromConfig(cfg.Email))
	notifSvc := notifService.NewNotificationService(
		notifR, tplRepo, notifService.NewTemplateEngine(tplRepo), channels,
	)
	backupRows := backupRepo.New(database.DB())
	svc := backupService.New(
		backupRows,
		backupStore,
		cfg.Database,
		cfg.Backup,
		backupService.NewNotifier(notifSvc, backupRows),
	)
	if code := backupcli.Execute(context.Background(), svc, args, os.Stdout, os.Stderr); code != 0 {
		return fmt.Errorf("backup CLI exited %d", code)
	}
	return nil
}
