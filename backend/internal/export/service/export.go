package service

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/internal/export/model"
	"github.com/Yogdunana/StarByte/backend/internal/export/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/google/uuid"
)

// asyncRowThreshold is the row count that triggers background processing.
// Tests may lower this (e.g. to 2) to exercise the goroutine path.
var asyncRowThreshold = 10000

// maxExportRows is a hard cap even for async jobs (error 17005).
var maxExportRows = 1000000

// ReadyNotifier is optional; used when an export file is ready.
type ReadyNotifier interface {
	NotifyReady(ctx context.Context, userID, filename, fileID string)
}

var validFormats = map[string]string{
	"excel": "xlsx",
	"csv":   "csv",
	"pdf":   "pdf",
	"json":  "json",
}

type ExportService interface {
	ExportTable(ctx context.Context, format, userID string, req *dto.TableExportRequest) (*dto.ExportTaskResponse, error)
	ExportTemplate(ctx context.Context, templateID, userID string, req *dto.TemplateExportRequest) (*dto.ExportTaskResponse, error)
	GetTask(ctx context.Context, taskID, callerID string, isSuper bool) (*dto.ExportTaskResponse, error)
	Download(ctx context.Context, fileID, callerID string, isSuper, loadBytes bool) (*dto.DownloadResult, error)
	ListTemplates() []dto.TemplateInfo
}

type exportService struct {
	repo     repo.ExportRepo
	store    storage.ObjectStorage
	notifier ReadyNotifier
}

func NewExportService(r repo.ExportRepo, store storage.ObjectStorage, notifier ReadyNotifier) ExportService {
	return &exportService{repo: r, store: store, notifier: notifier}
}

func (s *exportService) ListTemplates() []dto.TemplateInfo {
	return listBuiltinTemplates()
}

func (s *exportService) ExportTable(ctx context.Context, format, userID string, req *dto.TableExportRequest) (*dto.ExportTaskResponse, error) {
	if req == nil {
		return nil, response.NewError(response.CodeExportEmptyData, "导出数据为空")
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if _, ok := validFormats[format]; !ok {
		return nil, response.NewError(response.CodeExportInvalidFormat, "不支持的导出格式")
	}
	if err := validateTable(req); err != nil {
		return nil, err
	}
	copied := copyTableReq(req)
	task := newTask(filenameOf(copied.Filename, format), format, userID)
	if err := s.repo.SaveTask(ctx, task); err != nil {
		return nil, response.NewError(response.CodeInternalError, "创建导出任务失败")
	}
	if len(copied.Rows) > asyncRowThreshold {
		go func() {
			_ = s.runTableJob(context.Background(), task.ID, format, userID, copied)
		}()
		return toTaskDTO(task), nil
	}
	if err := s.runTableJob(ctx, task.ID, format, userID, copied); err != nil {
		return nil, err
	}
	return s.loadDoneTask(ctx, task.ID)
}

func (s *exportService) ExportTemplate(ctx context.Context, templateID, userID string, req *dto.TemplateExportRequest) (*dto.ExportTaskResponse, error) {
	if findBuiltin(templateID) == nil {
		return nil, response.NewError(response.CodeExportTplNotFound, "导出模板不存在")
	}
	if req == nil {
		req = &dto.TemplateExportRequest{}
	}
	name := req.Filename
	if strings.TrimSpace(name) == "" {
		name = templateID
	}
	task := newTask(filenameOf(name, "pdf"), "pdf", userID)
	if err := s.repo.SaveTask(ctx, task); err != nil {
		return nil, response.NewError(response.CodeInternalError, "创建导出任务失败")
	}
	copied := copyTplReq(req)
	if err := s.runTemplateJob(ctx, task.ID, templateID, userID, copied); err != nil {
		return nil, err
	}
	return s.loadDoneTask(ctx, task.ID)
}

func (s *exportService) GetTask(ctx context.Context, taskID, callerID string, isSuper bool) (*dto.ExportTaskResponse, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, response.NewError(response.CodeExportTaskNotFound, "导出任务不存在")
		}
		return nil, response.NewError(response.CodeInternalError, "查询导出任务失败")
	}
	if !canAccessExport(task.UserID, callerID, isSuper) {
		return nil, response.NewForbiddenError("无权查看该导出任务")
	}
	return toTaskDTO(task), nil
}

func (s *exportService) Download(ctx context.Context, fileID, callerID string, isSuper, loadBytes bool) (*dto.DownloadResult, error) {
	meta, err := s.repo.GetFile(ctx, fileID)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, response.NewError(response.CodeExportFileExpired, "导出文件已过期")
		}
		return nil, response.NewError(response.CodeInternalError, "读取导出文件失败")
	}
	if !meta.ExpiresAt.IsZero() && time.Now().After(meta.ExpiresAt) {
		return nil, response.NewError(response.CodeExportFileExpired, "导出文件已过期")
	}
	if !canAccessExport(meta.UserID, callerID, isSuper) {
		return nil, response.NewForbiddenError("无权下载该导出文件")
	}
	out := &dto.DownloadResult{
		FileID:      meta.FileID,
		Filename:    meta.Filename,
		ContentType: meta.ContentType,
		ExpiresAt:   meta.ExpiresAt,
	}
	if !loadBytes {
		return out, nil
	}
	data, err := s.loadExportBytes(ctx, meta)
	if err != nil {
		return nil, err
	}
	out.Bytes = data
	return out, nil
}

func (s *exportService) loadExportBytes(ctx context.Context, meta *model.FileMeta) ([]byte, error) {
	var storeErr error
	if s.store != nil && meta.ObjectKey != "" {
		rc, _, derr := s.store.Download(ctx, meta.ObjectKey)
		if derr != nil {
			storeErr = derr
		} else {
			defer rc.Close()
			data, rerr := io.ReadAll(rc)
			if rerr != nil {
				storeErr = rerr
			} else if len(data) > 0 {
				return data, nil
			}
		}
	}
	blob, berr := s.repo.GetBlob(ctx, meta.FileID)
	if berr == nil && len(blob) > 0 {
		return blob, nil
	}
	if storeErr != nil {
		return nil, response.NewError(response.CodeInternalError, "读取导出文件失败")
	}
	return nil, response.NewError(response.CodeExportFileExpired, "导出文件已过期")
}

func validateTable(req *dto.TableExportRequest) error {
	if len(req.Columns) == 0 || len(req.Rows) == 0 {
		return response.NewError(response.CodeExportEmptyData, "导出数据为空")
	}
	if len(req.Rows) > maxExportRows {
		return response.NewError(response.CodeExportTooManyRows, "数据量过大且无法异步导出")
	}
	return nil
}

func (s *exportService) runTableJob(ctx context.Context, taskID, format, userID string, req *dto.TableExportRequest) error {
	_ = s.patchTask(ctx, taskID, model.StatusRunning, 10, "", "")
	data, err := renderTable(format, req)
	if err != nil {
		_ = s.failTask(ctx, taskID, err.Error())
		return err
	}
	_ = s.patchTask(ctx, taskID, model.StatusRunning, 70, "", "")
	fileID, filename, err := s.persistFile(ctx, format, filenameOf(req.Filename, format), data, userID)
	if err != nil {
		_ = s.failTask(ctx, taskID, err.Error())
		return err
	}
	return s.finishTask(ctx, taskID, fileID, filename)
}

func (s *exportService) runTemplateJob(ctx context.Context, taskID, templateID, userID string, req *dto.TemplateExportRequest) error {
	_ = s.patchTask(ctx, taskID, model.StatusRunning, 10, "", "")
	htmlBody, err := renderTemplate(templateID, req.Vars)
	if err != nil {
		_ = s.failTask(ctx, taskID, err.Error())
		return err
	}
	title := ""
	if req.Vars != nil {
		title = req.Vars["Title"]
	}
	if title == "" {
		if meta := findBuiltin(templateID); meta != nil {
			title = meta.Name
		}
	}
	data, err := buildTemplatePDF(htmlBody, title, req.Watermark)
	if err != nil {
		_ = s.failTask(ctx, taskID, err.Error())
		return response.NewError(response.CodeInternalError, "生成 PDF 失败")
	}
	fileID, filename, err := s.persistFile(ctx, "pdf", filenameOf(req.Filename, "pdf"), data, userID)
	if err != nil {
		_ = s.failTask(ctx, taskID, err.Error())
		return err
	}
	return s.finishTask(ctx, taskID, fileID, filename)
}

func renderTable(format string, req *dto.TableExportRequest) ([]byte, error) {
	switch format {
	case "csv":
		return buildCSV(req)
	case "excel":
		return buildExcel(req)
	case "json":
		return buildJSON(req)
	case "pdf":
		return buildTablePDF(req)
	default:
		return nil, response.NewError(response.CodeExportInvalidFormat, "不支持的导出格式")
	}
}

// RenderTable 同步渲染表格（供审计合规报告等模块复用 #71 引擎，不落任务队列）。
func RenderTable(format string, req *dto.TableExportRequest) ([]byte, string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if _, ok := validFormats[format]; !ok {
		return nil, "", response.NewError(response.CodeExportInvalidFormat, "不支持的导出格式")
	}
	if req == nil || len(req.Columns) == 0 || len(req.Rows) == 0 {
		return nil, "", response.NewError(response.CodeExportEmptyData, "导出数据为空")
	}
	data, err := renderTable(format, req)
	if err != nil {
		return nil, "", err
	}
	return data, filenameOf(req.Filename, format), nil
}

func (s *exportService) persistFile(ctx context.Context, format, filename string, data []byte, userID string) (string, string, error) {
	fileID := uuid.NewString()
	ext := validFormats[format]
	meta := &model.FileMeta{
		FileID:      fileID,
		UserID:      userID,
		Filename:    filename,
		ContentType: contentTypeOf(format),
		Ext:         ext,
		ExpiresAt:   time.Now().Add(model.FileTTL),
	}
	if s.store != nil {
		key := "export/" + fileID + "." + ext
		if err := s.store.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), meta.ContentType); err == nil {
			meta.ObjectKey = key
		}
	}
	if meta.ObjectKey == "" {
		if err := s.repo.SaveBlob(ctx, fileID, data); err != nil {
			return "", "", response.NewError(response.CodeInternalError, "保存导出文件失败")
		}
	}
	if err := s.repo.SaveFile(ctx, meta); err != nil {
		return "", "", response.NewError(response.CodeInternalError, "保存导出文件失败")
	}
	return fileID, filename, nil
}
