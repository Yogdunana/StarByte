package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ScheduledTaskSpec upserts the managed scheduler row for full backups.
type ScheduledTaskSpec struct {
	Code       string
	Name       string
	CronExpr   string
	Timezone   string
	HandlerKey string
	Enabled    bool
	TimeoutSec int
	NextRunAt  *time.Time
}

type Repository interface {
	CreateRecord(ctx context.Context, rec *model.Record) error
	UpdateRecord(ctx context.Context, rec *model.Record) error
	DeleteRecord(ctx context.Context, id uuid.UUID) error
	GetRecord(ctx context.Context, id uuid.UUID) (*model.Record, error)
	ListRecords(ctx context.Context, req *dto.ListRequest) ([]model.Record, int64, error)
	CountBusy(ctx context.Context) (int64, error)
	ListStale(ctx context.Context, before time.Time) ([]model.Record, error)
	ListExpired(ctx context.Context, before time.Time) ([]model.Record, error)
	StorageStats(ctx context.Context) (count int64, size int64, err error)
	GetPolicy(ctx context.Context) (*model.Policy, error)
	UpsertPolicy(ctx context.Context, p *model.Policy) error
	SyncScheduledTask(ctx context.Context, spec ScheduledTaskSpec) error
	MarkTerminal(ctx context.Context, id uuid.UUID, status int16, msg string, now time.Time) error
	ListOpsUserIDs(ctx context.Context) ([]uuid.UUID, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateRecord(ctx context.Context, rec *model.Record) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *repository) UpdateRecord(ctx context.Context, rec *model.Record) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *repository) MarkTerminal(ctx context.Context, id uuid.UUID, status int16, msg string, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	return r.db.WithContext(ctx).Model(&model.Record{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        status,
			"error_message": msg,
			"finished_at":   now,
			"updated_at":    now,
		}).Error
}

func (r *repository) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Record{}, "id = ?", id).Error
}

func (r *repository) GetRecord(ctx context.Context, id uuid.UUID) (*model.Record, error) {
	var rec model.Record
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) ListRecords(ctx context.Context, req *dto.ListRequest) ([]model.Record, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Record{})
	if req != nil && req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := 1, 20
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.PageSize > 0 {
			size = req.PageSize
		}
	}
	var rows []model.Record
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *repository) CountBusy(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Record{}).
		Where("status IN ?", []int16{model.StatusPending, model.StatusRunning, model.StatusRestoring}).
		Count(&n).Error
	return n, err
}

func (r *repository) ListStale(ctx context.Context, before time.Time) ([]model.Record, error) {
	var rows []model.Record
	err := r.db.WithContext(ctx).
		Where("status IN ? AND updated_at < ?", []int16{model.StatusPending, model.StatusRunning, model.StatusRestoring}, before).
		Find(&rows).Error
	return rows, err
}

func (r *repository) ListExpired(ctx context.Context, before time.Time) ([]model.Record, error) {
	var rows []model.Record
	err := r.db.WithContext(ctx).
		Where("status IN ? AND COALESCE(finished_at, created_at) < ?",
			[]int16{model.StatusSuccess, model.StatusFailed, model.StatusRestored, model.StatusRestoreFailed}, before).
		Find(&rows).Error
	return rows, err
}

func (r *repository) StorageStats(ctx context.Context) (int64, int64, error) {
	var out struct {
		Count int64
		Size  int64
	}
	err := r.db.WithContext(ctx).Model(&model.Record{}).
		Select("COUNT(*) AS count, COALESCE(SUM(size_bytes), 0) AS size").
		Where("status IN ?", []int16{model.StatusSuccess, model.StatusRestored, model.StatusRestoreFailed}).
		Scan(&out).Error
	return out.Count, out.Size, err
}

func (r *repository) GetPolicy(ctx context.Context) (*model.Policy, error) {
	var p model.Policy
	err := r.db.WithContext(ctx).Order("created_at ASC").First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) UpsertPolicy(ctx context.Context, p *model.Policy) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(p).Error
}

func (r *repository) SyncScheduledTask(ctx context.Context, spec ScheduledTaskSpec) error {
	status := int16(1) // paused
	if spec.Enabled {
		status = 0
	}
	timeout := spec.TimeoutSec
	if timeout <= 0 {
		timeout = 1800
	}
	res := r.db.WithContext(ctx).Exec(`
		UPDATE scheduler_tasks
		SET name = ?, cron_expr = ?, timezone = ?, handler_key = ?,
		    status = ?, timeout_sec = ?, next_run_at = ?, updated_at = NOW()
		WHERE code = ? AND status <> 2
	`, spec.Name, spec.CronExpr, spec.Timezone, spec.HandlerKey, status, timeout, spec.NextRunAt, spec.Code)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO scheduler_tasks (id, name, code, cron_expr, timezone, handler_key, next_run_at, timeout_sec, max_retries, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?)
	`, uuid.New(), spec.Name, spec.Code, spec.CronExpr, spec.Timezone, spec.HandlerKey, spec.NextRunAt, timeout, status).Error
}

func (r *repository) ListOpsUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT u.id
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE u.status = 0 AND u.deleted_at IS NULL
		  AND r.status = 0
		  AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
		  AND r.code IN ('president', 'super_admin', 'vice_president')
	`).Scan(&ids).Error
	return ids, err
}
