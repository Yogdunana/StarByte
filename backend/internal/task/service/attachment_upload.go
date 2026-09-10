package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) UploadAttachment(ctx context.Context, taskID, operator uuid.UUID, header *multipart.FileHeader) (*dto.AttachmentResponse, error) {
	t, err := s.mustTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMutable(t, operator); err != nil {
		return nil, err
	}
	if s.bridge == nil || header == nil {
		return nil, response.NewError(response.CodeTaskUploadFail, "文件上传失败")
	}
	uploaded, err := s.bridge.Upload(ctx, operator, header, "document", false)
	if err != nil {
		return nil, fmt.Errorf("upload task attachment: %w", err)
	}
	fileID, err := uuid.Parse(uploaded.ID)
	if err != nil {
		return nil, response.NewError(response.CodeTaskUploadFail, "上传服务返回了无效文件 ID")
	}
	out, err := taskMutation(ctx, s, taskID, func(b *taskService) (*dto.AttachmentResponse, error) {
		// Recheck after upload: task ownership or status may have changed meanwhile.
		t, err := b.mustTask(ctx, taskID)
		if err != nil {
			return nil, err
		}
		if err := b.ensureMutable(t, operator); err != nil {
			return nil, err
		}
		row := &model.TaskAttachment{ID: uuid.New(), TaskID: taskID, FileID: fileID, UploadedBy: &operator, CreatedAt: time.Now()}
		if err := b.files.Create(ctx, row); err != nil {
			return nil, fmt.Errorf("create attachment: %w", err)
		}
		if err := b.addLog(ctx, taskID, operator, "attachment_add", "", uploaded.OriginalName, ""); err != nil {
			return nil, err
		}
		named, err := b.files.GetNamed(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		if named == nil {
			return nil, response.NewError(response.CodeTaskAttachGone, "附件不存在")
		}
		result := mapAttachment(*named)
		return &result, nil
	})
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()
		if cleanupErr := s.bridge.Delete(cleanupCtx, fileID, operator); cleanupErr != nil {
			return nil, fmt.Errorf("attach failed (%v); uploaded file cleanup failed: %w", err, cleanupErr)
		}
	}
	return out, err
}
