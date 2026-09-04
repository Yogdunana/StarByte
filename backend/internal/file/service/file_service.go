package service

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"path"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/file/dto"
	"github.com/Yogdunana/StarByte/backend/internal/file/model"
	"github.com/Yogdunana/StarByte/backend/internal/file/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const permFileDelete = "file:delete"

// FileService 文件管理服务
type FileService interface {
	Upload(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, category string, isPublic bool) (*dto.FileUploadResponse, error)
	UploadBatch(ctx context.Context, userID uuid.UUID, headers []*multipart.FileHeader) ([]*dto.FileUploadResponse, error)
	List(ctx context.Context, req *dto.ListFilesRequest) ([]*dto.FileListItem, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.FileDetailResponse, error)
	PresignDownload(ctx context.Context, id uuid.UUID) (string, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type permissionChecker interface {
	GetUserPermissionsAndSuperAdmin(ctx context.Context, userID uuid.UUID) ([]string, bool, error)
}

type fileService struct {
	fileRepo  repo.FileRepo
	store     storage.ObjectStorage
	permCache permissionChecker
	bucket    string
}

// NewFileService 创建文件服务
func NewFileService(
	fileRepo repo.FileRepo,
	store storage.ObjectStorage,
	permCache permissionChecker,
	bucket string,
) FileService {
	return &fileService{
		fileRepo:  fileRepo,
		store:     store,
		permCache: permCache,
		bucket:    bucket,
	}
}

func (s *fileService) Upload(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, category string, isPublic bool) (*dto.FileUploadResponse, error) {
	if s.store == nil {
		return nil, response.NewError(response.CodeInternalError, "对象存储未配置")
	}
	if header == nil {
		return nil, response.NewError(response.CodeBadRequest, "请选择要上传的文件")
	}
	validated, err := validateUpload(header.Filename, header.Size, strings.TrimSpace(category))
	if err != nil {
		return nil, err
	}

	src, err := header.Open()
	if err != nil {
		return nil, response.NewError(response.CodeInternalError, "读取上传文件失败")
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return nil, response.NewError(response.CodeInternalError, "读取上传文件失败")
	}

	fileID := uuid.New()
	objectName := path.Join(validated.Category, fileID.String()+validated.Ext)
	if err := s.store.Upload(ctx, objectName, bytes.NewReader(data), int64(len(data)), validated.MimeType); err != nil {
		logger.Error("upload object failed", zap.Error(err), zap.String("object", objectName))
		return nil, response.NewError(response.CodeInternalError, "文件上传失败")
	}

	thumbPath := s.uploadThumbnail(ctx, validated.Category, fileID, data)

	rec := &model.File{
		ID:            fileID,
		Name:          fileID.String() + validated.Ext,
		OriginalName:  header.Filename,
		Path:          objectName,
		Size:          int64(len(data)),
		MimeType:      validated.MimeType,
		StorageType:   model.StorageTypeMinIO,
		Bucket:        s.bucket,
		UploadedBy:    &userID,
		Category:      validated.Category,
		IsPublic:      isPublic,
		ThumbnailPath: thumbPath,
	}
	if err := s.fileRepo.Create(ctx, rec); err != nil {
		_ = s.store.Delete(ctx, objectName)
		if thumbPath != "" {
			_ = s.store.Delete(ctx, thumbPath)
		}
		logger.Error("save file metadata failed", zap.Error(err))
		return nil, response.NewError(response.CodeInternalError, "保存文件元数据失败")
	}

	url, thumbURL := s.presignPair(ctx, rec.Path, rec.ThumbnailPath)
	return toUploadResponse(rec, url, thumbURL), nil
}

func (s *fileService) UploadBatch(ctx context.Context, userID uuid.UUID, headers []*multipart.FileHeader) ([]*dto.FileUploadResponse, error) {
	if len(headers) == 0 {
		return nil, response.NewError(response.CodeBadRequest, "请选择要上传的文件")
	}
	if len(headers) > maxBatchFiles {
		return nil, response.NewError(response.CodeBadRequest, "批量上传最多 10 个文件")
	}
	results := make([]*dto.FileUploadResponse, 0, len(headers))
	for _, header := range headers {
		item, err := s.Upload(ctx, userID, header, "", false)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

func (s *fileService) uploadThumbnail(ctx context.Context, category string, fileID uuid.UUID, data []byte) string {
	if category != model.CategoryImage {
		return ""
	}
	thumb, err := generateThumbnail(data)
	if err != nil {
		logger.Error("generate thumbnail failed", zap.Error(err), zap.String("file_id", fileID.String()))
		return ""
	}
	thumbPath := path.Join(category, "thumb_"+fileID.String()+".jpg")
	if err := s.store.Upload(ctx, thumbPath, bytes.NewReader(thumb), int64(len(thumb)), "image/jpeg"); err != nil {
		logger.Error("upload thumbnail failed", zap.Error(err), zap.String("object", thumbPath))
		return ""
	}
	return thumbPath
}

func (s *fileService) presignPair(ctx context.Context, objectName, thumbPath string) (string, string) {
	url, err := s.store.PresignedURL(ctx, objectName, storage.PresignExpiry)
	if err != nil {
		logger.Error("presign url failed", zap.Error(err), zap.String("object", objectName))
		url = ""
	}
	thumbURL := ""
	if thumbPath != "" {
		thumbURL, err = s.store.PresignedURL(ctx, thumbPath, storage.PresignExpiry)
		if err != nil {
			logger.Error("presign thumbnail failed", zap.Error(err), zap.String("object", thumbPath))
			thumbURL = ""
		}
	}
	return url, thumbURL
}
