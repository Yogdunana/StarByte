package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/Yogdunana/StarByte/backend/pkg/cache"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (e *Engine) runLocked(ctx context.Context, task model.Task, manual bool, _ *cache.Lock) {
	if !manual {
		fresh, err := e.repo.GetTask(ctx, task.ID)
		if err != nil || fresh.Status != model.StatusActive {
			return
		}
		task = *fresh
	} else if task.Status == model.StatusDeleted {
		return
	}
	if !e.depsReady(ctx, &task) {
		e.shiftNextRun(ctx, task.ID, 5*time.Second)
		return
	}
	fn, ok := lookupHandler(task.HandlerKey)
	if !ok {
		logger.Warn("scheduler unknown handler", zap.String("handler", task.HandlerKey))
		e.shiftNextRun(ctx, task.ID, 5*time.Second)
		return
	}
	if !manual {
		lease := time.Duration(task.TimeoutSec+5) * time.Second
		if lease < 8*time.Second {
			lease = 8 * time.Second
		}
		e.shiftNextRun(ctx, task.ID, lease)
	}
	now := e.now()
	run := &model.Run{
		ID: uuid.New(), TaskID: task.ID, ScheduledAt: now,
		Status: model.RunRunning, Attempt: task.RetryCount + 1, WorkerID: e.workerID,
		CreatedAt: now,
	}
	st := now
	run.StartedAt = &st
	if err := e.repo.CreateRun(ctx, run); err != nil {
		logger.Warn("scheduler create run failed", zap.Error(err))
		return
	}
	timeout := time.Duration(task.TimeoutSec) * time.Second
	var logs []string
	logf := func(line string) {
		logs = append(logs, line)
		_ = e.repo.AddLog(ctx, &model.RunLog{
			ID: uuid.New(), RunID: run.ID, Level: "info", Line: line, CreatedAt: e.now(),
		})
	}
	err := runWithTimeout(ctx, timeout, func(c context.Context) error {
		return fn(c, task.Payload, logf)
	})
	fin := e.now()
	run.FinishedAt = &fin
	run.Output = strings.Join(logs, "\n")
	task.LastRunAt = &fin
	task.UpdatedAt = fin
	if err != nil {
		e.onFail(ctx, &task, run, err)
		return
	}
	e.onSuccess(ctx, &task, run)
}

func (e *Engine) onSuccess(ctx context.Context, task *model.Task, run *model.Run) {
	run.Status = model.RunSuccess
	_ = e.repo.UpdateRun(ctx, run)
	task.RetryCount = 0
	task.LastStatus = model.RunSuccess
	if task.IsCron() {
		n, err := nextRun(task.CronExpr, task.Timezone, e.now())
		if err == nil {
			task.NextRunAt = n
		}
	} else {
		task.Status = model.StatusFinished
		task.NextRunAt = nil
	}
	e.persistAfterRun(ctx, task)
}

func (e *Engine) onFail(ctx context.Context, task *model.Task, run *model.Run, execErr error) {
	_ = e.repo.AddLog(ctx, &model.RunLog{
		ID: uuid.New(), RunID: run.ID, Level: "error", Line: execErr.Error(), CreatedAt: e.now(),
	})
	run.ErrorText = execErr.Error()
	task.RetryCount++
	if task.RetryCount <= task.MaxRetries {
		run.Status = model.RunRetrying
		delay := backoff(task.RetryCount)
		n := e.now().Add(delay)
		task.NextRunAt = &n
		task.LastStatus = model.RunRetrying
		e.sleeper(0)
	} else {
		run.Status = model.RunDead
		task.LastStatus = model.RunDead
		task.RetryCount = 0
		if task.IsCron() {
			n, err := nextRun(task.CronExpr, task.Timezone, e.now())
			if err == nil {
				task.NextRunAt = n
			}
		} else {
			task.NextRunAt = nil
		}
		if e.alerter != nil && task.CreatedBy != nil {
			e.alerter.TaskFailed(ctx, *task.CreatedBy, task.Name, execErr.Error())
		}
	}
	_ = e.repo.UpdateRun(ctx, run)
	e.persistAfterRun(ctx, task)
}

func (e *Engine) shiftNextRun(ctx context.Context, id uuid.UUID, delay time.Duration) {
	rec, err := e.repo.GetTask(ctx, id)
	if err != nil || rec.Status != model.StatusActive {
		return
	}
	n := e.now().Add(delay)
	rec.NextRunAt = &n
	rec.UpdatedAt = e.now()
	_ = e.repo.UpdateTask(ctx, rec)
}

func (e *Engine) persistAfterRun(ctx context.Context, task *model.Task) {
	fresh, err := e.repo.GetTask(ctx, task.ID)
	if err != nil {
		_ = e.repo.UpdateTask(ctx, task)
		return
	}
	fresh.LastRunAt = task.LastRunAt
	fresh.LastStatus = task.LastStatus
	fresh.RetryCount = task.RetryCount
	fresh.UpdatedAt = task.UpdatedAt
	if fresh.Status == model.StatusActive {
		fresh.NextRunAt = task.NextRunAt
		if task.Status == model.StatusFinished {
			fresh.Status = model.StatusFinished
		}
	}
	_ = e.repo.UpdateTask(ctx, fresh)
}

func (e *Engine) depsReady(ctx context.Context, task *model.Task) bool {
	ids, err := parseDepends(decodeDepends(task.DependsOn))
	if err != nil || len(ids) == 0 {
		return err == nil
	}
	for _, id := range ids {
		ok, ferr := e.repo.LatestSuccess(ctx, id)
		if ferr != nil && !errors.Is(ferr, gorm.ErrRecordNotFound) {
			logger.Warn("scheduler dep check failed", zap.Error(ferr))
			return false
		}
		if ok == nil {
			return false
		}
	}
	return true
}
