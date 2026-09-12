package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNow() time.Time {
	return time.Date(2026, 9, 12, 8, 0, 0, 0, time.UTC)
}

func newTestSvc() (*backupService, *memRepo, *memStore, *fakeEngine, *recAlerter) {
	rows := newMemRepo()
	store := newMemStore()
	eng := &fakeEngine{}
	alert := &recAlerter{}
	svc := &backupService{
		rows:   rows,
		store:  store,
		engine: eng,
		live:   config.DatabaseConfig{Host: "postgres", Port: 5432, User: "starbyte", DBName: "starbyte"},
		cfg:    withBackupDefaults(config.BackupConfig{Prefix: "backups", TimeoutSec: 30, Bucket: "starbyte"}),
		alert:  alert,
		now:    testNow,
		run:    func(fn func(context.Context)) { fn(context.Background()) },
	}
	return svc, rows, store, eng, alert
}

func TestRetentionCutoff(t *testing.T) {
	now := testNow()
	assert.Equal(t, now.AddDate(0, 0, -30), RetentionCutoff(now, 0))
	assert.Equal(t, now.AddDate(0, 0, -7), RetentionCutoff(now, 7))
}

func TestCanTransition(t *testing.T) {
	assert.True(t, CanTransition(model.StatusPending, model.StatusRunning))
	assert.True(t, CanTransition(model.StatusPending, model.StatusFailed))
	assert.True(t, CanTransition(model.StatusRunning, model.StatusSuccess))
	assert.True(t, CanTransition(model.StatusRunning, model.StatusFailed))
	assert.True(t, CanTransition(model.StatusSuccess, model.StatusRestoring))
	assert.True(t, CanTransition(model.StatusRestored, model.StatusRestoring))
	assert.True(t, CanTransition(model.StatusRestoring, model.StatusRestored))
	assert.True(t, CanTransition(model.StatusRestoring, model.StatusRestoreFailed))
	assert.True(t, CanTransition(model.StatusRestoreFailed, model.StatusRestoring))
	assert.False(t, CanTransition(model.StatusRestoring, model.StatusFailed))
	assert.False(t, CanTransition(model.StatusFailed, model.StatusRunning))
	assert.False(t, CanTransition(model.StatusFailed, model.StatusRestoring))
	assert.False(t, CanTransition(model.StatusSuccess, model.StatusFailed))
	assert.False(t, CanTransition(model.StatusPending, model.StatusSuccess))
}

func TestValidatePolicyFields(t *testing.T) {
	assert.NoError(t, ValidatePolicyFields(30, "0 30 2 * * *", "Asia/Shanghai"))
	assert.Error(t, ValidatePolicyFields(0, "0 30 2 * * *", "Asia/Shanghai"))
	assert.Error(t, ValidatePolicyFields(30, "not-a-cron", "Asia/Shanghai"))
	assert.Error(t, ValidatePolicyFields(30, "0 30 2 * * *", "Not/AZone"))
}

func TestValidateRestoreConfirm(t *testing.T) {
	assert.Error(t, validateRestoreConfirm(nil))
	assert.Error(t, validateRestoreConfirm(&dto.RestoreRequest{Confirm: true, Confirmation: "restore"}))
	assert.NoError(t, validateRestoreConfirm(&dto.RestoreRequest{Confirm: true, Confirmation: "RESTORE"}))
}

func TestCreateBackup_SuccessAndChecksum(t *testing.T) {
	svc, _, store, _, _ := newTestSvc()
	uid := uuid.New()
	rec, err := svc.Create(context.Background(), uid, &dto.CreateRequest{})
	require.NoError(t, err)
	require.NotNil(t, rec)

	got, err := svc.Get(context.Background(), uuid.MustParse(rec.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got.Status)
	assert.Equal(t, model.TriggerManual, got.TriggerSource)
	assert.NotEmpty(t, got.ChecksumSHA256)
	assert.Greater(t, got.SizeBytes, int64(0))
	assert.Contains(t, store.data, got.ObjectKey)
}

func TestCreateBackup_DumpFailNotifies(t *testing.T) {
	svc, _, _, eng, alert := newTestSvc()
	eng.dumpErr = errors.New("pg_dump missing")
	rec, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	got, err := svc.Get(context.Background(), uuid.MustParse(rec.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusFailed, got.Status)
	assert.Contains(t, got.ErrorMessage, "pg_dump")
	assert.Equal(t, 1, alert.n)
}

func TestCreateBackup_Busy(t *testing.T) {
	svc, rows, _, _, _ := newTestSvc()
	now := testNow()
	rows.records[uuid.New()] = &model.Record{ID: uuid.New(), Status: model.StatusRunning, CreatedAt: now, UpdatedAt: now}
	_, err := svc.Create(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.Equal(t, response.CodeBackupBusy, err.(*response.AppError).Code)
}

func TestRestore_RequiresTypedConfirmation(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	_, err = svc.Restore(context.Background(), uuid.New(), id, &dto.RestoreRequest{Confirm: true, Confirmation: "YES"})
	require.Error(t, err)
	assert.Equal(t, response.CodeBackupConfirmRequired, err.(*response.AppError).Code)
}

func TestDrillRestore_IndependentDB(t *testing.T) {
	svc, _, _, eng, alert := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)

	_, err = svc.DrillRestore(context.Background(), uuid.New(), id, &dto.DrillRequest{
		Confirm: true, Confirmation: "RESTORE", TargetDBName: "starbyte_drill",
	})
	require.Error(t, err)

	_, err = svc.DrillRestore(context.Background(), uuid.New(), id, &dto.DrillRequest{
		Confirm: true, Confirmation: model.DrillConfirmToken, TargetDBName: "starbyte",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能是当前应用库")

	out, err := svc.DrillRestore(context.Background(), uuid.New(), id, &dto.DrillRequest{
		Confirm: true, Confirmation: model.DrillConfirmToken, TargetDBName: "starbyte_drill",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, out.Restored)
	assert.Equal(t, "starbyte_drill", out.TargetDBName)
	assert.Equal(t, "starbyte_drill", eng.target.DBName)
	assert.Equal(t, "postgres", eng.target.Host)
	got, err := svc.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got.Status)
	assert.Equal(t, 0, alert.n)
}

func TestDrillRestore_FailureAlertsWithoutFlippingRecord(t *testing.T) {
	svc, _, _, eng, alert := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	eng.restoreErr = errors.New("pg_restore boom")
	out, err := svc.DrillRestore(context.Background(), uuid.New(), uuid.MustParse(created.ID), &dto.DrillRequest{
		Confirm: true, Confirmation: "DRILL", TargetDBName: "starbyte_drill",
	})
	require.NoError(t, err)
	assert.False(t, out.Restored)
	assert.Contains(t, out.Error, "pg_restore")
	assert.Equal(t, 1, alert.n)
	got, err := svc.Get(context.Background(), uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got.Status)
}

func TestRestore_Success(t *testing.T) {
	svc, _, _, eng, _ := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	out, err := svc.Restore(context.Background(), uuid.New(), id, &dto.RestoreRequest{Confirm: true, Confirmation: model.RestoreConfirmToken})
	require.NoError(t, err)
	got, err := svc.Get(context.Background(), uuid.MustParse(out.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusRestored, got.Status)
	assert.Equal(t, "-- starbyte dump\nSELECT 1;\n", string(eng.restored))
}

func TestRestore_ChecksumMismatch(t *testing.T) {
	svc, rows, store, _, _ := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	rec, _ := rows.GetRecord(context.Background(), id)
	original := append([]byte(nil), store.data[rec.ObjectKey]...)
	store.data[rec.ObjectKey] = []byte("tampered")
	_, err = svc.Restore(context.Background(), uuid.New(), id, &dto.RestoreRequest{Confirm: true, Confirmation: "RESTORE"})
	require.NoError(t, err)
	got, err := svc.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, model.StatusRestoreFailed, got.Status)
	assert.NotEmpty(t, got.ErrorMessage)
	store.data[rec.ObjectKey] = original
	_, err = svc.Restore(context.Background(), uuid.New(), id, &dto.RestoreRequest{Confirm: true, Confirmation: "RESTORE"})
	require.NoError(t, err)
	got, err = svc.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, model.StatusRestored, got.Status)
}

func TestDeleteRemovesArtifact(t *testing.T) {
	svc, _, store, _, _ := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	got, _ := svc.Get(context.Background(), uuid.MustParse(created.ID))
	require.NoError(t, svc.Delete(context.Background(), uuid.MustParse(created.ID)))
	_, ok := store.data[got.ObjectKey]
	assert.False(t, ok)
	_, err = svc.Get(context.Background(), uuid.MustParse(created.ID))
	require.Error(t, err)
	assert.Equal(t, response.CodeBackupNotFound, err.(*response.AppError).Code)
}

func TestPolicyUpdateSyncsScheduler(t *testing.T) {
	svc, rows, _, _, _ := newTestSvc()
	enabled := true
	days := 14
	cron := "0 0 3 * * *"
	out, err := svc.UpdatePolicy(context.Background(), uuid.New(), &dto.UpdatePolicyRequest{
		Enabled: &enabled, RetentionDays: &days, CronExpr: &cron,
	})
	require.NoError(t, err)
	assert.True(t, out.Enabled)
	assert.Equal(t, 14, out.RetentionDays)
	assert.Equal(t, cron, out.CronExpr)
	spec := rows.tasks[model.ScheduledTaskCode]
	assert.True(t, spec.Enabled)
	assert.Equal(t, cron, spec.CronExpr)
}

func TestPolicyUpdate_SyncError(t *testing.T) {
	svc, rows, _, _, _ := newTestSvc()
	rows.syncErr = errors.New("scheduler down")
	enabled := true
	_, err := svc.UpdatePolicy(context.Background(), uuid.New(), &dto.UpdatePolicyRequest{Enabled: &enabled})
	require.Error(t, err)
	assert.Equal(t, response.CodeBackupPolicyInvalid, err.(*response.AppError).Code)
	assert.Contains(t, err.Error(), "调度任务同步失败")
}

func TestDump_PersistRunningFailMarksFailed(t *testing.T) {
	svc, rows, _, _, _ := newTestSvc()
	rows.updateFailN = 1
	rec, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	got, err := svc.Get(context.Background(), uuid.MustParse(rec.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusFailed, got.Status)
	busy, err := rows.CountBusy(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(0), busy)
}

func TestDump_GetRecordMissMarksFailed(t *testing.T) {
	svc, rows, _, _, _ := newTestSvc()
	rows.getErr = errors.New("db down")
	rec, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	rows.getErr = nil
	got, err := svc.Get(context.Background(), uuid.MustParse(rec.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusFailed, got.Status)
	assert.Contains(t, got.ErrorMessage, "读取备份记录失败")
	busy, err := rows.CountBusy(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(0), busy)
}

func TestRestore_GetRecordMissMarksRestoreFailed(t *testing.T) {
	svc, rows, _, _, _ := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	rows.getCalls = 0
	rows.getFailAfter = 1
	_, err = svc.Restore(context.Background(), uuid.New(), id, &dto.RestoreRequest{Confirm: true, Confirmation: "RESTORE"})
	require.NoError(t, err)
	rows.getFailAfter = 0
	got, err := svc.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, model.StatusRestoreFailed, got.Status)
	busy, err := rows.CountBusy(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(0), busy)
}

func TestCleanupExpired_RespectsRetention(t *testing.T) {
	svc, rows, store, _, _ := newTestSvc()
	keepID := uuid.New()
	dropID := uuid.New()
	now := testNow()
	old := now.AddDate(0, 0, -40)
	rows.records[keepID] = &model.Record{
		ID: keepID, Status: model.StatusSuccess, ObjectKey: "backups/keep.sql.gz",
		SizeBytes: 10, CreatedAt: now, UpdatedAt: now, FinishedAt: &now,
	}
	rows.records[dropID] = &model.Record{
		ID: dropID, Status: model.StatusSuccess, ObjectKey: "backups/drop.sql.gz",
		SizeBytes: 20, CreatedAt: old, UpdatedAt: old, FinishedAt: &old,
	}
	store.data["backups/keep.sql.gz"] = []byte("keep")
	store.data["backups/drop.sql.gz"] = []byte("drop")
	rows.policy = &model.Policy{ID: uuid.New(), RetentionDays: 30, CronExpr: model.DefaultCronExpr, Timezone: model.DefaultTimezone}

	err := svc.CleanupExpired(context.Background(), "", func(string) {})
	require.NoError(t, err)
	assert.NotNil(t, rows.records[keepID])
	assert.Nil(t, rows.records[dropID])
	_, dropped := store.data["backups/drop.sql.gz"]
	assert.False(t, dropped)
	_, kept := store.data["backups/keep.sql.gz"]
	assert.True(t, kept)
}

func TestRunScheduled_SkipWhenDisabled(t *testing.T) {
	svc, rows, _, _, _ := newTestSvc()
	err := svc.RunScheduled(context.Background(), "", func(string) {})
	require.NoError(t, err)
	assert.Empty(t, rows.records)
}

func TestStorageStats(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	_, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	st, err := svc.Storage(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), st.Count)
	assert.Greater(t, st.SizeBytes, int64(0))
	assert.Equal(t, "backups", st.Prefix)
	assert.Equal(t, "gzip", st.Compression)
	assert.False(t, st.EncryptionEnabled)
	assert.False(t, st.IncrementalEnabled)
	assert.False(t, st.PITREnabled)
}

func TestCreateBackup_EncryptsWhenKeySet(t *testing.T) {
	svc, _, store, eng, _ := newTestSvc()
	svc.cfg.EncryptionKey = "unit-test-backup-key"
	rec, err := svc.Create(context.Background(), uuid.New(), &dto.CreateRequest{})
	require.NoError(t, err)
	got, err := svc.Get(context.Background(), uuid.MustParse(rec.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got.Status)
	assert.True(t, got.Encrypted)
	assert.True(t, strings.HasSuffix(got.Filename, ".dump.gz.enc"))
	assert.True(t, isEncryptedBlob(store.data[got.ObjectKey]))

	out, err := svc.Restore(context.Background(), uuid.New(), uuid.MustParse(got.ID), &dto.RestoreRequest{
		Confirm: true, Confirmation: model.RestoreConfirmToken,
	})
	require.NoError(t, err)
	restored, err := svc.Get(context.Background(), uuid.MustParse(out.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusRestored, restored.Status)
	assert.Equal(t, "-- starbyte dump\nSELECT 1;\n", string(eng.restored))
}

func TestRestore_EncryptedWithoutKeyFails(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	svc.cfg.EncryptionKey = "unit-test-backup-key"
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	svc.cfg.EncryptionKey = ""
	_, err = svc.Restore(context.Background(), uuid.New(), id, &dto.RestoreRequest{Confirm: true, Confirmation: "RESTORE"})
	require.NoError(t, err)
	got, err := svc.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, model.StatusRestoreFailed, got.Status)
	assert.Contains(t, got.ErrorMessage, "STARBYTE_BACKUP_ENCRYPTION_KEY")
}

func TestPreview_ReadyOnIntactDump(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	prev, err := svc.Preview(context.Background(), uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.True(t, prev.ChecksumOK)
	assert.True(t, prev.GzipOK)
	assert.True(t, prev.TOCValid)
	assert.True(t, prev.Ready)
	assert.Equal(t, "gzip", prev.Compression)
	assert.False(t, prev.Encrypted)
}

func TestPreview_ChecksumMismatchNotReady(t *testing.T) {
	svc, rows, store, _, _ := newTestSvc()
	created, err := svc.Create(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	rec, _ := rows.GetRecord(context.Background(), id)
	store.data[rec.ObjectKey] = []byte("tampered")
	prev, err := svc.Preview(context.Background(), id)
	require.NoError(t, err)
	assert.False(t, prev.Ready)
	assert.False(t, prev.ChecksumOK)
	assert.NotEmpty(t, prev.Error)
}

func TestCleanupExpired_SkipsInFlight(t *testing.T) {
	svc, rows, store, _, _ := newTestSvc()
	now := testNow()
	old := now.AddDate(0, 0, -40)
	busyID := uuid.New()
	dropID := uuid.New()
	rows.records[busyID] = &model.Record{
		ID: busyID, Status: model.StatusRunning, ObjectKey: "backups/busy.sql.gz",
		CreatedAt: old, UpdatedAt: old,
	}
	rows.records[dropID] = &model.Record{
		ID: dropID, Status: model.StatusSuccess, ObjectKey: "backups/drop.sql.gz",
		CreatedAt: old, UpdatedAt: old, FinishedAt: &old,
	}
	store.data["backups/busy.sql.gz"] = []byte("busy")
	store.data["backups/drop.sql.gz"] = []byte("drop")
	rows.policy = &model.Policy{ID: uuid.New(), RetentionDays: 7, CronExpr: model.DefaultCronExpr, Timezone: model.DefaultTimezone}

	err := svc.CleanupExpired(context.Background(), "", func(string) {})
	require.NoError(t, err)
	assert.NotNil(t, rows.records[busyID])
	assert.Nil(t, rows.records[dropID])
}

func TestRunScheduled_FailNotifies(t *testing.T) {
	svc, rows, _, eng, alert := newTestSvc()
	rows.policy = &model.Policy{ID: uuid.New(), Enabled: true, RetentionDays: 30, CronExpr: model.DefaultCronExpr, Timezone: model.DefaultTimezone}
	eng.dumpErr = errors.New("disk full")
	err := svc.RunScheduled(context.Background(), "", func(string) {})
	require.Error(t, err)
	assert.Equal(t, 1, alert.n)
}

func TestCollectAlertUserIDs(t *testing.T) {
	created := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	ops := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	ids := collectAlertUserIDs(&created, []uuid.UUID{ops, created, uuid.Nil})
	assert.Equal(t, []uuid.UUID{created, ops}, ids)
	assert.Empty(t, collectAlertUserIDs(nil, nil))
}

func TestGetMissing(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	_, err := svc.Get(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, response.CodeBackupNotFound, err.(*response.AppError).Code)
}
