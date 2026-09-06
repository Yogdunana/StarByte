package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	task *dto.TaskResponse
	logs *dto.LogsResponse
	err  error
}

func (s *stubSvc) List(context.Context, dto.ListQuery) ([]dto.TaskResponse, int64, int, int, error) {
	if s.task == nil {
		return nil, 0, 1, 20, s.err
	}
	return []dto.TaskResponse{*s.task}, 1, 1, 20, s.err
}
func (s *stubSvc) Get(context.Context, uuid.UUID) (*dto.TaskResponse, error) { return s.task, s.err }
func (s *stubSvc) Create(context.Context, uuid.UUID, dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	return s.task, s.err
}
func (s *stubSvc) Update(context.Context, uuid.UUID, dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	return s.task, s.err
}
func (s *stubSvc) Delete(context.Context, uuid.UUID) error { return s.err }
func (s *stubSvc) Pause(context.Context, uuid.UUID) (*dto.TaskResponse, error) {
	return s.task, s.err
}
func (s *stubSvc) Resume(context.Context, uuid.UUID) (*dto.TaskResponse, error) {
	return s.task, s.err
}
func (s *stubSvc) RunNow(context.Context, uuid.UUID) error { return s.err }
func (s *stubSvc) Logs(context.Context, uuid.UUID, *uuid.UUID) (*dto.LogsResponse, error) {
	return s.logs, s.err
}
func (s *stubSvc) Handlers() []dto.HandlerInfo {
	return []dto.HandlerInfo{{Key: "noop", Description: "ok"}}
}

func doReq(h gin.HandlerFunc, method, path string, body any, params gin.Params, user string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	c.Request = httptest.NewRequest(method, path, &buf)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	c.Set("request_id", "rid")
	if user != "" {
		c.Set("user_id", user)
	}
	h(c)
	return w
}

func TestSchedulerHandler_CRUD(t *testing.T) {
	id := uuid.New()
	h := NewSchedulerHandler(&stubSvc{task: &dto.TaskResponse{ID: id.String(), Code: "echo"}})
	w := doReq(h.List, http.MethodGet, "/api/v1/system/scheduler/tasks", nil, nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(h.Get, http.MethodGet, "/tasks/"+id.String(), nil, gin.Params{{Key: "id", Value: id.String()}}, "")
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(h.Get, http.MethodGet, "/tasks/bad", nil, gin.Params{{Key: "id", Value: "bad"}}, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doReq(h.Handlers, http.MethodGet, "/handlers", nil, nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	uid := uuid.New().String()
	w = doReq(h.Create, http.MethodPost, "/tasks", dto.CreateTaskRequest{
		Name: "n", Code: "c", HandlerKey: "noop", CronExpr: "0 * * * * *",
	}, nil, uid)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(h.Create, http.MethodPost, "/tasks", dto.CreateTaskRequest{}, nil, uid)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doReq(h.Update, http.MethodPut, "/tasks/"+id.String(), dto.UpdateTaskRequest{}, gin.Params{{Key: "id", Value: id.String()}}, uid)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(h.Delete, http.MethodDelete, "/tasks/"+id.String(), nil, gin.Params{{Key: "id", Value: id.String()}}, uid)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSchedulerHandler_Actions(t *testing.T) {
	id := uuid.New()
	h := NewSchedulerHandler(&stubSvc{task: &dto.TaskResponse{ID: id.String()}, logs: &dto.LogsResponse{}})
	p := gin.Params{{Key: "id", Value: id.String()}}
	assert.Equal(t, http.StatusOK, doReq(h.Pause, http.MethodPost, "/pause", nil, p, "").Code)
	assert.Equal(t, http.StatusOK, doReq(h.Resume, http.MethodPost, "/resume", nil, p, "").Code)
	assert.Equal(t, http.StatusOK, doReq(h.Run, http.MethodPost, "/run", nil, p, "").Code)
	assert.Equal(t, http.StatusOK, doReq(h.Logs, http.MethodGet, "/logs", nil, p, "").Code)
	assert.Equal(t, http.StatusOK, doReq(h.Logs, http.MethodGet, "/logs?run_id="+id.String(), nil, p, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Logs, http.MethodGet, "/logs?run_id=bad", nil, p, "").Code)

	h = NewSchedulerHandler(&stubSvc{err: response.NewError(response.CodeSchedulerNotFound, "定时任务不存在")})
	w := doReq(h.Pause, http.MethodPost, "/pause", nil, p, "")
	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Resume, http.MethodPost, "/resume", nil, p, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Run, http.MethodPost, "/run", nil, p, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Logs, http.MethodGet, "/logs", nil, p, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Get, http.MethodGet, "/x", nil, p, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Update, http.MethodPut, "/x", dto.UpdateTaskRequest{}, p, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Delete, http.MethodDelete, "/x", nil, p, "").Code)
	assert.Equal(t, http.StatusUnauthorized, doReq(h.Create, http.MethodPost, "/tasks", dto.CreateTaskRequest{
		Name: "n", Code: "c", HandlerKey: "noop", CronExpr: "0 * * * * *",
	}, nil, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Pause, http.MethodPost, "/pause", nil, gin.Params{{Key: "id", Value: "bad"}}, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Resume, http.MethodPost, "/resume", nil, gin.Params{{Key: "id", Value: "bad"}}, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Run, http.MethodPost, "/run", nil, gin.Params{{Key: "id", Value: "bad"}}, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Logs, http.MethodGet, "/logs", nil, gin.Params{{Key: "id", Value: "bad"}}, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Delete, http.MethodDelete, "/x", nil, gin.Params{{Key: "id", Value: "bad"}}, "").Code)
	assert.Equal(t, http.StatusBadRequest, doReq(h.Update, http.MethodPut, "/x", "not-json", p, "").Code)

	r := gin.New()
	RegisterRoutes(r.Group("/api/v1"), h, nil)
}

func TestSchedulerHandler_ListBindError(t *testing.T) {
	h := NewSchedulerHandler(&stubSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tasks?page=abc", nil)
	c.Set("request_id", "rid")
	h.List(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
