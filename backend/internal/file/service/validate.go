package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/file/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

const (
	maxImageSize    int64 = 10 << 20
	maxDocumentSize int64 = 50 << 20
	maxVideoSize    int64 = 500 << 20
	maxBatchFiles         = 10
	mb                    = 1 << 20
)

var extCategory = map[string]string{
	".jpg": model.CategoryImage, ".jpeg": model.CategoryImage,
	".png": model.CategoryImage, ".gif": model.CategoryImage, ".webp": model.CategoryImage,
	".pdf": model.CategoryDocument, ".doc": model.CategoryDocument, ".docx": model.CategoryDocument,
	".xls": model.CategoryDocument, ".xlsx": model.CategoryDocument,
	".ppt": model.CategoryDocument, ".pptx": model.CategoryDocument,
	".mp4": model.CategoryVideo,
}

var mimeByExt = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	".gif": "image/gif", ".webp": "image/webp",
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".mp4":  "video/mp4",
}

var categoryLimit = map[string]int64{
	model.CategoryImage:    maxImageSize,
	model.CategoryDocument: maxDocumentSize,
	model.CategoryVideo:    maxVideoSize,
}

type validatedFile struct {
	Category string
	Ext      string
	MimeType string
}

func validateUpload(filename string, size int64, declaredCategory string) (*validatedFile, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = filepath.Ext(filename)
	}
	category, ok := extCategory[ext]
	if !ok {
		if ext == "" {
			ext = "(none)"
		}
		return nil, response.NewError(response.CodeBadRequest, "文件类型不允许: "+ext)
	}
	if declaredCategory != "" && declaredCategory != category {
		return nil, response.NewError(response.CodeBadRequest, "文件类型不允许: "+ext)
	}
	limit := categoryLimit[category]
	if size > limit {
		return nil, response.NewError(response.CodeBadRequest,
			fmt.Sprintf("文件大小超过限制: %dMB > %dMB", size/mb, limit/mb))
	}
	return &validatedFile{
		Category: category,
		Ext:      ext,
		MimeType: mimeByExt[ext],
	}, nil
}
