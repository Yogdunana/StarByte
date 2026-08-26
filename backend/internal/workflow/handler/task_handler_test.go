package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock TaskService ---

type mockTaskService struct {
	listTodoFunc func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error)
	listDoneFunc func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error)
	getTaskFunc  func(ctx context.Context, id uuid.UUID) (*model.FlowTask, error)
	completeFunc func(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action string, comment string, formData map[string]interface{}) error
	transferFunc func(ctx context.Context, taskID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID, comment string) error
	rollbackFunc func(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, targetNodeID string, comment string) error
}

func (m *mockTaskService) ListTodoTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	if m.listTodoFunc != nil {
		return m.listTodoFunc(ctx, userID, page, pageSize)
	}
	return []model.FlowTask{}, 0, nil
}

func (m *mockTaskService) ListDoneTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	if m.listDoneFunc != nil {
		return m.listDoneFunc(ctx, userID, page, pageSize)
	}
	return []model.FlowTask{}, 0, nil
}

func (m *mockTaskService) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.FlowTask, error) {
	if m.getTaskFunc != nil {
		return m.getTaskFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockTaskService) CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action string, comment string, formData map[string]interface{}) error {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, taskID, userID, action, comment, formData)
	}
	return nil
}

func (m *mockTaskService) TransferTask(ctx context.Context, taskID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID, comment string) error {
	if m.transferFunc != nil {
		return m.transferFunc(ctx, taskID, fromUserID, toUserID, comment)
	}
	return nil
}

func (m *mockTaskService) RollbackTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, targetNodeID string, comment string) error {
	if m.rollbackFunc != nil {
		return m.rollbackFunc(ctx, taskID, userID, targetNodeID, comment)
	}
	return nil
}

// Compile-time check.
var _ service.TaskService = (*mockTaskService)(nil)

// --- TaskHandler Tests ---

func TestTaskHandler_ListTodoTasks_Success(t *testing.T) {
	userID := uuid.New()
	var capturedUserID uuid.UUID

	svc := &mockTaskService{
		listTodoFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
			capturedUserID = uid
			return []model.FlowTask{
				{ID: uuid.New(), NodeName: "Manager Approval", Status: 0},
			}, 1, nil
		},
	}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.GET("/tasks/todo", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.ListTodoTasks(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/tasks/todo?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, userID, capturedUserID)
}

func TestTaskHandler_ListTodoTasks_NoAuth(t *testing.T) {
	svc := &mockTaskService{}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.GET("/tasks/todo", h.ListTodoTasks)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/tasks/todo", nil)
	r.ServeHTTP(w, req)

	// Without auth middleware setting user_id, getUserID returns error.
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_ListDoneTasks_Success(t *testing.T) {
	userID := uuid.New()

	svc := &mockTaskService{
		listDoneFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
			return []model.FlowTask{
				{ID: uuid.New(), NodeName: "Manager Approval", Status: 1, Action: "approve"},
			}, 1, nil
		},
	}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.GET("/tasks/done", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.ListDoneTasks(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/tasks/done", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_GetByID_Success(t *testing.T) {
	taskID := uuid.New()
	svc := &mockTaskService{
		getTaskFunc: func(ctx context.Context, id uuid.UUID) (*model.FlowTask, error) {
			return &model.FlowTask{ID: id, NodeName: "Approval", Status: 0}, nil
		},
	}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.GET("/tasks/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/tasks/"+taskID.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Approval", data["node_name"])
}

func TestTaskHandler_GetByID_InvalidUUID(t *testing.T) {
	svc := &mockTaskService{}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.GET("/tasks/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/tasks/invalid-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Approve_Success(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()
	var capturedAction string
	var capturedComment string

	svc := &mockTaskService{
		completeFunc: func(ctx context.Context, tID uuid.UUID, uID uuid.UUID, action string, comment string, formData map[string]interface{}) error {
			capturedAction = action
			capturedComment = comment
			return nil
		},
	}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.POST("/tasks/:id/approve", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Approve(c)
	})

	body := `{"action":"approve","comment":"同意"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/tasks/"+taskID.String()+"/approve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "approve", capturedAction)
	assert.Equal(t, "同意", capturedComment)
}

func TestTaskHandler_Reject_Success(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()
	var capturedAction string

	svc := &mockTaskService{
		completeFunc: func(ctx context.Context, tID uuid.UUID, uID uuid.UUID, action string, comment string, formData map[string]interface{}) error {
			capturedAction = action
			return nil
		},
	}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.POST("/tasks/:id/reject", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Reject(c)
	})

	body := `{"action":"reject","comment":"不同意"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/tasks/"+taskID.String()+"/reject", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "reject", capturedAction)
}

func TestTaskHandler_Transfer_Success(t *testing.T) {
	taskID := uuid.New()
	fromUserID := uuid.New()
	toUserID := uuid.New()
	var capturedFromID uuid.UUID
	var capturedToID uuid.UUID

	svc := &mockTaskService{
		transferFunc: func(ctx context.Context, tID uuid.UUID, fromID uuid.UUID, toID uuid.UUID, comment string) error {
			capturedFromID = fromID
			capturedToID = toID
			return nil
		},
	}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.POST("/tasks/:id/transfer", func(c *gin.Context) {
		setAuthUser(c, fromUserID.String())
		h.Transfer(c)
	})

	body := `{"target_user_id":"` + toUserID.String() + `","comment":"转交给张三"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/tasks/"+taskID.String()+"/transfer", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fromUserID, capturedFromID)
	assert.Equal(t, toUserID, capturedToID)
}

func TestTaskHandler_Transfer_MissingTargetUserID(t *testing.T) {
	svc := &mockTaskService{}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.POST("/tasks/:id/transfer", func(c *gin.Context) {
		setAuthUser(c, uuid.New().String())
		h.Transfer(c)
	})

	body := `{"comment":"missing target"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/tasks/"+uuid.New().String()+"/transfer", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Rollback_Success(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()
	var capturedTargetNodeID string

	svc := &mockTaskService{
		rollbackFunc: func(ctx context.Context, tID uuid.UUID, uID uuid.UUID, targetNodeID string, comment string) error {
			capturedTargetNodeID = targetNodeID
			return nil
		},
	}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.POST("/tasks/:id/rollback", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Rollback(c)
	})

	body := `{"target_node_id":"start_node","comment":"退回起点"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/tasks/"+taskID.String()+"/rollback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "start_node", capturedTargetNodeID)
}

func TestTaskHandler_Rollback_MissingTargetNodeID(t *testing.T) {
	svc := &mockTaskService{}
	h := NewTaskHandler(svc)

	r := setupRouter()
	r.POST("/tasks/:id/rollback", func(c *gin.Context) {
		setAuthUser(c, uuid.New().String())
		h.Rollback(c)
	})

	body := `{"comment":"missing target_node_id"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/tasks/"+uuid.New().String()+"/rollback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
