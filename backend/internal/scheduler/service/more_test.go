package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAlert struct{ n int }

func (s *stubAlert) TaskFailed(context.Context, uuid.UUID, string, string) { s.n++ }

func TestServiceOnceAndLogs(t *testing.T) {
	mem := newMemRepo()
	eng := NewEngine(mem, nil, nil)
	eng.sleeper = func(time.Duration) {}
	svc := NewService(mem, eng)
	ctx := context.Background()
	uid := uuid.New()
	runAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	created, err := svc.Create(ctx, uid, dto.CreateTaskRequest{
		Name: "once", Code: "once-job", HandlerKey: "echo", Payload: "ping", RunAt: &runAt,
	})
	require.NoError(t, err)
	assert.Empty(t, created.CronExpr)

	cron := "0 0 12 * * *"
	payload := "pong"
	handler := "noop"
	retries := 2
	timeout := 10
	shard := "a"
	updated, err := svc.Update(ctx, uuid.MustParse(created.ID), dto.UpdateTaskRequest{
		CronExpr: &cron, Payload: &payload, HandlerKey: &handler,
		MaxRetries: &retries, TimeoutSec: &timeout, ShardKey: &shard, DependsOn: []string{},
	})
	require.NoError(t, err)
	assert.Equal(t, "noop", updated.HandlerKey)
	assert.Equal(t, 2, updated.MaxRetries)

	require.NoError(t, svc.RunNow(ctx, uuid.MustParse(created.ID)))
	time.Sleep(30 * time.Millisecond)
	logs, err := svc.Logs(ctx, uuid.MustParse(created.ID), nil)
	require.NoError(t, err)
	assert.NotNil(t, logs)

	_, err = svc.Pause(ctx, uuid.MustParse(created.ID))
	if err == nil {
		_, err = svc.Pause(ctx, uuid.MustParse(created.ID))
		require.Error(t, err)
		_, err = svc.Resume(ctx, uuid.MustParse(created.ID))
		require.NoError(t, err)
		_, err = svc.Resume(ctx, uuid.MustParse(created.ID))
		require.Error(t, err)
	}

	assert.Nil(t, NewNotifAlerter(nil))
	_, err = parseTimePtr(strPtr("bad-time"))
	require.Error(t, err)
	assert.Equal(t, []string{"a"}, decodeDepends(encodeDepends([]string{"a"})))
	assert.Empty(t, decodeDepends("not-json"))
	ids, err := parseDepends([]string{uuid.NewString(), ""})
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func strPtr(s string) *string { return &s }

func TestEngineStartStopAndAlert(t *testing.T) {
	mem := newMemRepo()
	alert := &stubAlert{}
	eng := NewEngine(mem, nil, alert)
	eng.tick = 15 * time.Millisecond
	eng.sleeper = func(time.Duration) {}
	now := time.Now()
	uid := uuid.New()
	task := &model.Task{
		ID: uuid.New(), Name: "due", Code: "due", HandlerKey: "echo", Payload: "hi",
		Status: model.StatusActive, TimeoutSec: 5, MaxRetries: 0,
		NextRunAt: &now, CreatedBy: &uid, CreatedAt: now, UpdatedAt: now,
	}
	fail := &model.Task{
		ID: uuid.New(), Name: "dead", Code: "dead", HandlerKey: "fail",
		Status: model.StatusActive, TimeoutSec: 5, MaxRetries: 0,
		NextRunAt: &now, CreatedBy: &uid, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, mem.CreateTask(context.Background(), task))
	require.NoError(t, mem.CreateTask(context.Background(), fail))
	eng.Start()
	time.Sleep(40 * time.Millisecond)
	eng.Stop()
	got, err := mem.GetTask(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, model.RunSuccess, got.LastStatus)
	assert.Equal(t, 1, alert.n)
	require.NoError(t, eng.Trigger(context.Background(), task.ID))
	time.Sleep(20 * time.Millisecond)
}

func TestRunWithTimeout(t *testing.T) {
	err := runWithTimeout(context.Background(), time.Millisecond, func(ctx context.Context) error {
		<-ctx.Done()
		time.Sleep(20 * time.Millisecond)
		return ctx.Err()
	})
	require.Error(t, err)
	assert.Equal(t, time.Local, loadLocation("Not/AZone"))
}

func TestServiceListDefaultsAndBadUpdate(t *testing.T) {
	svc := NewService(newMemRepo(), nil)
	ctx := context.Background()
	uid := uuid.New()
	created, err := svc.Create(ctx, uid, dto.CreateTaskRequest{
		Name: "n", Code: "n1", HandlerKey: "noop", CronExpr: "0 * * * * *",
	})
	require.NoError(t, err)
	list, total, page, size, err := svc.List(ctx, dto.ListQuery{})
	require.NoError(t, err)
	assert.Equal(t, 1, page)
	assert.Equal(t, 20, size)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	bad := "nope"
	_, err = svc.Update(ctx, uuid.MustParse(created.ID), dto.UpdateTaskRequest{HandlerKey: &bad})
	require.Error(t, err)
	_, err = parseDepends([]string{"not-a-uuid"})
	require.Error(t, err)
	_, err = svc.Get(ctx, uuid.New())
	require.Error(t, err)
	require.Error(t, svc.Delete(ctx, uuid.New()))
	_, err = svc.Logs(ctx, uuid.New(), nil)
	require.Error(t, err)
}

func TestEngineRedisLock(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	mem := newMemRepo()
	eng := NewEngine(mem, rdb, nil)
	eng.sleeper = func(time.Duration) {}
	now := time.Now()
	task := &model.Task{
		ID: uuid.New(), Name: "lock", Code: "lock", HandlerKey: "noop",
		Status: model.StatusActive, TimeoutSec: 5, NextRunAt: &now,
	}
	require.NoError(t, mem.CreateTask(context.Background(), task))
	eng.dispatch(*task, true)
	got, err := mem.GetTask(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, model.RunSuccess, got.LastStatus)
}
