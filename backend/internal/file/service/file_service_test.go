package service

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/file/dto"
	"github.com/Yogdunana/StarByte/backend/internal/file/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFileRepo struct {
	mock.Mock
}

func (m *mockFileRepo) Create(ctx context.Context, file *model.File) error {
	return m.Called(ctx, file).Error(0)
}
func (m *mockFileRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.File), args.Error(1)
}
func (m *mockFileRepo) GetByIDWithUploader(ctx context.Context, id uuid.UUID) (*model.FileWithUploader, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FileWithUploader), args.Error(1)
}
func (m *mockFileRepo) List(ctx context.Context, req *dto.ListFilesRequest, ownerScope *uuid.UUID) ([]model.FileWithUploader, int64, error) {
	args := m.Called(ctx, req, ownerScope)
	return args.Get(0).([]model.FileWithUploader), args.Get(1).(int64), args.Error(2)
}
func (m *mockFileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockPermCache struct {
	mock.Mock
}

func (m *mockPermCache) GetUserPermissionsAndSuperAdmin(ctx context.Context, userID uuid.UUID) ([]string, bool, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Bool(1), args.Error(2)
}

func makePNGHeader(t *testing.T, name string) (*multipart.FileHeader, []byte) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	data := buf.Bytes()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", name)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(1<<20))
	return req.MultipartForm.File["file"][0], data
}

func TestUpload_RejectsExe(t *testing.T) {
	header, _ := makePNGHeader(t, "virus.exe")
	svc := NewFileService(&mockFileRepo{}, storage.NewMemory("starbyte"), &mockPermCache{}, "starbyte")
	_, err := svc.Upload(context.Background(), uuid.New(), header, "", false)
	requireAppError(t, err, response.CodeBadRequest, "文件类型不允许: .exe")
}

func TestUpload_ImageOK(t *testing.T) {
	repo := &mockFileRepo{}
	store := storage.NewMemory("starbyte")
	svc := NewFileService(repo, store, &mockPermCache{}, "starbyte")
	header, _ := makePNGHeader(t, "photo.png")
	userID := uuid.New()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.File")).Return(nil).Run(func(args mock.Arguments) {
		f := args.Get(1).(*model.File)
		assert.Equal(t, model.CategoryImage, f.Category)
		assert.NotEmpty(t, f.ThumbnailPath)
		assert.Equal(t, userID, *f.UploadedBy)
	})

	got, err := svc.Upload(context.Background(), userID, header, "", false)
	require.NoError(t, err)
	assert.Equal(t, "photo.png", got.OriginalName)
	assert.Equal(t, model.CategoryImage, got.Category)
	assert.NotEmpty(t, got.URL)
	assert.NotEmpty(t, got.ThumbnailURL)
	repo.AssertExpectations(t)
}

func TestDelete_OwnerOK(t *testing.T) {
	repo := &mockFileRepo{}
	store := storage.NewMemory("b")
	svc := NewFileService(repo, store, &mockPermCache{}, "b")
	userID := uuid.New()
	fileID := uuid.New()
	_ = store.Upload(context.Background(), "document/a.pdf", bytes.NewReader([]byte("x")), 1, "application/pdf")

	repo.On("GetByID", mock.Anything, fileID).Return(&model.File{
		ID: fileID, Path: "document/a.pdf", UploadedBy: &userID, CreatedAt: time.Now(),
	}, nil)
	repo.On("Delete", mock.Anything, fileID).Return(nil)

	err := svc.Delete(context.Background(), fileID, userID)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_Forbidden(t *testing.T) {
	repo := &mockFileRepo{}
	perms := &mockPermCache{}
	svc := NewFileService(repo, storage.NewMemory("b"), perms, "b")
	owner := uuid.New()
	other := uuid.New()
	fileID := uuid.New()
	repo.On("GetByID", mock.Anything, fileID).Return(&model.File{
		ID: fileID, Path: "document/a.pdf", UploadedBy: &owner,
	}, nil)
	perms.On("GetUserPermissionsAndSuperAdmin", mock.Anything, other).Return([]string{"file:read"}, false, nil)

	err := svc.Delete(context.Background(), fileID, other)
	requireAppError(t, err, response.CodeForbidden, "无权删除该文件")
}

func TestDelete_AdminOK(t *testing.T) {
	repo := &mockFileRepo{}
	perms := &mockPermCache{}
	store := storage.NewMemory("b")
	svc := NewFileService(repo, store, perms, "b")
	owner := uuid.New()
	admin := uuid.New()
	fileID := uuid.New()
	_ = store.Upload(context.Background(), "document/a.pdf", bytes.NewReader([]byte("x")), 1, "application/pdf")
	repo.On("GetByID", mock.Anything, fileID).Return(&model.File{
		ID: fileID, Path: "document/a.pdf", UploadedBy: &owner,
	}, nil)
	perms.On("GetUserPermissionsAndSuperAdmin", mock.Anything, admin).Return([]string{"file:delete"}, false, nil)
	repo.On("Delete", mock.Anything, fileID).Return(nil)

	err := svc.Delete(context.Background(), fileID, admin)
	require.NoError(t, err)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockFileRepo{}
	svc := NewFileService(repo, storage.NewMemory("b"), &mockPermCache{}, "b")
	id := uuid.New()
	repo.On("GetByIDWithUploader", mock.Anything, id).Return(nil, nil)
	_, err := svc.GetByID(context.Background(), id, uuid.New())
	requireAppError(t, err, response.CodeNotFound, "文件不存在")
}

func TestUploadBatch_TooMany(t *testing.T) {
	svc := NewFileService(&mockFileRepo{}, storage.NewMemory("b"), &mockPermCache{}, "b")
	headers := make([]*multipart.FileHeader, 11)
	_, err := svc.UploadBatch(context.Background(), uuid.New(), headers)
	requireAppError(t, err, response.CodeBadRequest, "批量上传最多 10 个文件")
}

// ---- 越权修复（Issue：文件读取/下载越权） ----

func TestPresignDownload_Forbidden(t *testing.T) {
	repo := &mockFileRepo{}
	perms := &mockPermCache{}
	svc := NewFileService(repo, storage.NewMemory("b"), perms, "b")
	owner := uuid.New()
	other := uuid.New()
	fileID := uuid.New()
	repo.On("GetByID", mock.Anything, fileID).Return(&model.File{
		ID: fileID, Path: "document/a.pdf", UploadedBy: &owner,
	}, nil)
	perms.On("GetUserPermissionsAndSuperAdmin", mock.Anything, other).Return([]string{"file:read"}, false, nil)

	_, err := svc.PresignDownload(context.Background(), fileID, other)
	requireAppError(t, err, response.CodeForbidden, "无权访问该文件")
}

func TestPresignDownload_OwnerOK(t *testing.T) {
	repo := &mockFileRepo{}
	svc := NewFileService(repo, storage.NewMemory("b"), &mockPermCache{}, "b")
	owner := uuid.New()
	fileID := uuid.New()
	repo.On("GetByID", mock.Anything, fileID).Return(&model.File{
		ID: fileID, Path: "document/a.pdf", UploadedBy: &owner,
	}, nil)

	_, err := svc.PresignDownload(context.Background(), fileID, owner)
	require.NoError(t, err)
}

func TestPresignDownload_PublicOK(t *testing.T) {
	repo := &mockFileRepo{}
	svc := NewFileService(repo, storage.NewMemory("b"), &mockPermCache{}, "b")
	owner := uuid.New()
	other := uuid.New()
	fileID := uuid.New()
	repo.On("GetByID", mock.Anything, fileID).Return(&model.File{
		ID: fileID, Path: "document/a.pdf", UploadedBy: &owner, IsPublic: true,
	}, nil)

	_, err := svc.PresignDownload(context.Background(), fileID, other)
	require.NoError(t, err)
}

func TestPresignDownload_SuperOK(t *testing.T) {
	repo := &mockFileRepo{}
	perms := &mockPermCache{}
	svc := NewFileService(repo, storage.NewMemory("b"), perms, "b")
	owner := uuid.New()
	admin := uuid.New()
	fileID := uuid.New()
	repo.On("GetByID", mock.Anything, fileID).Return(&model.File{
		ID: fileID, Path: "document/a.pdf", UploadedBy: &owner,
	}, nil)
	perms.On("GetUserPermissionsAndSuperAdmin", mock.Anything, admin).Return([]string{}, true, nil)

	_, err := svc.PresignDownload(context.Background(), fileID, admin)
	require.NoError(t, err)
}

func TestGetByID_Forbidden(t *testing.T) {
	repo := &mockFileRepo{}
	perms := &mockPermCache{}
	svc := NewFileService(repo, storage.NewMemory("b"), perms, "b")
	owner := uuid.New()
	other := uuid.New()
	fileID := uuid.New()
	repo.On("GetByIDWithUploader", mock.Anything, fileID).Return(&model.FileWithUploader{
		File: model.File{ID: fileID, Path: "document/a.pdf", UploadedBy: &owner},
	}, nil)
	perms.On("GetUserPermissionsAndSuperAdmin", mock.Anything, other).Return([]string{"file:read"}, false, nil)

	_, err := svc.GetByID(context.Background(), fileID, other)
	requireAppError(t, err, response.CodeForbidden, "无权访问该文件")
}

func TestGetByID_OwnerOK(t *testing.T) {
	repo := &mockFileRepo{}
	svc := NewFileService(repo, storage.NewMemory("b"), &mockPermCache{}, "b")
	owner := uuid.New()
	fileID := uuid.New()
	repo.On("GetByIDWithUploader", mock.Anything, fileID).Return(&model.FileWithUploader{
		File: model.File{ID: fileID, Path: "document/a.pdf", UploadedBy: &owner},
	}, nil)

	_, err := svc.GetByID(context.Background(), fileID, owner)
	require.NoError(t, err)
}

func TestList_ScopeForMember(t *testing.T) {
	repo := &mockFileRepo{}
	perms := &mockPermCache{}
	svc := NewFileService(repo, storage.NewMemory("b"), perms, "b")
	userID := uuid.New()
	// 会员只有 file:read，没有 file:read:all —— 应被收窄到「自己的 + 公开的」。
	perms.On("GetUserPermissionsAndSuperAdmin", mock.Anything, userID).Return([]string{"file:read"}, false, nil)
	repo.On("List", mock.Anything, mock.Anything, mock.MatchedBy(func(s *uuid.UUID) bool {
		return s != nil && *s == userID
	})).Return([]model.FileWithUploader{}, int64(0), nil)

	_, _, err := svc.List(context.Background(), &dto.ListFilesRequest{}, userID)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestList_ScopeForAllReader(t *testing.T) {
	repo := &mockFileRepo{}
	perms := &mockPermCache{}
	svc := NewFileService(repo, storage.NewMemory("b"), perms, "b")
	userID := uuid.New()
	perms.On("GetUserPermissionsAndSuperAdmin", mock.Anything, userID).Return([]string{"file:read:all"}, false, nil)
	repo.On("List", mock.Anything, mock.Anything, mock.MatchedBy(func(s *uuid.UUID) bool {
		return s == nil
	})).Return([]model.FileWithUploader{}, int64(0), nil)

	_, _, err := svc.List(context.Background(), &dto.ListFilesRequest{}, userID)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}
