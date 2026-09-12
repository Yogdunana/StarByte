package service

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/Yogdunana/StarByte/backend/internal/backup/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service is the backup vertical cut (#88 phase-1).
type Service interface {
	List(ctx context.Context, req *dto.ListRequest) ([]dto.Record, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.Record, error)
	Create(ctx context.Context, userID uuid.UUID, req *dto.CreateRequest) (*dto.Record, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.RestoreRequest) (*dto.Record, error)
	GetPolicy(ctx context.Context) (*dto.Policy, error)
	UpdatePolicy(ctx context.Context, userID uuid.UUID, req *dto.UpdatePolicyRequest) (*dto.Policy, error)
	Storage(ctx context.Context) (*dto.StorageStats, error)
	RunScheduled(ctx context.Context, payload string, logf func(string)) error
	CleanupExpired(ctx context.Context, payload string, logf func(string)) error
	SyncSchedule(ctx context.Context) error
}

type backupService struct {
	rows   repo.Repository
	store  ArtifactStore
	engine Engine
	cfg    config.BackupConfig
	alert  Alerter
	now    func() time.Time
	run    func(func(context.Context))
	mu     sync.Mutex
	busy   bool
}

// New builds the production service. objectStore may be nil when local fallback is set.
func New(rows repo.Repository, objectStore storage.ObjectStorage, db config.DatabaseConfig, cfg config.BackupConfig, alert Alerter) Service {
	if alert == nil {
		alert = noopAlerter{}
	}
	cfg = withBackupDefaults(cfg)
	return &backupService{
		rows:   rows,
		store:  newArtifactStore(objectStore, cfg.LocalPath),
		engine: newPGEngine(db, cfg.PgDumpBin, cfg.PgRestoreBin),
		cfg:    cfg,
		alert:  alert,
		now:    time.Now,
		run:    goRun,
	}
}

func withBackupDefaults(cfg config.BackupConfig) config.BackupConfig {
	if strings.TrimSpace(cfg.Prefix) == "" {
		cfg.Prefix = "backups"
	}
	if strings.TrimSpace(cfg.PgDumpBin) == "" {
		cfg.PgDumpBin = "pg_dump"
	}
	if strings.TrimSpace(cfg.PgRestoreBin) == "" {
		cfg.PgRestoreBin = "psql"
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 1800
	}
	return cfg
}

func goRun(fn func(context.Context)) {
	go fn(context.Background())
}

func (s *backupService) timeout() time.Duration {
	return time.Duration(s.cfg.TimeoutSec) * time.Second
}

func (s *backupService) tryBegin() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return false
	}
	s.busy = true
	return true
}

func (s *backupService) end() {
	s.mu.Lock()
	s.busy = false
	s.mu.Unlock()
}

func (s *backupService) List(ctx context.Context, req *dto.ListRequest) ([]dto.Record, int64, error) {
	if req == nil {
		req = &dto.ListRequest{}
	}
	rows, total, err := s.rows.ListRecords(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	return toRecordDTOs(rows), total, nil
}

func (s *backupService) Get(ctx context.Context, id uuid.UUID) (*dto.Record, error) {
	rec, err := s.rows.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errNotFound()
	}
	return toRecordDTO(rec), nil
}

func (s *backupService) Create(ctx context.Context, userID uuid.UUID, _ *dto.CreateRequest) (*dto.Record, error) {
	return s.startBackup(ctx, userID, model.TriggerManual)
}

func (s *backupService) startBackup(ctx context.Context, userID uuid.UUID, trigger string) (*dto.Record, error) {
	busy, err := s.rows.CountBusy(ctx)
	if err != nil {
		return nil, err
	}
	if busy > 0 || !s.tryBegin() {
		return nil, errBusy()
	}
	now := s.now()
	rec := &model.Record{
		ID:            uuid.New(),
		TriggerSource: trigger,
		Status:        model.StatusPending,
		CreatedBy:     ptrUUID(userID),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.rows.CreateRecord(ctx, rec); err != nil {
		s.end()
		return nil, err
	}
	id := rec.ID
	s.run(func(parent context.Context) {
		defer s.end()
		jobCtx, cancel := context.WithTimeout(parent, s.timeout())
		defer cancel()
		s.executeDump(jobCtx, id)
	})
	return toRecordDTO(rec), nil
}

func (s *backupService) executeDump(ctx context.Context, id uuid.UUID) {
	rec, err := s.rows.GetRecord(ctx, id)
	if err != nil || rec == nil {
		logger.Error("backup record missing before dump", zap.String("id", id.String()), zap.Error(err))
		return
	}
	now := s.now()
	if err := applyTransition(rec, model.StatusRunning, now, ""); err != nil {
		logger.Error("backup cannot start", zap.Error(err))
		return
	}
	if err := s.rows.UpdateRecord(ctx, rec); err != nil {
		logger.Error("backup mark running failed", zap.Error(err))
		return
	}

	filename := fmt.Sprintf("starbyte-%s-%s.sql.gz", now.UTC().Format("20060102-150405"), rec.ID.String()[:8])
	key := objectKey(s.cfg.Prefix, rec.ID.String(), filename)
	tmp, err := os.CreateTemp("", "starbyte-backup-*.sql.gz")
	if err != nil {
		s.fail(ctx, rec, "创建临时文件失败: "+err.Error())
		return
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	hash := sha256.New()
	gz := gzip.NewWriter(io.MultiWriter(tmp, hash))
	if err := s.engine.Dump(ctx, gz); err != nil {
		_ = gz.Close()
		s.fail(ctx, rec, "pg_dump 失败: "+err.Error())
		return
	}
	if err := gz.Close(); err != nil {
		s.fail(ctx, rec, "gzip 失败: "+err.Error())
		return
	}
	info, err := tmp.Stat()
	if err != nil {
		s.fail(ctx, rec, "读取备份大小失败: "+err.Error())
		return
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	kind, err := s.store.Put(ctx, key, tmp, info.Size())
	if err != nil {
		s.fail(ctx, rec, "存储备份失败: "+err.Error())
		return
	}

	rec.Storage = kind
	rec.ObjectKey = key
	rec.Filename = filename
	rec.ChecksumSHA256 = hex.EncodeToString(hash.Sum(nil))
	rec.SizeBytes = info.Size()
	if err := applyTransition(rec, model.StatusSuccess, s.now(), ""); err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	if err := s.rows.UpdateRecord(ctx, rec); err != nil {
		logger.Error("backup mark success failed", zap.Error(err))
	}
	logger.Info("backup finished",
		zap.String("id", rec.ID.String()),
		zap.String("storage", rec.Storage),
		zap.Int64("bytes", rec.SizeBytes),
	)
}

func (s *backupService) fail(ctx context.Context, rec *model.Record, msg string) {
	logger.Error("backup failed", zap.String("id", rec.ID.String()), zap.String("error", msg))
	if err := applyTransition(rec, model.StatusFailed, s.now(), msg); err != nil {
		rec.Status = model.StatusFailed
		rec.ErrorMessage = msg
		now := s.now()
		rec.FinishedAt = &now
		rec.UpdatedAt = now
	}
	if err := s.rows.UpdateRecord(ctx, rec); err != nil {
		logger.Error("backup persist failure failed", zap.Error(err))
	}
	name := rec.Filename
	if name == "" {
		name = rec.ID.String()
	}
	s.alert.Failed(ctx, rec.CreatedBy, name, msg)
}

func (s *backupService) Delete(ctx context.Context, id uuid.UUID) error {
	rec, err := s.rows.GetRecord(ctx, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errNotFound()
	}
	if rec.Status == model.StatusPending || rec.Status == model.StatusRunning || rec.Status == model.StatusRestoring {
		return errInvalidState("进行中的任务不能删除")
	}
	if rec.ObjectKey != "" {
		if err := s.store.Delete(ctx, rec.Storage, rec.ObjectKey); err != nil {
			logger.Warn("backup artifact delete failed", zap.Error(err), zap.String("key", rec.ObjectKey))
		}
	}
	return s.rows.DeleteRecord(ctx, id)
}

func (s *backupService) Restore(ctx context.Context, _ uuid.UUID, id uuid.UUID, req *dto.RestoreRequest) (*dto.Record, error) {
	if err := validateRestoreConfirm(req); err != nil {
		return nil, err
	}
	rec, err := s.rows.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errNotFound()
	}
	if rec.Status != model.StatusSuccess && rec.Status != model.StatusRestored {
		return nil, errNotReady("只能从成功的备份恢复")
	}
	if rec.ObjectKey == "" || rec.ChecksumSHA256 == "" {
		return nil, errNotReady("备份产物不完整")
	}
	busy, err := s.rows.CountBusy(ctx)
	if err != nil {
		return nil, err
	}
	if busy > 0 || !s.tryBegin() {
		return nil, errBusy()
	}
	if err := applyTransition(rec, model.StatusRestoring, s.now(), ""); err != nil {
		s.end()
		return nil, err
	}
	if err := s.rows.UpdateRecord(ctx, rec); err != nil {
		s.end()
		return nil, err
	}
	idCopy := rec.ID
	s.run(func(parent context.Context) {
		defer s.end()
		jobCtx, cancel := context.WithTimeout(parent, s.timeout())
		defer cancel()
		s.executeRestore(jobCtx, idCopy)
	})
	return toRecordDTO(rec), nil
}

func (s *backupService) executeRestore(ctx context.Context, id uuid.UUID) {
	rec, err := s.rows.GetRecord(ctx, id)
	if err != nil || rec == nil {
		logger.Error("backup record missing before restore", zap.String("id", id.String()), zap.Error(err))
		return
	}
	src, err := s.store.Get(ctx, rec.Storage, rec.ObjectKey)
	if err != nil {
		s.fail(ctx, rec, "下载备份失败: "+err.Error())
		return
	}
	defer src.Close()

	tmp, err := os.CreateTemp("", "starbyte-restore-*.sql.gz")
	if err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, hash), src); err != nil {
		s.fail(ctx, rec, "读取备份失败: "+err.Error())
		return
	}
	if hex.EncodeToString(hash.Sum(nil)) != rec.ChecksumSHA256 {
		s.fail(ctx, rec, errChecksum().Error())
		return
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	gz, err := gzip.NewReader(tmp)
	if err != nil {
		s.fail(ctx, rec, "解压失败: "+err.Error())
		return
	}
	defer gz.Close()
	if err := s.engine.Restore(ctx, gz); err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	if err := applyTransition(rec, model.StatusRestored, s.now(), ""); err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	if err := s.rows.UpdateRecord(ctx, rec); err != nil {
		logger.Error("backup mark restored failed", zap.Error(err))
	}
	logger.Info("backup restore finished", zap.String("id", rec.ID.String()))
}

func (s *backupService) GetPolicy(ctx context.Context) (*dto.Policy, error) {
	p, err := s.ensurePolicy(ctx)
	if err != nil {
		return nil, err
	}
	return toPolicyDTO(p), nil
}

func (s *backupService) UpdatePolicy(ctx context.Context, userID uuid.UUID, req *dto.UpdatePolicyRequest) (*dto.Policy, error) {
	if req == nil {
		return nil, errPolicy("缺少策略字段")
	}
	p, err := s.ensurePolicy(ctx)
	if err != nil {
		return nil, err
	}
	if err := applyPolicyPatch(p, req); err != nil {
		return nil, err
	}
	p.UpdatedBy = ptrUUID(userID)
	p.UpdatedAt = s.now()
	if err := s.rows.UpsertPolicy(ctx, p); err != nil {
		return nil, err
	}
	if err := s.syncTask(ctx, p); err != nil {
		logger.Error("sync backup scheduler task failed", zap.Error(err))
	}
	return toPolicyDTO(p), nil
}

func (s *backupService) Storage(ctx context.Context) (*dto.StorageStats, error) {
	count, size, err := s.rows.StorageStats(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.StorageStats{
		Count:     count,
		SizeBytes: size,
		Prefix:    s.cfg.Prefix,
		Bucket:    s.cfg.Bucket,
		LocalPath: s.cfg.LocalPath,
	}, nil
}

func (s *backupService) RunScheduled(ctx context.Context, _ string, logf func(string)) error {
	if logf == nil {
		logf = func(string) {}
	}
	s.failStale(ctx)
	p, err := s.ensurePolicy(ctx)
	if err != nil {
		return err
	}
	if !p.Enabled {
		logf("policy disabled, skip")
		return nil
	}
	rec, err := s.startBackup(ctx, uuid.Nil, model.TriggerScheduled)
	if err != nil {
		return err
	}
	logf("queued backup " + rec.ID)
	s.waitUntilSettled(ctx, rec.ID)
	latest, err := s.rows.GetRecord(ctx, mustParse(rec.ID))
	if err != nil {
		return err
	}
	if latest != nil && latest.Status == model.StatusFailed {
		return errDump(latest.ErrorMessage)
	}
	return nil
}

func (s *backupService) CleanupExpired(ctx context.Context, _ string, logf func(string)) error {
	if logf == nil {
		logf = func(string) {}
	}
	s.failStale(ctx)
	p, err := s.ensurePolicy(ctx)
	if err != nil {
		return err
	}
	before := RetentionCutoff(s.now(), p.RetentionDays)
	rows, err := s.rows.ListExpired(ctx, before)
	if err != nil {
		return err
	}
	removed := 0
	for i := range rows {
		if err := s.Delete(ctx, rows[i].ID); err != nil {
			logger.Warn("backup retention delete failed", zap.Error(err), zap.String("id", rows[i].ID.String()))
			continue
		}
		removed++
	}
	logf(fmt.Sprintf("removed %d expired backups (retention %d days)", removed, p.RetentionDays))
	return nil
}

func (s *backupService) SyncSchedule(ctx context.Context) error {
	p, err := s.ensurePolicy(ctx)
	if err != nil {
		return err
	}
	return s.syncTask(ctx, p)
}

func (s *backupService) syncTask(ctx context.Context, p *model.Policy) error {
	return s.rows.SyncScheduledTask(ctx, repo.ScheduledTaskSpec{
		Code:       model.ScheduledTaskCode,
		Name:       "数据库全量备份",
		CronExpr:   p.CronExpr,
		Timezone:   p.Timezone,
		HandlerKey: model.ScheduledTaskCode,
		Enabled:    p.Enabled,
		TimeoutSec: s.cfg.TimeoutSec,
		NextRunAt:  nextRunAt(p.CronExpr, p.Timezone, s.now()),
	})
}

func (s *backupService) ensurePolicy(ctx context.Context) (*model.Policy, error) {
	p, err := s.rows.GetPolicy(ctx)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}
	now := s.now()
	p = &model.Policy{
		ID:            uuid.MustParse("00000000-0000-4000-8000-000000000060"),
		Enabled:       false,
		RetentionDays: model.DefaultRetentionDays,
		CronExpr:      model.DefaultCronExpr,
		Timezone:      model.DefaultTimezone,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.rows.UpsertPolicy(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *backupService) failStale(ctx context.Context) {
	before := s.now().Add(-s.timeout())
	rows, err := s.rows.ListStale(ctx, before)
	if err != nil {
		return
	}
	for i := range rows {
		s.fail(ctx, &rows[i], "任务超时或进程中断")
	}
}

func (s *backupService) waitUntilSettled(ctx context.Context, id string) {
	uid := mustParse(id)
	deadline := time.Now().Add(s.timeout() + 2*time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(50 * time.Millisecond):
		}
		rec, err := s.rows.GetRecord(ctx, uid)
		if err != nil || rec == nil {
			return
		}
		if rec.Status == model.StatusSuccess || rec.Status == model.StatusFailed {
			return
		}
	}
}

func mustParse(id string) uuid.UUID {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return u
}
