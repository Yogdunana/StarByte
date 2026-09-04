package service

import (
	"github.com/Yogdunana/StarByte/backend/internal/file/dto"
	"github.com/Yogdunana/StarByte/backend/internal/file/model"
)

func toUploadResponse(f *model.File, url, thumbURL string) *dto.FileUploadResponse {
	return &dto.FileUploadResponse{
		ID:           f.ID.String(),
		Filename:     f.Name,
		Name:         f.Name,
		OriginalName: f.OriginalName,
		FileSize:     f.Size,
		Size:         f.Size,
		MimeType:     f.MimeType,
		Category:     f.Category,
		URL:          url,
		ThumbnailURL: thumbURL,
		UploadedAt:   f.CreatedAt,
	}
}

func toListItem(row *model.FileWithUploader, url, thumbURL string) *dto.FileListItem {
	uploaderID, uploaderName := uploaderFields(row)
	return &dto.FileListItem{
		ID:           row.ID.String(),
		Filename:     row.Name,
		Name:         row.Name,
		OriginalName: row.OriginalName,
		FileSize:     row.Size,
		Size:         row.Size,
		MimeType:     row.MimeType,
		Category:     row.Category,
		URL:          url,
		ThumbnailURL: thumbURL,
		Uploader:     dto.UploaderInfo{ID: uploaderID, Name: uploaderName},
		UploaderID:   uploaderID,
		UploaderName: uploaderName,
		CreatedAt:    row.CreatedAt,
	}
}

func toDetail(row *model.FileWithUploader, url, thumbURL string) *dto.FileDetailResponse {
	uploaderID, uploaderName := uploaderFields(row)
	return &dto.FileDetailResponse{
		ID:           row.ID.String(),
		Filename:     row.Name,
		Name:         row.Name,
		OriginalName: row.OriginalName,
		FileSize:     row.Size,
		Size:         row.Size,
		MimeType:     row.MimeType,
		Category:     row.Category,
		StorageType:  row.StorageType,
		Bucket:       row.Bucket,
		Path:         row.Path,
		URL:          url,
		ThumbnailURL: thumbURL,
		IsPublic:     row.IsPublic,
		Uploader:     dto.UploaderInfo{ID: uploaderID, Name: uploaderName},
		UploaderID:   uploaderID,
		UploaderName: uploaderName,
		CreatedAt:    row.CreatedAt,
		UploadedAt:   row.CreatedAt,
	}
}

func uploaderFields(row *model.FileWithUploader) (string, string) {
	id := ""
	if row.UploadedBy != nil {
		id = row.UploadedBy.String()
	}
	return id, row.UploaderName
}
