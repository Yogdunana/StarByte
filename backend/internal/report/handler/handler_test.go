package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/report/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	list          []*dto.ReportResponse
	total         int64
	page          int
	pageSize      int
	err           error
	lastUserID    uuid.UUID
	lastRequest   *dto.ListReportRequest
	lastDataScope *rbacModel.DataScopeCondition
}

func (f *fakeService) List(
	_ context.Context,
	userID uuid.UUID,
	req *dto.ListReportRequest,
	scope *rbacModel.DataScopeCondition,
) ([]*dto.ReportResponse, int64, int, int, error) {
	requestCopy := *req
	f.lastUserID = userID
	f.lastRequest = &requestCopy
	f.lastDataScope = scope
	return f.list, f.total, f.page, f.pageSize, f.err
}

func setupHandlerRouter(svc *fakeService, userID string, scope *rbacModel.DataScopeCondition) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if userID != "" || scope != nil {
		router.Use(func(c *gin.Context) {
			if userID != "" {
				c.Set(auth.ContextKeyUserID, userID)
			}
			if scope != nil {
				c.Set("data_scope_condition", scope)
			}
			c.Next()
		})
	}
	router.GET("/api/v1/reports", New(svc).List)
	return router
}

func TestListReturnsPaginatedReportsAndPassesViewer(t *testing.T) {
	userID := uuid.New()
	departmentID := uuid.New()
	scope := &rbacModel.DataScopeCondition{
		Query: "department_id = ?",
		Args:  []interface{}{departmentID},
	}
	svc := &fakeService{
		list:     []*dto.ReportResponse{{ID: uuid.New().String()}},
		total:    1,
		page:     2,
		pageSize: 10,
	}
	router := setupHandlerRouter(svc, userID.String(), scope)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/reports?page=2&page_size=10&report_type=weekly&review_status=pending&user_id="+userID.String()+"&department_id="+departmentID.String()+"&period_start=2026-09-01&period_end=2026-09-30",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var envelope struct {
		Data struct {
			List     []dto.ReportResponse `json:"list"`
			Total    int64                `json:"total"`
			Page     int                  `json:"page"`
			PageSize int                  `json:"page_size"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
	assert.Len(t, envelope.Data.List, 1)
	assert.Equal(t, int64(1), envelope.Data.Total)
	assert.Equal(t, 2, envelope.Data.Page)
	assert.Equal(t, 10, envelope.Data.PageSize)
	assert.Equal(t, userID, svc.lastUserID)
	assert.Same(t, scope, svc.lastDataScope)
	require.NotNil(t, svc.lastRequest)
	assert.Equal(t, 2, svc.lastRequest.Page)
	assert.Equal(t, 10, svc.lastRequest.PageSize)
	assert.Equal(t, "weekly", svc.lastRequest.ReportType)
	assert.Equal(t, "pending", svc.lastRequest.ReviewStatus)
	assert.Equal(t, userID.String(), svc.lastRequest.UserID)
	assert.Equal(t, departmentID.String(), svc.lastRequest.DepartmentID)
	require.NotNil(t, svc.lastRequest.PeriodStart)
	require.NotNil(t, svc.lastRequest.PeriodEnd)
}

func TestListRejectsInvalidQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "invalid enum", query: "?report_type=custom"},
		{name: "invalid page", query: "?page=-1"},
		{name: "invalid date", query: "?period_start=not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{}
			router := setupHandlerRouter(svc, uuid.New().String(), nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports"+tt.query, nil))
			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.Nil(t, svc.lastRequest)
		})
	}
}

func TestListRejectsInvalidUUIDQuery(t *testing.T) {
	svc := &fakeService{}
	router := setupHandlerRouter(svc, uuid.New().String(), nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports?user_id=not-a-uuid", nil))

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Nil(t, svc.lastRequest)
}

func TestListRequiresAuthenticatedUser(t *testing.T) {
	svc := &fakeService{}
	router := setupHandlerRouter(svc, "", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil))

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Nil(t, svc.lastRequest)
}

func TestListRejectsInvalidAuthenticatedUserID(t *testing.T) {
	svc := &fakeService{}
	router := setupHandlerRouter(svc, "not-a-uuid", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil))

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Nil(t, svc.lastRequest)
}

func TestListHandlesServiceError(t *testing.T) {
	svc := &fakeService{err: errors.New("database unavailable")}
	router := setupHandlerRouter(svc, uuid.New().String(), nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil))

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestListPassesMissingDataScopeAsNil(t *testing.T) {
	svc := &fakeService{list: []*dto.ReportResponse{}, page: 1, pageSize: 20}
	router := setupHandlerRouter(svc, uuid.New().String(), nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil))

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Nil(t, svc.lastDataScope)
}
