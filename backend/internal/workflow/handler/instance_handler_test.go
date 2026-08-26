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

// --- Mock InstanceService ---

type mockInstService struct {
	startFunc       func(ctx context.Context, defID uuid.UUID, businessKey, businessType string, initiatorID uuid.UUID, variables map[string]interface{}) (*model.FlowInstance, error)
	getByIDFunc     func(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error)
	listFunc        func(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error)
	terminateFunc   func(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error
	suspendFunc     func(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error
	resumeFunc      func(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error
	listHistoryFunc func(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error)
}

func (m *mockInstService) Start(ctx context.Context, defID uuid.UUID, businessKey, businessType string, initiatorID uuid.UUID, variables map[string]interface{}) (*model.FlowInstance, error) {
	if m.startFunc != nil {
		return m.startFunc(ctx, defID, businessKey, businessType, initiatorID, variables)
	}
	return &model.FlowInstance{ID: uuid.New(), DefinitionID: defID, InitiatorID: initiatorID}, nil
}

func (m *mockInstService) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockInstService) List(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, page, pageSize, status, definitionID, initiatorID)
	}
	return []model.FlowInstance{}, 0, nil
}

func (m *mockInstService) Terminate(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	if m.terminateFunc != nil {
		return m.terminateFunc(ctx, instanceID, operatorID, reason)
	}
	return nil
}

func (m *mockInstService) Suspend(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	if m.suspendFunc != nil {
		return m.suspendFunc(ctx, instanceID, operatorID, reason)
	}
	return nil
}

func (m *mockInstService) Resume(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error {
	if m.resumeFunc != nil {
		return m.resumeFunc(ctx, instanceID, operatorID)
	}
	return nil
}

func (m *mockInstService) ListHistory(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error) {
	if m.listHistoryFunc != nil {
		return m.listHistoryFunc(ctx, instanceID)
	}
	return []model.FlowHistory{}, nil
}

// Compile-time check.
var _ service.InstanceService = (*mockInstService)(nil)

// --- InstanceHandler Tests ---

func TestInstanceHandler_Start_Success(t *testing.T) {
	userID := uuid.New()
	defID := uuid.New()
	var capturedDefID uuid.UUID
	var capturedInitiatorID uuid.UUID

	svc := &mockInstService{
		startFunc: func(ctx context.Context, d uuid.UUID, bk, bt string, initID uuid.UUID, vars map[string]interface{}) (*model.FlowInstance, error) {
			capturedDefID = d
			capturedInitiatorID = initID
			return &model.FlowInstance{ID: uuid.New(), DefinitionID: d, InitiatorID: initID, Status: 0}, nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.POST("/instances", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Start(c)
	})

	body := `{"definition_id":"` + defID.String() + `","business_key":"ORDER-001","business_type":"order"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/instances", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, defID, capturedDefID)
	assert.Equal(t, userID, capturedInitiatorID)
}

func TestInstanceHandler_Start_MissingDefinitionID(t *testing.T) {
	svc := &mockInstService{}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.POST("/instances", func(c *gin.Context) {
		setAuthUser(c, uuid.New().String())
		h.Start(c)
	})

	body := `{"business_key":"ORDER-001"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/instances", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInstanceHandler_List_Success(t *testing.T) {
	svc := &mockInstService{
		listFunc: func(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error) {
			return []model.FlowInstance{
				{ID: uuid.New(), Status: 0},
				{ID: uuid.New(), Status: 1},
			}, 2, nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.GET("/instances", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/instances?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestInstanceHandler_List_WithFilters(t *testing.T) {
	var capturedStatus *int
	var capturedDefID *uuid.UUID
	var capturedInitID *uuid.UUID

	svc := &mockInstService{
		listFunc: func(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error) {
			capturedStatus = status
			capturedDefID = definitionID
			capturedInitID = initiatorID
			return []model.FlowInstance{}, 0, nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.GET("/instances", h.List)

	defID := uuid.New()
	initID := uuid.New()
	path := "/instances?status=1&definition_id=" + defID.String() + "&initiator_id=" + initID.String()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", path, nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedStatus)
	assert.Equal(t, 1, *capturedStatus)
	require.NotNil(t, capturedDefID)
	assert.Equal(t, defID, *capturedDefID)
	require.NotNil(t, capturedInitID)
	assert.Equal(t, initID, *capturedInitID)
}

func TestInstanceHandler_GetByID_Success(t *testing.T) {
	instID := uuid.New()
	svc := &mockInstService{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error) {
			return &model.FlowInstance{ID: id, Status: 0}, nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.GET("/instances/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/instances/"+instID.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestInstanceHandler_GetByID_InvalidUUID(t *testing.T) {
	svc := &mockInstService{}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.GET("/instances/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/instances/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInstanceHandler_Terminate_Success(t *testing.T) {
	instID := uuid.New()
	userID := uuid.New()
	var capturedReason string
	var capturedOperatorID uuid.UUID

	svc := &mockInstService{
		terminateFunc: func(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
			capturedReason = reason
			capturedOperatorID = operatorID
			return nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.POST("/instances/:id/terminate", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Terminate(c)
	})

	body := `{"reason":"不再需要"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/instances/"+instID.String()+"/terminate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "不再需要", capturedReason)
	assert.Equal(t, userID, capturedOperatorID)
}

func TestInstanceHandler_Terminate_MissingReason(t *testing.T) {
	svc := &mockInstService{}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.POST("/instances/:id/terminate", func(c *gin.Context) {
		setAuthUser(c, uuid.New().String())
		h.Terminate(c)
	})

	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/instances/"+uuid.New().String()+"/terminate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInstanceHandler_Suspend_Success(t *testing.T) {
	instID := uuid.New()
	userID := uuid.New()
	var capturedReason string

	svc := &mockInstService{
		suspendFunc: func(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
			capturedReason = reason
			return nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.POST("/instances/:id/suspend", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Suspend(c)
	})

	body := `{"reason":"等待审批"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/instances/"+instID.String()+"/suspend", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "等待审批", capturedReason)
}

func TestInstanceHandler_Resume_Success(t *testing.T) {
	instID := uuid.New()
	userID := uuid.New()
	called := false

	svc := &mockInstService{
		resumeFunc: func(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error {
			called = true
			assert.Equal(t, instID, instanceID)
			assert.Equal(t, userID, operatorID)
			return nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.POST("/instances/:id/resume", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Resume(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/instances/"+instID.String()+"/resume", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
}

func TestInstanceHandler_ListHistory_Success(t *testing.T) {
	instID := uuid.New()
	svc := &mockInstService{
		listHistoryFunc: func(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error) {
			return []model.FlowHistory{
				{ID: uuid.New(), InstanceID: instanceID, NodeID: "n1", Action: "start"},
				{ID: uuid.New(), InstanceID: instanceID, NodeID: "n2", Action: "approve"},
			}, nil
		},
	}
	h := NewInstanceHandler(svc)

	r := setupRouter()
	r.GET("/instances/:id/history", h.ListHistory)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/instances/"+instID.String()+"/history", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)
}
