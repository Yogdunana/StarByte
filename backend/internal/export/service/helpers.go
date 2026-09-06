package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/internal/export/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func canAccessExport(ownerID, callerID string, isSuper bool) bool {
	if isSuper {
		return true
	}
	return ownerID != "" && ownerID == callerID
}

func (s *exportService) loadDoneTask(ctx context.Context, id string) (*dto.ExportTaskResponse, error) {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, response.NewError(response.CodeInternalError, "查询导出任务失败")
	}
	return toTaskDTO(task), nil
}

func (s *exportService) patchTask(ctx context.Context, id, status string, progress int, fileID, filename string) error {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	task.Status = status
	task.Progress = progress
	if fileID != "" {
		task.FileID = fileID
	}
	if filename != "" {
		task.Filename = filename
	}
	return s.repo.SaveTask(ctx, task)
}

func (s *exportService) finishTask(ctx context.Context, id, fileID, filename string) error {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	task.Status = model.StatusDone
	task.Progress = 100
	task.FileID = fileID
	task.Filename = filename
	task.Error = ""
	if err := s.repo.SaveTask(ctx, task); err != nil {
		return err
	}
	if s.notifier != nil && task.UserID != "" {
		s.notifier.NotifyReady(ctx, task.UserID, filename, fileID)
	}
	return nil
}

func (s *exportService) failTask(ctx context.Context, id, msg string) error {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	task.Status = model.StatusFailed
	task.Error = msg
	return s.repo.SaveTask(ctx, task)
}

func newTask(filename, format, userID string) *model.ExportTask {
	return &model.ExportTask{
		ID:        uuid.NewString(),
		UserID:    userID,
		Status:    model.StatusPending,
		Progress:  0,
		Filename:  filename,
		Format:    format,
		CreatedAt: time.Now(),
	}
}

func toTaskDTO(t *model.ExportTask) *dto.ExportTaskResponse {
	out := &dto.ExportTaskResponse{
		TaskID:   t.ID,
		Status:   t.Status,
		Progress: t.Progress,
		FileID:   t.FileID,
		Error:    t.Error,
		Filename: t.Filename,
		Format:   t.Format,
	}
	if !t.CreatedAt.IsZero() {
		out.CreatedAt = t.CreatedAt.Format(time.RFC3339)
	}
	return out
}

func filenameOf(name, format string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "export"
	}
	name = strings.ReplaceAll(name, "/", "_")
	ext := validFormats[format]
	if ext == "" {
		ext = format
	}
	lower := strings.ToLower(name)
	if !strings.HasSuffix(lower, "."+ext) {
		name += "." + ext
	}
	return name
}

func contentTypeOf(format string) string {
	switch format {
	case "excel":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "csv":
		return "text/csv; charset=utf-8"
	case "json":
		return "application/json"
	default:
		return "application/pdf"
	}
}

func copyTableReq(req *dto.TableExportRequest) *dto.TableExportRequest {
	out := *req
	out.Columns = append([]string(nil), req.Columns...)
	out.Rows = make([][]string, len(req.Rows))
	for i, row := range req.Rows {
		out.Rows[i] = append([]string(nil), row...)
	}
	return &out
}

func copyTplReq(req *dto.TemplateExportRequest) *dto.TemplateExportRequest {
	out := *req
	merged := map[string]string{}
	for k, v := range req.Variables {
		merged[k] = v
	}
	for k, v := range req.Vars {
		merged[k] = v
	}
	out.Vars = merged
	out.Variables = nil
	return &out
}

func sanitizeSheet(name string) string {
	repl := strings.NewReplacer(":", "_", "\\", "_", "/", "_", "?", "_", "*", "_", "[", "_", "]", "_")
	name = repl.Replace(strings.TrimSpace(name))
	if name == "" {
		return "Sheet1"
	}
	runes := []rune(name)
	if len(runes) > 31 {
		name = string(runes[:31])
	}
	return name
}
