package service

import (
	"fmt"
	"net/http"
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

func validateUpload(filename string, size int64, declaredCategory string, data []byte) (*validatedFile, error) {
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
	// 内容嗅探交叉校验：防止「改扩展名藏可执行文件」的绕过。
	if err := sniffCheck(ext, category, data); err != nil {
		return nil, err
	}
	return &validatedFile{
		Category: category,
		Ext:      ext,
		MimeType: mimeByExt[ext],
	}, nil
}

// sniffCheck 用文件头 magic bytes 与扩展名声明的类型做交叉校验。
//
// 策略（刻意非完整校验，仅拦常见绕过）：
//   - 图片(image/*)、application/pdf：用 http.DetectContentType 做强校验；
//   - video/mp4：Go 的 DetectContentType 无法识别 mp4（会返回 application/octet-stream），
//     故额外用 mp4 文件头(ftyp box)做专用嗅探，避免误杀正常 mp4 又把伪装档放过去；
//   - 其余（office/zip/txt 等）：Go 嗅探为 text/plain 或 octet-stream，判断很宽，
//     不强行比对，只交给 isBannedSniff 拦截被明确识别为可执行/脚本的嗅探结果。
func sniffCheck(ext, category string, data []byte) error {
	if len(data) == 0 {
		return nil // 无内容可嗅探时跳过
	}
	sniffed := http.DetectContentType(data)

	switch category {
	case model.CategoryImage:
		if !strings.HasPrefix(sniffed, "image/") {
			return response.NewError(response.CodeBadRequest, "文件内容与扩展名不符")
		}
	case model.CategoryDocument:
		if ext == ".pdf" {
			if !strings.HasPrefix(sniffed, "application/pdf") {
				return response.NewError(response.CodeBadRequest, "文件内容与扩展名不符")
			}
		}
	case model.CategoryVideo:
		// 真实 mp4 经 DetectContentType 得到 application/octet-stream，
		// 故允许 video/* 或带 ftyp 文件头两种情况。
		if !strings.HasPrefix(sniffed, "video/") && !isMP4(data) {
			return response.NewError(response.CodeBadRequest, "文件内容与扩展名不符")
		}
	}

	if isBannedSniff(sniffed) {
		return response.NewError(response.CodeBadRequest, "文件内容与扩展名不符")
	}
	return nil
}

// isMP4 判断数据是否以标准 mp4 文件头(第一个 box 类型为 ftyp)开头。
func isMP4(data []byte) bool {
	return len(data) >= 12 && string(data[4:8]) == "ftyp"
}

// isBannedSniff 判断嗅探出的 MIME 是否属于被明确禁止的可执行/脚本类型。
func isBannedSniff(sniffed string) bool {
	banned := []string{
		"application/x-msdownload", "application/x-executable", "application/x-dosexec",
		"application/x-sh", "application/x-shellscript", "text/x-shellscript",
		"application/x-python", "text/x-python", "application/javascript", "text/javascript",
	}
	for _, b := range banned {
		if strings.HasPrefix(sniffed, b) {
			return true
		}
	}
	return false
}
