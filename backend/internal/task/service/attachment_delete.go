package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) DeleteAttachment(ctx context.Context, taskID, attachID, operator uuid.UUID) error {
	jobID := uuid.New()
	err := s.transaction(ctx, taskID, func(b *taskService) error {
		t, err := b.mustTask(ctx, taskID)
		if err != nil {
			return err
		}
		if err = b.ensureMutable(t, operator); err != nil {
			return err
		}
		file, err := b.mustAttachment(ctx, attachID, taskID)
		if err != nil {
			return err
		}
		if file.UploadedBy == nil {
			return response.NewError(response.CodeTaskUploadFail, "附件缺少上传者，需先核实文件归属")
		}
		if s.db != nil {
			n, err := b.cleanup.OtherReferences(ctx, file.FileID, attachID)
			if err != nil {
				return err
			}
			if n > 0 {
				return response.NewError(response.CodeTaskInvalidState, "文件仍被其他任务引用，不能直接删除")
			}
			now := time.Now()
			job := &model.FileCleanup{ID: jobID, AttachmentID: attachID, TaskID: taskID, FileID: file.FileID, UploadedBy: *file.UploadedBy, RequestedBy: operator, Status: "pending", NextAttemptAt: now, CreatedAt: now, UpdatedAt: now}
			if err := b.cleanup.Create(ctx, job); err != nil {
				return err
			}
		}
		if err := b.addLog(ctx, taskID, operator, "attachment_remove_requested", file.FileName, "", "已记录附件及存储文件删除请求"); err != nil {
			return err
		}
		if s.db == nil && b.bridge != nil {
			if err := b.bridge.Delete(ctx, file.FileID, *file.UploadedBy); err != nil {
				return err
			}
		}
		return b.files.Delete(ctx, attachID)
	})
	if err != nil || s.db == nil {
		return err
	}
	if err := s.deleteStoredAttachment(ctx, jobID); err != nil {
		return response.NewError(response.CodeInternalError, "删除请求已记录，存储清理暂未完成，系统将自动重试")
	}
	return nil
}

// The durable job is committed before touching object storage. Retrying after
// either an object-delete or database failure is safe, including after restart.
func (s *taskService) deleteStoredAttachment(ctx context.Context, id uuid.UUID) error {
	var deliveryErr error
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		queue := repo.NewCleanupRepo(tx)
		job, err := queue.Lock(ctx, id)
		if err != nil {
			return err
		}
		if job.Status == "completed" {
			return nil
		}
		job.Attempts++
		job.UpdatedAt = time.Now()
		if s.bridge == nil {
			deliveryErr = fmt.Errorf("file service unavailable")
		} else {
			deliveryErr = s.bridge.Delete(ctx, job.FileID, job.UploadedBy)
		}
		var app *response.AppError
		if errors.As(deliveryErr, &app) && app.Code == response.CodeNotFound {
			deliveryErr = nil
		}
		if deliveryErr != nil {
			delay := time.Duration(job.Attempts) * 30 * time.Second
			if delay > 15*time.Minute {
				delay = 15 * time.Minute
			}
			job.NextAttemptAt = time.Now().Add(delay)
			job.LastError = deliveryErr.Error()
			return queue.Save(ctx, job)
		}
		now := time.Now()
		job.Status = "completed"
		job.CompletedAt = &now
		job.LastError = ""
		if err := repo.NewLogRepo(tx).Create(ctx, &model.TaskLog{ID: uuid.New(), TaskID: job.TaskID, OperatorID: job.RequestedBy, ActionType: "attachment_remove", OldValue: job.FileID.String(), Comment: "附件及存储文件已删除", CreatedAt: now}); err != nil {
			return err
		}
		return queue.Save(ctx, job)
	})
	if err != nil {
		return err
	}
	return deliveryErr
}
func (s *taskService) ProcessAttachmentDeletions(ctx context.Context) (int, error) {
	if s.db == nil {
		return 0, nil
	}
	ids, err := s.cleanup.Pending(ctx, time.Now())
	if err != nil {
		return 0, err
	}
	done := 0
	var failures []error
	for _, id := range ids {
		if err := s.deleteStoredAttachment(ctx, id); err != nil {
			failures = append(failures, err)
		} else {
			done++
		}
	}
	return done, errors.Join(failures...)
}
