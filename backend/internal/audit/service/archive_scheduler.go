package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"go.uber.org/zap"
)

// ArchiveScheduler 审计日志归档定时任务调度器
// 每天 02:00 执行一次归档操作，将 90 天前的日志归档到 MinIO
type ArchiveScheduler struct {
	auditService AuditService
	stopCh       chan struct{}
	done         chan struct{}
}

// NewArchiveScheduler 创建归档调度器
func NewArchiveScheduler(auditService AuditService) *ArchiveScheduler {
	return &ArchiveScheduler{
		auditService: auditService,
		stopCh:       make(chan struct{}),
		done:         make(chan struct{}),
	}
}

// Start 启动归档调度器
func (s *ArchiveScheduler) Start() {
	go s.run()
	logger.Info("audit log archive scheduler started")
}

// Stop 停止归档调度器
func (s *ArchiveScheduler) Stop() {
	close(s.stopCh)
	<-s.done
	logger.Info("audit log archive scheduler stopped")
}

// run 调度循环
func (s *ArchiveScheduler) run() {
	defer close(s.done)

	for {
		// 计算到下次 02:00 的等待时间
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
		if now.After(next) || now.Equal(next) {
			next = next.Add(24 * time.Hour)
		}
		waitDuration := next.Sub(now)

		logger.Info("audit archive: next run scheduled",
			zap.Time("next_run", next),
			zap.Duration("wait", waitDuration),
		)

		select {
		case <-time.After(waitDuration):
			s.executeArchive()
		case <-s.stopCh:
			return
		}
	}
}

// executeArchive 执行归档操作
func (s *ArchiveScheduler) executeArchive() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("audit archive: starting scheduled archive job")

	result, err := s.auditService.Archive(ctx, 90)
	if err != nil {
		logger.Error("audit archive: scheduled archive failed", zap.Error(err))
		return
	}

	logger.Info("audit archive: scheduled archive completed",
		zap.String("archive_date", result.ArchiveDate),
		zap.Int64("record_count", result.RecordCount),
		zap.String("minio_object", result.MinIOObject),
	)
}
