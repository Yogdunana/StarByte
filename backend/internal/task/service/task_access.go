package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) mustTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if t == nil {
		return nil, response.NewError(response.CodeTaskNotFound, "任务不存在")
	}
	if viewer, ok := model.ViewerFromContext(ctx); ok && !canViewTask(t, viewer.ID, viewer.Scope) {
		return nil, response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	return t, nil
}

func (s *taskService) mustVisible(ctx context.Context, id, viewer uuid.UUID, scope *rbacModel.DataScopeCondition) (*model.Task, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canViewTask(t, viewer, scope) {
		return nil, response.NewError(response.CodeTaskNoAccess, "无权查看该任务")
	}
	return t, nil
}

func (s *taskService) ensureMutable(t *model.Task, operator uuid.UUID) error {
	if t.WorkflowInstanceID != nil && t.WorkflowStage != "assignment" && t.WorkflowStage != "execution" {
		return response.NewError(response.CodeConflict, "交付正在审核或验收，需退回执行后才能修改任务及附件")
	}
	if model.IsClosed(t.Status) {
		return response.NewError(response.CodeTaskClosed, "任务已关闭，无法操作")
	}
	if t.CreatorID != operator && (t.AssigneeID == nil || *t.AssigneeID != operator) {
		return response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	return nil
}
