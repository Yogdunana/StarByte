package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuditService struct {
	mock.Mock
}

func (m *mockAuditService) Log(ctx context.Context, entry *model.AuditEntry) error {
	return m.Called(ctx, entry).Error(0)
}
func (m *mockAuditService) LogAsync(ctx context.Context, entry *model.AuditEntry) error {
	return m.Called(ctx, entry).Error(0)
}
func (m *mockAuditService) Query(ctx context.Context, req *dto.ListAuditLogRequest) ([]dto.AuditLogListResponse, int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]dto.AuditLogListResponse), args.Get(1).(int64), args.Error(2)
}
func (m *mockAuditService) GetByID(ctx context.Context, id uuid.UUID) (*dto.AuditLogResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuditLogResponse), args.Error(1)
}
func (m *mockAuditService) Export(ctx context.Context, req *dto.ExportAuditLogRequest) ([]byte, string, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}
func (m *mockAuditService) Archive(ctx context.Context, beforeDays int) (*dto.ArchiveResponse, error) {
	args := m.Called(ctx, beforeDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ArchiveResponse), args.Error(1)
}
func (m *mockAuditService) Trace(ctx context.Context, entityType, entityID string, req *dto.TraceQueryRequest) ([]dto.AuditTraceItem, int64, error) {
	args := m.Called(ctx, entityType, entityID, req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]dto.AuditTraceItem), args.Get(1).(int64), args.Error(2)
}
func (m *mockAuditService) Report(ctx context.Context, req *dto.ReportRequest) (*dto.ReportResponse, []byte, string, error) {
	args := m.Called(ctx, req)
	var report *dto.ReportResponse
	if args.Get(0) != nil {
		report = args.Get(0).(*dto.ReportResponse)
	}
	var data []byte
	if args.Get(1) != nil {
		data = args.Get(1).([]byte)
	}
	return report, data, args.String(2), args.Error(3)
}
func (m *mockAuditService) ListArchives(ctx context.Context, req *dto.ArchiveListRequest) ([]dto.ArchiveListItem, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]dto.ArchiveListItem), args.Get(1).(int64), args.Error(2)
}
func (m *mockAuditService) PullArchive(ctx context.Context, id uuid.UUID, req *dto.ArchiveListRequest) (*dto.ArchivePullResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ArchivePullResponse), args.Error(1)
}

func TestGetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockAuditService{}
	h := NewAuditHandler(svc)
	r := gin.New()
	r.GET("/system/audit-logs/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system/audit-logs/not-a-uuid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestList_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockAuditService{}
	h := NewAuditHandler(svc)
	r := gin.New()
	r.GET("/system/audit-logs", h.List)
	svc.On("Query", mock.Anything, mock.Anything).Return([]dto.AuditLogListResponse{
		{ID: uuid.New().String(), Action: "CREATE"},
	}, int64(1), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system/audit-logs?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(0), body["code"])
}
