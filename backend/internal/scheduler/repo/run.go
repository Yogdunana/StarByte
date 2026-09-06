package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/google/uuid"
)

func (r *repository) CreateRun(ctx context.Context, rec *model.Run) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *repository) UpdateRun(ctx context.Context, rec *model.Run) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *repository) GetRun(ctx context.Context, id uuid.UUID) (*model.Run, error) {
	var rec model.Run
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) LatestSuccess(ctx context.Context, taskID uuid.UUID) (*model.Run, error) {
	var rec model.Run
	err := r.db.WithContext(ctx).
		Where("task_id = ? AND status = ?", taskID, model.RunSuccess).
		Order("finished_at DESC").First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) ListRuns(ctx context.Context, taskID uuid.UUID, offset, limit int) ([]model.Run, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Run{}).Where("task_id = ?", taskID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Run
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *repository) AddLog(ctx context.Context, rec *model.RunLog) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *repository) ListLogs(ctx context.Context, runID uuid.UUID) ([]model.RunLog, error) {
	var list []model.RunLog
	err := r.db.WithContext(ctx).Where("run_id = ?", runID).Order("created_at ASC").Find(&list).Error
	return list, err
}
