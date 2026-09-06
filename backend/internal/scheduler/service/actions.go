package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *schedulerService) Pause(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error) {
	rec, err := s.getActive(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.Status != model.StatusActive {
		return nil, response.NewError(response.CodeSchedulerPaused, "任务未处于可暂停状态")
	}
	rec.Status = model.StatusPaused
	rec.NextRunAt = nil
	rec.UpdatedAt = time.Now()
	if err := s.repo.UpdateTask(ctx, rec); err != nil {
		return nil, err
	}
	out := toTaskDTO(rec)
	return &out, nil
}

func (s *schedulerService) Resume(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error) {
	rec, err := s.getActive(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.Status != model.StatusPaused {
		return nil, response.NewError(response.CodeSchedulerPaused, "仅暂停中的任务可恢复")
	}
	rec.Status = model.StatusActive
	rec.UpdatedAt = time.Now()
	if err := s.computeNext(rec); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateTask(ctx, rec); err != nil {
		return nil, err
	}
	out := toTaskDTO(rec)
	return &out, nil
}

func (s *schedulerService) RunNow(ctx context.Context, id uuid.UUID) error {
	rec, err := s.getActive(ctx, id)
	if err != nil {
		return err
	}
	if rec.Status == model.StatusPaused {
		return response.NewError(response.CodeSchedulerPaused, "已暂停的任务不可手动执行")
	}
	if s.engine == nil {
		return response.NewError(response.CodeSchedulerBusy, "调度引擎未启动")
	}
	return s.engine.Trigger(ctx, rec.ID)
}

func (s *schedulerService) Logs(ctx context.Context, id uuid.UUID, runID *uuid.UUID) (*dto.LogsResponse, error) {
	if _, err := s.getActive(ctx, id); err != nil {
		return nil, err
	}
	runs, _, err := s.repo.ListRuns(ctx, id, 0, 20)
	if err != nil {
		return nil, err
	}
	out := &dto.LogsResponse{Runs: make([]dto.RunResponse, 0, len(runs)), Logs: []dto.LogLine{}}
	for i := range runs {
		out.Runs = append(out.Runs, toRunDTO(&runs[i]))
	}
	target := runID
	if target == nil && len(runs) > 0 {
		rid := runs[0].ID
		target = &rid
	}
	if target != nil {
		lines, err := s.repo.ListLogs(ctx, *target)
		if err != nil {
			return nil, err
		}
		for i := range lines {
			out.Logs = append(out.Logs, toLogDTO(&lines[i]))
		}
	}
	return out, nil
}
