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

// Service is the backup vertical cut (#88 phase-2: AES, preview, alerts).
type Service interface {
	List(ctx context.Context, req *dto.ListRequest) ([]dto.Record, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.Record, error)
	Create(ctx context.Context, userID uuid.UUID, req *dto.CreateRequest) (*dto.Record, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.RestoreRequest) (*dto.Record, error)
	DrillRestore(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.DrillRequest) (*dto.DrillResult, error)
	GetDrill(ctx context.Context, id uuid.UUID) (*dto.DrillResult, error)
	Wait(ctx context.Context, id uuid.UUID) (*dto.Record, error)
	GetPolicy(ctx context.Context) (*dto.Policy, error)
	UpdatePolicy(ctx context.Context, userID uuid.UUID, req *dto.UpdatePolicyRequest) (*dto.Policy, error)
	Storage(ctx context.Context) (*dto.StorageStats, error)
	Preview(ctx context.Context, id uuid.UUID) (*dto.Preview, error)
	RunScheduled(ctx context.Context, payload string, logf func(string)) error
	CleanupExpired(ctx context.Context, payload string, logf func(string)) error
	SyncSchedule(ctx context.Context) error
}

type backupService struct {
	rows    repo.Repository
	store   ArtifactStore
	engine  Engine
	live    config.DatabaseConfig
	cfg     config.BackupConfig
	alert   Alerter
	now     func() time.Time
	run     func(func(context.Context))
	mu      sync.Mutex
	busy    bool
	drillMu sync.Mutex
	drills  map[uuid.UUID]dto.DrillResult
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
		live:   db,
		cfg:    cfg,
		alert:  alert,
		now:    time.Now,
		run:    goRun,
		drills: make(map[uuid.UUID]dto.DrillResult),
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
		cfg.PgRestoreBin = "pg_restore"
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
		s.failByID(ctx, id, model.StatusFailed, "读取备份记录失败")
		return
	}
	now := s.now()
	if err := applyTransition(rec, model.StatusRunning, now, ""); err != nil {
		s.fail(ctx, rec, "无法开始备份: "+err.Error())
		return
	}
	if err := s.rows.UpdateRecord(ctx, rec); err != nil {
		s.fail(ctx, rec, "标记备份进行中失败: "+err.Error())
		return
	}

	encKey := parseEncryptionKey(s.cfg.EncryptionKey)
	filename := fmt.Sprintf("starbyte-%s-%s.dump.gz", now.UTC().Format("20060102-150405"), rec.ID.String()[:8])
	if len(encKey) > 0 {
		filename += ".enc"
	}
	key := objectKey(s.cfg.Prefix, rec.ID.String(), filename)
	plain, err := os.CreateTemp("", "starbyte-backup-*.dump.gz")
	if err != nil {
		s.fail(ctx, rec, "创建临时文件失败: "+err.Error())
		return
	}
	plainName := plain.Name()
	defer func() {
		_ = plain.Close()
		_ = os.Remove(plainName)
	}()

	gz := gzip.NewWriter(plain)
	if err := s.engine.Dump(ctx, gz); err != nil {
		_ = gz.Close()
		s.fail(ctx, rec, "pg_dump 失败: "+err.Error())
		return
	}
	if err := gz.Close(); err != nil {
		s.fail(ctx, rec, "gzip 失败: "+err.Error())
		return
	}
	if _, err := plain.Seek(0, io.SeekStart); err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}

	out, err := os.CreateTemp("", "starbyte-backup-*.store")
	if err != nil {
		s.fail(ctx, rec, "创建存储临时文件失败: "+err.Error())
		return
	}
	outName := out.Name()
	defer func() {
		_ = out.Close()
		_ = os.Remove(outName)
	}()
	hash := sha256.New()
	dest := io.MultiWriter(out, hash)
	encrypted := false
	if len(encKey) > 0 {
		if err := encryptStream(dest, plain, encKey); err != nil {
			s.fail(ctx, rec, "加密失败: "+err.Error())
			return
		}
		encrypted = true
	} else if _, err := io.Copy(dest, plain); err != nil {
		s.fail(ctx, rec, "写入备份失败: "+err.Error())
		return
	}
	info, err := out.Stat()
	if err != nil {
		s.fail(ctx, rec, "读取备份大小失败: "+err.Error())
		return
	}
	if _, err := out.Seek(0, io.SeekStart); err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	kind, err := s.store.Put(ctx, key, out, info.Size())
	if err != nil {
		s.fail(ctx, rec, errStore("存储备份失败: "+err.Error()).Error())
		return
	}

	rec.Storage = kind
	rec.ObjectKey = key
	rec.Filename = filename
	rec.ChecksumSHA256 = hex.EncodeToString(hash.Sum(nil))
	rec.SizeBytes = info.Size()
	rec.Encrypted = encrypted
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
	to := model.StatusFailed
	if rec.Status == model.StatusRestoring {
		to = model.StatusRestoreFailed
	}
	logger.Error("backup failed", zap.String("id", rec.ID.String()), zap.String("error", msg))
	if err := applyTransition(rec, to, s.now(), msg); err != nil {
		rec.Status = to
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

func (s *backupService) failByID(ctx context.Context, id uuid.UUID, status int16, msg string) {
	if err := s.rows.MarkTerminal(ctx, id, status, msg, s.now()); err != nil {
		logger.Error("backup mark terminal failed", zap.String("id", id.String()), zap.Error(err))
	}
	s.alert.Failed(ctx, nil, id.String(), msg)
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
	if rec.Status != model.StatusSuccess && rec.Status != model.StatusRestored && rec.Status != model.StatusRestoreFailed {
		return nil, errNotReady("只能从成功或可重试的备份恢复")
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
		s.failByID(ctx, id, model.StatusRestoreFailed, "读取备份记录失败")
		return
	}
	dump, _, err := s.downloadUnwrapped(ctx, rec)
	if err != nil {
		s.fail(ctx, rec, err.Error())
		return
	}
	defer func() { _ = dump.Close() }()
	if err := s.engine.Restore(ctx, dump); err != nil {
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

func (s *backupService) DrillRestore(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.DrillRequest) (*dto.DrillResult, error) {
	if err := validateDrillConfirm(req); err != nil {
		return nil, err
	}
	target, err := ResolveDrillTarget(s.liveDB(), req)
	if err != nil {
		return nil, err
	}
	rec, err := s.rows.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errNotFound()
	}
	if rec.Status != model.StatusSuccess && rec.Status != model.StatusRestored && rec.Status != model.StatusRestoreFailed {
		return nil, errNotReady("只能从成功或可重试的备份演练恢复")
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
	queued := &dto.DrillResult{
		ID:           rec.ID.String(),
		Filename:     rec.Filename,
		Queued:       true,
		Status:       "queued",
		TargetHost:   target.Host,
		TargetPort:   normPort(target.Port),
		TargetDBName: target.DBName,
	}
	s.storeDrill(queued)
	recCopy := *rec
	s.run(func(parent context.Context) {
		defer s.end()
		jobCtx, cancel := context.WithTimeout(parent, s.timeout())
		defer cancel()
		s.executeDrill(jobCtx, userID, &recCopy, target)
	})
	return s.GetDrill(ctx, rec.ID)
}

func (s *backupService) executeDrill(ctx context.Context, userID uuid.UUID, rec *model.Record, target config.DatabaseConfig) {
	out := &dto.DrillResult{
		ID:           rec.ID.String(),
		Filename:     rec.Filename,
		Queued:       true,
		Ready:        true,
		Status:       "running",
		TargetHost:   target.Host,
		TargetPort:   normPort(target.Port),
		TargetDBName: target.DBName,
	}
	s.storeDrill(out)

	latest, err := s.rows.GetRecord(ctx, rec.ID)
	if err != nil || latest == nil {
		s.finishDrill(ctx, out, userID, "读取备份记录失败")
		return
	}
	dump, _, err := s.downloadUnwrapped(ctx, latest)
	if err != nil {
		s.finishDrill(ctx, out, userID, err.Error())
		return
	}
	defer func() { _ = dump.Close() }()
	if err := s.engine.RestoreTo(ctx, dump, target); err != nil {
		s.finishDrill(ctx, out, userID, err.Error())
		return
	}
	out.Queued = false
	out.Restored = true
	out.Status = "restored"
	s.storeDrill(out)
	logger.Info("backup drill restore finished",
		zap.String("id", rec.ID.String()),
		zap.String("target_db", target.DBName),
		zap.String("target_host", target.Host),
	)
}

func (s *backupService) finishDrill(ctx context.Context, out *dto.DrillResult, userID uuid.UUID, msg string) {
	out.Queued = false
	out.Ready = false
	out.Restored = false
	out.Status = "failed"
	out.Error = msg
	s.storeDrill(out)
	alertCtx, cancel := alertContext(ctx)
	defer cancel()
	s.alert.Failed(alertCtx, ptrUUID(userID), out.Filename, "演练: "+msg)
}

func (s *backupService) storeDrill(out *dto.DrillResult) {
	if out == nil {
		return
	}
	id, err := uuid.Parse(out.ID)
	if err != nil {
		return
	}
	cp := *out
	s.drillMu.Lock()
	if s.drills == nil {
		s.drills = make(map[uuid.UUID]dto.DrillResult)
	}
	s.drills[id] = cp
	s.drillMu.Unlock()
}

func (s *backupService) GetDrill(_ context.Context, id uuid.UUID) (*dto.DrillResult, error) {
	s.drillMu.Lock()
	defer s.drillMu.Unlock()
	got, ok := s.drills[id]
	if !ok {
		return nil, errNotFound()
	}
	cp := got
	return &cp, nil
}

func (s *backupService) Wait(ctx context.Context, id uuid.UUID) (*dto.Record, error) {
	s.waitUntilSettled(ctx, id.String())
	return s.Get(ctx, id)
}

func (s *backupService) liveDB() config.DatabaseConfig {
	return s.live
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
		return nil, errPolicy("调度任务同步失败: " + err.Error())
	}
	return toPolicyDTO(p), nil
}

func (s *backupService) Storage(ctx context.Context) (*dto.StorageStats, error) {
	count, size, err := s.rows.StorageStats(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.StorageStats{
		Count:              count,
		SizeBytes:          size,
		Prefix:             s.cfg.Prefix,
		Bucket:             s.cfg.Bucket,
		LocalPath:          s.cfg.LocalPath,
		Compression:        CompressionName,
		EncryptionEnabled:  len(parseEncryptionKey(s.cfg.EncryptionKey)) > 0,
		IncrementalEnabled: IncrementalSupported,
		PITREnabled:        PITRSupported,
	}, nil
}

func (s *backupService) Preview(ctx context.Context, id uuid.UUID) (*dto.Preview, error) {
	rec, err := s.rows.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errNotFound()
	}
	out := &dto.Preview{
		ID:                   rec.ID.String(),
		Filename:             rec.Filename,
		SizeBytes:            rec.SizeBytes,
		Encrypted:            rec.Encrypted,
		Compression:          CompressionName,
		EncryptionConfigured: len(parseEncryptionKey(s.cfg.EncryptionKey)) > 0,
	}
	if rec.ObjectKey == "" || rec.ChecksumSHA256 == "" {
		out.Error = "备份产物不完整"
		return out, nil
	}
	dump, info, err := s.downloadUnwrapped(ctx, rec)
	out.ChecksumOK = info.ChecksumOK
	out.Encrypted = info.Encrypted || rec.Encrypted
	out.DecryptOK = info.DecryptOK
	out.GzipOK = info.GzipOK
	if info.SizeBytes > 0 {
		out.SizeBytes = info.SizeBytes
	}
	if err != nil {
		out.Error = err.Error()
		return out, nil
	}
	defer func() { _ = dump.Close() }()
	toc, listErr := s.engine.List(ctx, dump)
	if listErr != nil {
		out.Error = listErr.Error()
		return out, nil
	}
	out.TOCValid = strings.TrimSpace(toc) != ""
	out.TOC = toc
	out.Ready = out.ChecksumOK && out.DecryptOK && out.GzipOK && out.TOCValid
	return out, nil
}

func (s *backupService) downloadUnwrapped(ctx context.Context, rec *model.Record) (io.ReadCloser, unwrapInfo, error) {
	var info unwrapInfo
	src, err := s.store.Get(ctx, rec.Storage, rec.ObjectKey)
	if err != nil {
		return nil, info, errStore("下载备份失败: " + err.Error())
	}
	defer func() { _ = src.Close() }()

	tmp, err := os.CreateTemp("", "starbyte-artifact-*.bin")
	if err != nil {
		return nil, info, err
	}
	tmpName := tmp.Name()
	if _, err := io.Copy(tmp, src); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return nil, info, errStore("读取备份失败: " + err.Error())
	}
	dump, info, err := unwrapStored(tmp, rec.ChecksumSHA256, parseEncryptionKey(s.cfg.EncryptionKey))
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return nil, info, err
	}
	return dump, info, nil
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
		if rec.Status == model.StatusSuccess || rec.Status == model.StatusFailed ||
			rec.Status == model.StatusRestored || rec.Status == model.StatusRestoreFailed {
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
