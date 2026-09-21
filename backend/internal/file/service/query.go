package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/file/dto"
	"github.com/Yogdunana/StarByte/backend/internal/file/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *fileService) List(ctx context.Context, req *dto.ListFilesRequest, userID uuid.UUID) ([]*dto.FileListItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	// 非提权用户只能看到「自己的 + 公开的」，由 repo 强制过滤。
	var scope *uuid.UUID
	if !s.hasElevatedFilePerm(ctx, userID) {
		scope = &userID
	}
	rows, total, err := s.fileRepo.List(ctx, req, scope)
	if err != nil {
		logger.Error("list files failed", zap.Error(err))
		return nil, 0, response.NewError(response.CodeInternalError, "查询文件列表失败")
	}
	list := make([]*dto.FileListItem, 0, len(rows))
	for i := range rows {
		url, thumb := s.presignPair(ctx, rows[i].Path, rows[i].ThumbnailPath)
		list = append(list, toListItem(&rows[i], url, thumb))
	}
	return list, total, nil
}

func (s *fileService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.FileDetailResponse, error) {
	row, err := s.fileRepo.GetByIDWithUploader(ctx, id)
	if err != nil {
		logger.Error("get file failed", zap.Error(err), zap.String("id", id.String()))
		return nil, response.NewError(response.CodeInternalError, "查询文件失败")
	}
	if row == nil {
		return nil, response.NewError(response.CodeNotFound, "文件不存在")
	}
	if !s.canAccessFile(ctx, &row.File, userID) {
		return nil, response.NewForbiddenError("无权访问该文件")
	}
	url, thumb := s.presignPair(ctx, row.Path, row.ThumbnailPath)
	return toDetail(row, url, thumb), nil
}

func (s *fileService) PresignDownload(ctx context.Context, id, userID uuid.UUID) (string, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return "", response.NewError(response.CodeInternalError, "查询文件失败")
	}
	if file == nil {
		return "", response.NewError(response.CodeNotFound, "文件不存在")
	}
	if !s.canAccessFile(ctx, file, userID) {
		return "", response.NewForbiddenError("无权访问该文件")
	}
	if s.store == nil {
		return "", response.NewError(response.CodeInternalError, "对象存储未配置")
	}
	url, err := s.store.PresignedURL(ctx, file.Path, 0)
	if err != nil {
		logger.Error("presign download failed", zap.Error(err), zap.String("id", id.String()))
		return "", response.NewError(response.CodeInternalError, "生成下载地址失败")
	}
	return url, nil
}

func (s *fileService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return response.NewError(response.CodeInternalError, "查询文件失败")
	}
	if file == nil {
		return response.NewError(response.CodeNotFound, "文件不存在")
	}
	if !s.canDelete(ctx, file, userID) {
		return response.NewForbiddenError("无权删除该文件")
	}
	if s.store != nil {
		if err := s.store.Delete(ctx, file.Path); err != nil {
			logger.Error("delete object failed", zap.Error(err), zap.String("object", file.Path))
			return response.NewError(response.CodeInternalError, "删除存储对象失败")
		}
		if file.ThumbnailPath != "" {
			if err := s.store.Delete(ctx, file.ThumbnailPath); err != nil {
				logger.Error("delete thumbnail failed", zap.Error(err), zap.String("object", file.ThumbnailPath))
			}
		}
	}
	if err := s.fileRepo.Delete(ctx, id); err != nil {
		logger.Error("delete file metadata failed", zap.Error(err), zap.String("id", id.String()))
		return response.NewError(response.CodeInternalError, "删除文件记录失败")
	}
	return nil
}

// canAccessFile 判断 userID 是否有权读取 file。
// 上传者本人、公开文件、超管、持 file:read:all 或 file:delete 者可读。
func (s *fileService) canAccessFile(ctx context.Context, file *model.File, userID uuid.UUID) bool {
	if file.UploadedBy != nil && *file.UploadedBy == userID {
		return true
	}
	if file.IsPublic {
		return true
	}
	return s.hasElevatedFilePerm(ctx, userID)
}

// hasElevatedFilePerm 判断是否为超管或持文件全量权限。
// 注：file:delete 也放行是刻意兼容——该权限已存在于 canDelete，授予的是管理层角色；
// 这样不依赖重新播种 RBAC 也不会让管理层突然看不到文件。
func (s *fileService) hasElevatedFilePerm(ctx context.Context, userID uuid.UUID) bool {
	if s.permCache == nil {
		return false
	}
	perms, isSuper, err := s.permCache.GetUserPermissionsAndSuperAdmin(ctx, userID)
	if err != nil {
		logger.Error("check file permission failed", zap.Error(err), zap.String("user_id", userID.String()))
		return false
	}
	if isSuper {
		return true
	}
	for _, p := range perms {
		if p == permFileReadAll || p == permFileDelete || p == "*" {
			return true
		}
	}
	return false
}

func (s *fileService) canDelete(ctx context.Context, file *model.File, userID uuid.UUID) bool {
	if file.UploadedBy != nil && *file.UploadedBy == userID {
		return true
	}
	if s.permCache == nil {
		return false
	}
	perms, isSuper, err := s.permCache.GetUserPermissionsAndSuperAdmin(ctx, userID)
	if err != nil {
		logger.Error("check delete permission failed", zap.Error(err), zap.String("user_id", userID.String()))
		return false
	}
	if isSuper {
		return true
	}
	for _, p := range perms {
		if p == permFileDelete || p == "*" {
			return true
		}
	}
	return false
}
