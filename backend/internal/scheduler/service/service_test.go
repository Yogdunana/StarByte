package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronParseAndBackoff(t *testing.T) {
	_, err := parseCron("0 * * * * *")
	require.NoError(t, err)
	_, err = parseCron("not-a-cron")
	require.Error(t, err)
	n, err := nextRun("0 0 0 * * *", "Asia/Shanghai", time.Now())
	require.NoError(t, err)
	require.NotNil(t, n)
	assert.True(t, n.After(time.Now().Add(-time.Second)))
	assert.Equal(t, time.Second, backoff(1))
	assert.Equal(t, 2*time.Second, backoff(2))
	assert.Equal(t, 30*time.Second, backoff(20))
}

func TestServiceCRUD(t *testing.T) {
	svc := NewService(newMemRepo(), nil)
	ctx := context.Background()
	uid := uuid.New()

	_, err := svc.Create(ctx, uid, dto.CreateTaskRequest{Name: "x", Code: "x", HandlerKey: "noop"})
	require.Error(t, err)

	created, err := svc.Create(ctx, uid, dto.CreateTaskRequest{
		Name: "echo job", Code: "echo-job", HandlerKey: "echo",
		CronExpr: "0 */5 * * * *", Payload: "hi",
	})
	require.NoError(t, err)
	assert.Equal(t, "echo-job", created.Code)
	assert.NotNil(t, created.NextRunAt)

	_, err = svc.Create(ctx, uid, dto.CreateTaskRequest{
		Name: "dup", Code: "echo-job", HandlerKey: "noop", CronExpr: "0 * * * * *",
	})
	require.Error(t, err)

	got, err := svc.Get(ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Equal(t, "echo job", got.Name)

	name := "echo renamed"
	updated, err := svc.Update(ctx, uuid.MustParse(created.ID), dto.UpdateTaskRequest{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "echo renamed", updated.Name)

	paused, err := svc.Pause(ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Equal(t, int16(1), paused.Status)
	assert.Nil(t, paused.NextRunAt)

	err = svc.RunNow(ctx, uuid.MustParse(created.ID))
	require.Error(t, err)

	resumed, err := svc.Resume(ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Equal(t, int16(0), resumed.Status)

	require.NoError(t, svc.Delete(ctx, uuid.MustParse(created.ID)))
	_, err = svc.Get(ctx, uuid.MustParse(created.ID))
	require.Error(t, err)

	list, total, _, _, err := svc.List(ctx, dto.ListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
	assert.NotEmpty(t, svc.Handlers())
}

func TestServiceUnknownHandler(t *testing.T) {
	svc := NewService(newMemRepo(), nil)
	_, err := svc.Create(context.Background(), uuid.New(), dto.CreateTaskRequest{
		Name: "bad", Code: "bad", HandlerKey: "nope", CronExpr: "0 * * * * *",
	})
	require.Error(t, err)
	var app *response.AppError
	require.ErrorAs(t, err, &app)
	assert.Equal(t, response.CodeSchedulerBadHandler, app.Code)
}
