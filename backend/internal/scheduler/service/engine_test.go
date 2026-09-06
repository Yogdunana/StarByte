package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_NoopAndRetryDead(t *testing.T) {
	mem := newMemRepo()
	eng := NewEngine(mem, nil, nil)
	eng.sleeper = func(time.Duration) {}
	ctx := context.Background()
	now := time.Now()
	eng.now = func() time.Time { return now }

	okTask := &model.Task{
		ID: uuid.New(), Name: "ok", Code: "ok", HandlerKey: "noop",
		Status: model.StatusActive, TimeoutSec: 5, MaxRetries: 1,
		CronExpr: "0 0 0 * * *", NextRunAt: &now, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, mem.CreateTask(ctx, okTask))
	eng.dispatch(*okTask, true)
	got, err := mem.GetTask(ctx, okTask.ID)
	require.NoError(t, err)
	assert.Equal(t, model.RunSuccess, got.LastStatus)

	fail := &model.Task{
		ID: uuid.New(), Name: "fail", Code: "fail", HandlerKey: "fail",
		Payload: "boom", Status: model.StatusActive, TimeoutSec: 5, MaxRetries: 1,
		NextRunAt: &now, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, mem.CreateTask(ctx, fail))
	eng.dispatch(*fail, true)
	got, err = mem.GetTask(ctx, fail.ID)
	require.NoError(t, err)
	assert.Equal(t, model.RunRetrying, got.LastStatus)
	assert.Equal(t, 1, got.RetryCount)

	eng.dispatch(*got, true)
	got, err = mem.GetTask(ctx, fail.ID)
	require.NoError(t, err)
	assert.Equal(t, model.RunDead, got.LastStatus)

	runs, total, err := mem.ListRuns(ctx, fail.ID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, runs, 2)
}

func TestEngine_SkipWhenDepsMissing(t *testing.T) {
	mem := newMemRepo()
	eng := NewEngine(mem, nil, nil)
	ctx := context.Background()
	now := time.Now()
	dep := uuid.New()
	task := &model.Task{
		ID: uuid.New(), Name: "child", Code: "child", HandlerKey: "noop",
		DependsOn: encodeDepends([]string{dep.String()}),
		Status:    model.StatusActive, TimeoutSec: 5, NextRunAt: &now,
	}
	require.NoError(t, mem.CreateTask(ctx, task))
	eng.dispatch(*task, true)
	got, err := mem.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Empty(t, got.LastStatus)
}

func TestLockName(t *testing.T) {
	assert.Equal(t, "scheduler:task:abc", lockName("abc"))
}
