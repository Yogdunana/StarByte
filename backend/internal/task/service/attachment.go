package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) ListAttachments(ctx context.Context, viewer, taskID uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.AttachmentResponse, error) {
	if _, err := s.mustVisible(ctx, taskID, viewer, scope); err != nil {
		return nil, err
	}
	rows, err := s.files.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	out := make([]dto.AttachmentResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapAttachment(row))
	}
	return out, nil
}

func (s *taskService) DownloadAttachment(ctx context.Context, viewer, taskID, attachID uuid.UUID, scope *rbacModel.DataScopeCondition) (io.ReadCloser, string, string, error) {
	if _, err := s.mustVisible(ctx, taskID, viewer, scope); err != nil {
		return nil, "", "", err
	}
	named, err := s.mustAttachment(ctx, attachID, taskID)
	if err != nil {
		return nil, "", "", err
	}
	if s.store == nil || named.FilePath == "" {
		return nil, "", "", response.NewError(response.CodeTaskAttachGone, "附件不存在")
	}
	rc, contentType, err := s.store.Download(ctx, named.FilePath)
	if err != nil {
		return nil, "", "", response.NewError(response.CodeTaskAttachGone, "附件不存在")
	}
	if contentType == "" {
		contentType = named.FileType
	}
	name := named.FileName
	if name == "" {
		name = "attachment"
	}
	return rc, name, contentType, nil
}

func (s *taskService) mustAttachment(ctx context.Context, id, taskID uuid.UUID) (*model.TaskAttachmentNamed, error) {
	row, err := s.files.GetNamed(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get attachment: %w", err)
	}
	if row == nil || row.TaskID != taskID {
		return nil, response.NewError(response.CodeTaskAttachGone, "附件不存在")
	}
	return row, nil
}
