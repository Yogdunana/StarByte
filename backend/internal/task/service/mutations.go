package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
)

func (s *taskService) Create(ctx context.Context, operator uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	lock := uuid.Nil
	if req != nil && req.ParentID != "" {
		if parent := parseUUIDPtr(req.ParentID); parent != nil {
			lock = *parent
		}
	}
	return taskMutation(ctx, s, lock, func(b *taskService) (*dto.TaskResponse, error) { return b.create(ctx, operator, req) })
}
func (s *taskService) Update(ctx context.Context, id, operator uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	return taskMutation(ctx, s, id, func(b *taskService) (*dto.TaskResponse, error) { return b.update(ctx, id, operator, req) })
}
func (s *taskService) Delete(ctx context.Context, id, operator uuid.UUID) error {
	return s.transaction(ctx, id, func(b *taskService) error { return b.delete(ctx, id, operator) })
}
func (s *taskService) Assign(ctx context.Context, id, operator uuid.UUID, assigneeRaw string) (*dto.TaskResponse, error) {
	return taskMutation(ctx, s, id, func(b *taskService) (*dto.TaskResponse, error) { return b.assign(ctx, id, operator, assigneeRaw) })
}
func (s *taskService) Transfer(ctx context.Context, id, operator uuid.UUID, req *dto.TransferRequest) (*dto.TaskResponse, error) {
	return taskMutation(ctx, s, id, func(b *taskService) (*dto.TaskResponse, error) { return b.transfer(ctx, id, operator, req) })
}
func (s *taskService) ChangeStatus(ctx context.Context, id, operator uuid.UUID, req *dto.StatusRequest) (*dto.TaskResponse, error) {
	return taskMutation(ctx, s, id, func(b *taskService) (*dto.TaskResponse, error) { return b.changeStatus(ctx, id, operator, req) })
}
func (s *taskService) Urge(ctx context.Context, id, operator uuid.UUID, message string) error {
	return s.transaction(ctx, id, func(b *taskService) error { return b.urge(ctx, id, operator, message) })
}
func (s *taskService) AddComment(ctx context.Context, taskID, operator uuid.UUID, req *dto.CommentRequest) (*dto.CommentResponse, error) {
	return taskMutation(ctx, s, taskID, func(b *taskService) (*dto.CommentResponse, error) { return b.addComment(ctx, taskID, operator, req) })
}
func (s *taskService) UpdateComment(ctx context.Context, taskID, commentID, operator uuid.UUID, content string) (*dto.CommentResponse, error) {
	return taskMutation(ctx, s, taskID, func(b *taskService) (*dto.CommentResponse, error) {
		return b.updateComment(ctx, taskID, commentID, operator, content)
	})
}
func (s *taskService) DeleteComment(ctx context.Context, taskID, commentID, operator uuid.UUID) error {
	return s.transaction(ctx, taskID, func(b *taskService) error { return b.deleteComment(ctx, taskID, commentID, operator) })
}
