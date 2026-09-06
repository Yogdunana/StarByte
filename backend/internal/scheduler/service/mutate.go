package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *schedulerService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	code := strings.TrimSpace(req.Code)
	if existing, err := s.repo.GetTaskByCode(ctx, code); err == nil && existing != nil {
		return nil, response.NewError(response.CodeSchedulerCodeExists, "任务编码已存在")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if _, ok := lookupHandler(req.HandlerKey); !ok {
		return nil, response.NewError(response.CodeSchedulerBadHandler, "未知的任务处理器")
	}
	if req.CronExpr == "" && req.RunAt == nil {
		return nil, response.NewError(response.CodeSchedulerInvalidCron, "必须提供 Cron 表达式或一次性执行时间")
	}
	if _, err := parseCron(req.CronExpr); err != nil {
		return nil, err
	}
	if _, err := parseDepends(req.DependsOn); err != nil {
		return nil, response.NewError(response.CodeBadRequest, "依赖任务 ID 无效")
	}
	runAt, err := parseTimePtr(req.RunAt)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "执行时间格式无效")
	}
	tz := strings.TrimSpace(req.Timezone)
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	now := time.Now()
	rec := &model.Task{
		ID: uuid.New(), Name: strings.TrimSpace(req.Name), Code: code,
		CronExpr: strings.TrimSpace(req.CronExpr), RunAt: runAt, Timezone: tz,
		HandlerKey: req.HandlerKey, Payload: req.Payload, DependsOn: encodeDepends(req.DependsOn),
		ShardKey: strings.TrimSpace(req.ShardKey), Status: model.StatusActive,
		MaxRetries: defaultInt(req.MaxRetries, 3, 0, 20),
		TimeoutSec: defaultInt(req.TimeoutSec, 60, 1, 3600),
		CreatedBy:  &userID, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.computeNext(rec); err != nil {
		return nil, err
	}
	if err := s.repo.CreateTask(ctx, rec); err != nil {
		return nil, err
	}
	out := toTaskDTO(rec)
	return &out, nil
}

func (s *schedulerService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	rec, err := s.getActive(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.Status == model.StatusFinished {
		return nil, response.NewError(response.CodeSchedulerPaused, "已结束的一次性任务不可修改")
	}
	if req.Name != nil {
		rec.Name = strings.TrimSpace(*req.Name)
	}
	if req.CronExpr != nil {
		if _, err := parseCron(*req.CronExpr); err != nil {
			return nil, err
		}
		rec.CronExpr = strings.TrimSpace(*req.CronExpr)
	}
	if req.RunAt != nil {
		t, err := parseTimePtr(req.RunAt)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "执行时间格式无效")
		}
		rec.RunAt = t
	}
	if req.Timezone != nil && strings.TrimSpace(*req.Timezone) != "" {
		rec.Timezone = strings.TrimSpace(*req.Timezone)
	}
	if req.HandlerKey != nil {
		if _, ok := lookupHandler(*req.HandlerKey); !ok {
			return nil, response.NewError(response.CodeSchedulerBadHandler, "未知的任务处理器")
		}
		rec.HandlerKey = *req.HandlerKey
	}
	if req.Payload != nil {
		rec.Payload = *req.Payload
	}
	if req.DependsOn != nil {
		if _, err := parseDepends(req.DependsOn); err != nil {
			return nil, response.NewError(response.CodeBadRequest, "依赖任务 ID 无效")
		}
		rec.DependsOn = encodeDepends(req.DependsOn)
	}
	if req.ShardKey != nil {
		rec.ShardKey = strings.TrimSpace(*req.ShardKey)
	}
	if req.MaxRetries != nil {
		rec.MaxRetries = defaultInt(req.MaxRetries, rec.MaxRetries, 0, 20)
	}
	if req.TimeoutSec != nil {
		rec.TimeoutSec = defaultInt(req.TimeoutSec, rec.TimeoutSec, 1, 3600)
	}
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

func (s *schedulerService) Delete(ctx context.Context, id uuid.UUID) error {
	rec, err := s.getActive(ctx, id)
	if err != nil {
		return err
	}
	rec.Status = model.StatusDeleted
	rec.NextRunAt = nil
	rec.UpdatedAt = time.Now()
	return s.repo.UpdateTask(ctx, rec)
}
