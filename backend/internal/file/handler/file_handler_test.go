package handler

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/file/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFileService struct {
	mock.Mock
}

func (m *mockFileService) Upload(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, category string, isPublic bool) (*dto.FileUploadResponse, error) {
	args := m.Called(ctx, userID, header, category, isPublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.FileUploadResponse), args.Error(1)
}
func (m *mockFileService) UploadBatch(ctx context.Context, userID uuid.UUID, headers []*multipart.FileHeader) ([]*dto.FileUploadResponse, error) {
	args := m.Called(ctx, userID, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.FileUploadResponse), args.Error(1)
}
func (m *mockFileService) List(ctx context.Context, req *dto.ListFilesRequest) ([]*dto.FileListItem, int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*dto.FileListItem), args.Get(1).(int64), args.Error(2)
}
func (m *mockFileService) GetByID(ctx context.Context, id uuid.UUID) (*dto.FileDetailResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.FileDetailResponse), args.Error(1)
}
func (m *mockFileService) PresignDownload(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}
func (m *mockFileService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func TestGetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockFileService{}
	h := NewFileHandler(svc)
	r := gin.New()
	r.GET("/files/:id", h.GetByID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/not-a-uuid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDownload_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockFileService{}
	h := NewFileHandler(svc)
	r := gin.New()
	r.GET("/files/:id/download", h.Download)
	id := uuid.New()
	svc.On("PresignDownload", mock.Anything, id).Return("https://minio.example/obj?X-Amz-Expires=3600", nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/"+id.String()+"/download", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "X-Amz-Expires=3600")
}

func TestDelete_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockFileService{}
	h := NewFileHandler(svc)
	r := gin.New()
	r.DELETE("/files/:id", func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, uuid.New().String())
		h.Delete(c)
	})
	id := uuid.New()
	svc.On("Delete", mock.Anything, id, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/"+id.String(), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 0, body.Code)
	assert.Equal(t, "删除成功", body.Message)
}

func TestList_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockFileService{}
	h := NewFileHandler(svc)
	r := gin.New()
	r.GET("/files", h.List)
	svc.On("List", mock.Anything, mock.Anything).Return([]*dto.FileListItem{
		{ID: uuid.New().String(), Filename: "a.png"},
	}, int64(1), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files?page=1&page_size=10&category=image", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
