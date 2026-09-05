package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *taskService) UploadAttachment(ctx context.Context, taskID, operator uuid.UUID, header *multipart.FileHeader) (*dto.AttachmentResponse, error) {
	t, err := s.mustTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMutable(t, operator); err != nil {
		return nil, err
	}
	if s.bridge == nil {
		return nil, response.NewError(response.CodeTaskUploadFail, "文件上传失败")
	}
	uploaded, err := s.bridge.Upload(ctx, operator, header, "document", false)
	if err != nil {
		return nil, response.NewError(response.CodeTaskUploadFail, "文件上传失败")
	}
	fileID, err := uuid.Parse(uploaded.ID)
	if err != nil {
		return nil, response.NewError(response.CodeTaskUploadFail, "文件上传失败")
	}
	row := &model.TaskAttachment{
		ID:         uuid.New(),
		TaskID:     taskID,
		FileID:     fileID,
		UploadedBy: &operator,
		CreatedAt:  time.Now(),
	}
	if err := s.files.Create(ctx, row); err != nil {
		return nil, fmt.Errorf("create attachment: %w", err)
	}
	named, err := s.files.GetNamed(ctx, row.ID)
	if err != nil || named == nil {
		return &dto.AttachmentResponse{
			ID: row.ID.String(), TaskID: taskID.String(), FileID: fileID.String(),
			FileName: uploaded.OriginalName, FileSize: uploaded.FileSize, FileType: uploaded.MimeType,
			UploadedBy: operator.String(), CreatedAt: row.CreatedAt,
		}, nil
	}
	out := mapAttachment(*named)
	return &out, nil
}

func (s *taskService) ListAttachments(ctx context.Context, taskID uuid.UUID) ([]dto.AttachmentResponse, error) {
	if _, err := s.mustTask(ctx, taskID); err != nil {
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

func (s *taskService) DownloadAttachment(ctx context.Context, taskID, attachID uuid.UUID) (io.ReadCloser, string, string, error) {
	if _, err := s.mustTask(ctx, taskID); err != nil {
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

func (s *taskService) DeleteAttachment(ctx context.Context, taskID, attachID, operator uuid.UUID) error {
	t, err := s.mustTask(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.ensureMutable(t, operator); err != nil {
		return err
	}
	named, err := s.mustAttachment(ctx, attachID, taskID)
	if err != nil {
		return err
	}
	if err := s.files.Delete(ctx, attachID); err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}
	if s.bridge != nil {
		_ = s.bridge.Delete(ctx, named.FileID, operator)
	}
	return nil
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
