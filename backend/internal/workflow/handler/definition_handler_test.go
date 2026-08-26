package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/dto"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock DefinitionService ---

type mockDefService struct {
	createFunc      func(ctx context.Context, req *dto.CreateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error)
	getByIDFunc     func(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error)
	updateFunc      func(ctx context.Context, id uuid.UUID, req *dto.UpdateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error)
	deleteFunc      func(ctx context.Context, id uuid.UUID) error
	listFunc        func(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error)
	publishFunc     func(ctx context.Context, id uuid.UUID, req *dto.PublishDefinitionRequest, userID uuid.UUID) (*model.FlowDefinitionVersion, error)
	listVersionsFunc func(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error)
	getVersionFunc  func(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error)
}

func (m *mockDefService) Create(ctx context.Context, req *dto.CreateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req, userID)
	}
	return &model.FlowDefinition{ID: uuid.New(), Key: req.Key, Name: req.Name, Category: req.Category}, nil
}

func (m *mockDefService) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockDefService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, req, userID)
	}
	return &model.FlowDefinition{ID: id, Name: req.Name, Description: req.Description}, nil
}

func (m *mockDefService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockDefService) List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, page, pageSize, keyword, category, status)
	}
	return []model.FlowDefinition{}, 0, nil
}

func (m *mockDefService) Publish(ctx context.Context, id uuid.UUID, req *dto.PublishDefinitionRequest, userID uuid.UUID) (*model.FlowDefinitionVersion, error) {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, id, req, userID)
	}
	return &model.FlowDefinitionVersion{ID: uuid.New(), DefinitionID: id, Version: 1}, nil
}

func (m *mockDefService) ListVersions(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error) {
	if m.listVersionsFunc != nil {
		return m.listVersionsFunc(ctx, definitionID)
	}
	return []model.FlowDefinitionVersion{}, nil
}

func (m *mockDefService) GetVersionByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error) {
	if m.getVersionFunc != nil {
		return m.getVersionFunc(ctx, id)
	}
	return nil, nil
}

// Compile-time check.
var _ service.DefinitionService = (*mockDefService)(nil)

// --- Test Helpers ---

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setAuthUser(c *gin.Context, userID string) {
	c.Set(auth.ContextKeyUserID, userID)
}

// --- DefinitionHandler Tests ---

func TestDefinitionHandler_List(t *testing.T) {
	svc := &mockDefService{
		listFunc: func(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error) {
			return []model.FlowDefinition{
				{ID: uuid.New(), Key: "flow1", Name: "Flow 1"},
				{ID: uuid.New(), Key: "flow2", Name: "Flow 2"},
			}, 2, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.GET("/definitions", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/definitions?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 200, int(resp["code"].(float64)))
}

func TestDefinitionHandler_List_WithFilters(t *testing.T) {
	var capturedKeyword, capturedCategory string
	var capturedStatus *int

	svc := &mockDefService{
		listFunc: func(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error) {
			capturedKeyword = keyword
			capturedCategory = category
			capturedStatus = status
			return []model.FlowDefinition{}, 0, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.GET("/definitions", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/definitions?keyword=leave&category=hr&status=1", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "leave", capturedKeyword)
	assert.Equal(t, "hr", capturedCategory)
	require.NotNil(t, capturedStatus)
	assert.Equal(t, 1, *capturedStatus)
}

func TestDefinitionHandler_GetByID_Success(t *testing.T) {
	defID := uuid.New()
	svc := &mockDefService{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error) {
			return &model.FlowDefinition{ID: id, Key: "test", Name: "Test Flow"}, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.GET("/definitions/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/definitions/"+defID.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "test", data["key"])
	assert.Equal(t, "Test Flow", data["name"])
}

func TestDefinitionHandler_GetByID_InvalidUUID(t *testing.T) {
	svc := &mockDefService{}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.GET("/definitions/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/definitions/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDefinitionHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	var capturedReq *dto.CreateDefinitionRequest
	var capturedUserID uuid.UUID

	svc := &mockDefService{
		createFunc: func(ctx context.Context, req *dto.CreateDefinitionRequest, uid uuid.UUID) (*model.FlowDefinition, error) {
			capturedReq = req
			capturedUserID = uid
			return &model.FlowDefinition{ID: uuid.New(), Key: req.Key, Name: req.Name, Category: req.Category, CreatedBy: &uid}, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.POST("/definitions", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Create(c)
	})

	body := `{"key":"leave-approval","name":"请假审批","category":"hr"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/definitions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "leave-approval", capturedReq.Key)
	assert.Equal(t, userID, capturedUserID)
}

func TestDefinitionHandler_Create_MissingFields(t *testing.T) {
	svc := &mockDefService{}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.POST("/definitions", func(c *gin.Context) {
		setAuthUser(c, uuid.New().String())
		h.Create(c)
	})

	body := `{"description":"missing key and name"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/definitions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDefinitionHandler_Update_Success(t *testing.T) {
	defID := uuid.New()
	userID := uuid.New()

	svc := &mockDefService{
		updateFunc: func(ctx context.Context, id uuid.UUID, req *dto.UpdateDefinitionRequest, uid uuid.UUID) (*model.FlowDefinition, error) {
			return &model.FlowDefinition{ID: id, Name: req.Name, Description: req.Description, UpdatedBy: &uid}, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.PUT("/definitions/:id", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Update(c)
	})

	body := `{"name":"New Name","description":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/definitions/"+defID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDefinitionHandler_Delete_Success(t *testing.T) {
	defID := uuid.New()
	deleted := false
	svc := &mockDefService{
		deleteFunc: func(ctx context.Context, id uuid.UUID) error {
			deleted = true
			assert.Equal(t, defID, id)
			return nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.DELETE("/definitions/:id", h.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/definitions/"+defID.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, deleted)
}

func TestDefinitionHandler_Publish_Success(t *testing.T) {
	defID := uuid.New()
	userID := uuid.New()

	svc := &mockDefService{
		publishFunc: func(ctx context.Context, id uuid.UUID, req *dto.PublishDefinitionRequest, uid uuid.UUID) (*model.FlowDefinitionVersion, error) {
			return &model.FlowDefinitionVersion{ID: uuid.New(), DefinitionID: id, Version: 1, PublishedBy: &uid}, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.POST("/definitions/:id/publish", func(c *gin.Context) {
		setAuthUser(c, userID.String())
		h.Publish(c)
	})

	body := `{"graph_data":{"nodes":[],"edges":[]}}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/definitions/"+defID.String()+"/publish", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDefinitionHandler_Publish_MissingGraphData(t *testing.T) {
	svc := &mockDefService{}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.POST("/definitions/:id/publish", func(c *gin.Context) {
		setAuthUser(c, uuid.New().String())
		h.Publish(c)
	})

	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/definitions/"+uuid.New().String()+"/publish", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDefinitionHandler_ListVersions_Success(t *testing.T) {
	defID := uuid.New()
	svc := &mockDefService{
		listVersionsFunc: func(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error) {
			return []model.FlowDefinitionVersion{
				{ID: uuid.New(), DefinitionID: definitionID, Version: 1},
				{ID: uuid.New(), DefinitionID: definitionID, Version: 2},
			}, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.GET("/definitions/:id/versions", h.ListVersions)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/definitions/"+defID.String()+"/versions", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestDefinitionHandler_GetVersionByID_Success(t *testing.T) {
	defID := uuid.New()
	verID := uuid.New()
	svc := &mockDefService{
		getVersionFunc: func(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error) {
			return &model.FlowDefinitionVersion{ID: id, DefinitionID: defID, Version: 1}, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.GET("/definitions/:id/versions/:versionId", h.GetVersionByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/definitions/"+defID.String()+"/versions/"+verID.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDefinitionHandler_GetVersionByID_WrongDefinition(t *testing.T) {
	defID := uuid.New()
	verID := uuid.New()
	otherDefID := uuid.New()
	svc := &mockDefService{
		getVersionFunc: func(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error) {
			return &model.FlowDefinitionVersion{ID: id, DefinitionID: otherDefID, Version: 1}, nil
		},
	}
	h := NewDefinitionHandler(svc)

	r := setupRouter()
	r.GET("/definitions/:id/versions/:versionId", h.GetVersionByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/definitions/"+defID.String()+"/versions/"+verID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
