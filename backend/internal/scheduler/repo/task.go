package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateTask(ctx context.Context, rec *model.Task) error
	UpdateTask(ctx context.Context, rec *model.Task) error
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	GetTaskByCode(ctx context.Context, code string) (*model.Task, error)
	ListTasks(ctx context.Context, keyword string, status *int16, offset, limit int) ([]model.Task, int64, error)
	ListDue(ctx context.Context, now time.Time, limit int) ([]model.Task, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Task, error)

	CreateRun(ctx context.Context, rec *model.Run) error
	UpdateRun(ctx context.Context, rec *model.Run) error
	GetRun(ctx context.Context, id uuid.UUID) (*model.Run, error)
	LatestSuccess(ctx context.Context, taskID uuid.UUID) (*model.Run, error)
	ListRuns(ctx context.Context, taskID uuid.UUID, offset, limit int) ([]model.Run, int64, error)

	AddLog(ctx context.Context, rec *model.RunLog) error
	ListLogs(ctx context.Context, runID uuid.UUID) ([]model.RunLog, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateTask(ctx context.Context, rec *model.Task) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *repository) UpdateTask(ctx context.Context, rec *model.Task) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *repository) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var rec model.Task
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) GetTaskByCode(ctx context.Context, code string) (*model.Task, error) {
	var rec model.Task
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) ListTasks(ctx context.Context, keyword string, status *int16, offset, limit int) ([]model.Task, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Task{}).Where("status <> ?", model.StatusDeleted)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Task
	err := q.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *repository) ListDue(ctx context.Context, now time.Time, limit int) ([]model.Task, error) {
	var list []model.Task
	err := r.db.WithContext(ctx).
		Where("status = ? AND next_run_at IS NOT NULL AND next_run_at <= ?", model.StatusActive, now).
		Order("next_run_at ASC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *repository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Task, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.Task
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	return list, err
}
