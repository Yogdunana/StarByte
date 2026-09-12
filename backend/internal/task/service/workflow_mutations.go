package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) workflowAssignment(ctx context.Context, t *model.Task, actor, target uuid.UUID) error {
	if t.WorkflowInstanceID == nil {
		return nil
	}
	if t.WorkflowStage != "assignment" || actor != t.CreatorID {
		return response.NewError(response.CodeForbidden, "流程任务须由发布人分配；执行中请使用转办")
	}
	if err := validateWorkflowAssignee(t, &target); err != nil {
		return err
	}
	t.AssigneeID = &target
	if err := s.tasks.Update(ctx, t); err != nil {
		return err
	}
	if err := s.flow.TaskCheckpoint(ctx, *t.WorkflowInstanceID, "assignment", actor, "approve", "指定执行人"); err != nil {
		return err
	}
	return s.projectWorkflow(ctx, t)
}
func (s *taskService) workflowTransfer(ctx context.Context, t *model.Task, actor, target uuid.UUID, reason string) (applied bool, err error) {
	if t.WorkflowInstanceID == nil {
		return true, nil
	}
	_, err = s.requestHandover(ctx, t, actor, &dto.HandoverRequest{TargetID: target.String(), Reason: reason, Revision: t.WorkflowRevision})
	return false, err
}
func (s *taskService) workflowStatus(ctx context.Context, t *model.Task, actor uuid.UUID, req *dto.StatusRequest) error {
	if t.WorkflowInstanceID == nil {
		return nil
	}
	if req.Status == model.StatusDone {
		return response.NewError(response.CodeConflict, "请提交交付并完成审核、验收后结束任务")
	}
	if req.Status == model.StatusCancelled {
		if actor != t.CreatorID || strings.TrimSpace(req.Comment) == "" {
			return response.NewError(response.CodeForbidden, "取消流程须由发布人填写原因")
		}
		if err := s.flow.Terminate(ctx, *t.WorkflowInstanceID, actor, req.Comment); err != nil {
			return err
		}
		t.WorkflowStage = "cancelled"
		return nil
	}
	if t.WorkflowStage != "execution" || t.AssigneeID == nil || *t.AssigneeID != actor {
		return response.NewError(response.CodeForbidden, "仅执行人可在执行环节开始或暂停任务")
	}
	return nil
}
