package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrace_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockAuditService{}
	h := NewAuditHandler(svc)
	r := gin.New()
	r.GET("/system/audit-logs/traces/:entity_type/:entity_id", h.Trace)
	eid := uuid.New().String()
	svc.On("Trace", mock.Anything, "user", eid, mock.Anything).
		Return([]dto.AuditTraceItem{{ID: uuid.New().String(), Action: "UPDATE"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system/audit-logs/traces/user/"+eid+"?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReport_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockAuditService{}
	h := NewAuditHandler(svc)
	r := gin.New()
	r.GET("/system/audit-logs/reports", h.Report)
	svc.On("Report", mock.Anything, mock.Anything).
		Return(&dto.ReportResponse{Total: 3, Note: "n"}, []byte(nil), "", nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system/audit-logs/reports", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(0), body["code"])
}

func TestArchives_ListAndPull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockAuditService{}
	h := NewAuditHandler(svc)
	r := gin.New()
	r.GET("/system/audit-logs/archives", h.Archives)
	svc.On("ListArchives", mock.Anything, mock.Anything).
		Return([]dto.ArchiveListItem{{ID: uuid.New().String()}}, int64(1), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system/audit-logs/archives?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	aid := uuid.New()
	svc.On("PullArchive", mock.Anything, aid, mock.Anything).
		Return(&dto.ArchivePullResponse{Total: 0, Page: 1, PageSize: 20}, nil)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/system/audit-logs/archives?id="+aid.String(), nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestArchives_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockAuditService{}
	h := NewAuditHandler(svc)
	r := gin.New()
	r.GET("/system/audit-logs/archives", h.Archives)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system/audit-logs/archives?id=not-a-uuid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
