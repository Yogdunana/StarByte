package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) prepareWorkflow(ctx context.Context, t *model.Task, config *dto.WorkflowConfig) error {
	if s.flow == nil || s.afterCommit == nil {
		return response.NewError(response.CodeConflict, "任务审批服务未启用")
	}
	reviewer, err := s.mustUser(ctx, config.ReviewerID)
	if err != nil {
		return err
	}
	acceptor, err := s.mustUser(ctx, config.AcceptorID)
	if err != nil {
		return err
	}
	t.ReviewerID, t.AcceptorID = &reviewer.ID, &acceptor.ID
	if err := s.autoAssign(ctx, t, config.Assignment); err != nil {
		return err
	}
	if err := validateWorkflowAssignee(t, t.AssigneeID); err != nil {
		return err
	}
	t.WorkflowStage = "assignment"
	return nil
}
func validateWorkflowAssignee(t *model.Task, assignee *uuid.UUID) error {
	if assignee != nil && ((t.ReviewerID != nil && *assignee == *t.ReviewerID) || (t.AcceptorID != nil && *assignee == *t.AcceptorID)) {
		return response.NewError(response.CodeBadRequest, "执行人不能审核或验收自己的交付")
	}
	return nil
}
func (s *taskService) startWorkflow(ctx context.Context, t *model.Task) error {
	inst, err := s.flow.Start(ctx, engine.TaskDefinitionKey, t.ID.String(), engine.TaskBusinessType, t.CreatorID, map[string]interface{}{"task_id": t.ID.String()})
	if err != nil {
		return err
	}
	t.WorkflowInstanceID = &inst.ID
	if err := s.tasks.Update(ctx, t); err != nil {
		return err
	}
	if t.AssigneeID != nil {
		if err := s.flow.TaskCheckpoint(ctx, inst.ID, "assignment", t.CreatorID, "approve", "发布时指定执行人"); err != nil {
			return err
		}
	}
	return s.projectWorkflow(ctx, t)
}
func (s *taskService) projectWorkflow(ctx context.Context, t *model.Task) error {
	if s.flow == nil || t.WorkflowInstanceID == nil {
		return response.NewError(response.CodeConflict, "任务缺少关联流程")
	}
	stage, done, err := s.flow.BusinessStage(ctx, *t.WorkflowInstanceID)
	if err != nil {
		return err
	}
	if done {
		stage = "completed"
		t.Status = model.StatusDone
		t.Progress = 100
		now := time.Now()
		t.CompletedAt = &now
	}
	t.WorkflowRevision++
	t.WorkflowStage = stage
	t.UpdatedAt = time.Now()
	return s.tasks.Update(ctx, t)
}
func workflowParticipant(t *model.Task, id uuid.UUID) bool {
	return id != uuid.Nil && (t.CreatorID == id || (t.AssigneeID != nil && *t.AssigneeID == id) || (t.ReviewerID != nil && *t.ReviewerID == id) || (t.AcceptorID != nil && *t.AcceptorID == id))
}
func (s *taskService) workflowTask(ctx context.Context, id, actor uuid.UUID) (*model.Task, error) {
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, response.NewError(response.CodeTaskNotFound, "任务不存在")
	}
	if t.WorkflowInstanceID == nil {
		return nil, response.NewError(response.CodeConflict, "该任务未启用审核验收流程")
	}
	// This panel is for explicitly designated participants; broad task management
	// permission alone must never grant a signature on their behalf.
	if !workflowParticipant(t, actor) {
		return nil, response.NewError(response.CodeForbidden, "仅任务参与人可查看或处理审批")
	}
	return t, nil
}
func (s *taskService) ActWorkflow(ctx context.Context, id, actor uuid.UUID, req *dto.WorkflowActionRequest) (*dto.WorkflowResponse, error) {
	if req == nil || strings.TrimSpace(req.Comment) == "" || utf8.RuneCountInString(req.Comment) > 5000 {
		return nil, response.NewError(response.CodeBadRequest, "请填写 1 至 5000 字交付说明或审批意见")
	}
	err := s.transaction(ctx, id, func(b *taskService) error {
		t, err := b.workflowTask(ctx, id, actor)
		if err != nil {
			return err
		}
		if b.flow == nil || model.IsClosed(t.Status) {
			return response.NewError(response.CodeConflict, "任务流程当前不可处理")
		}
		if req.Revision != t.WorkflowRevision {
			return response.NewError(response.CodeConflict, "任务审批已更新，请刷新后再处理")
		}
		activeUser, err := b.tasks.GetUser(ctx, actor)
		if err != nil {
			return err
		}
		if activeUser == nil {
			return response.NewError(response.CodeForbidden, "当前处理人不存在或已停用")
		}
		stage := t.WorkflowStage
		action := req.Action
		switch action {
		case "start", "pause", "resume":
			if stage != "execution" || t.AssigneeID == nil || *t.AssigneeID != actor {
				return response.NewError(response.CodeForbidden, "仅本环节的执行人可更新执行进度")
			}
			target := map[string]int16{"start": model.StatusDoing, "pause": model.StatusHeld, "resume": model.StatusDoing}[action]
			source := map[string]int16{"start": model.StatusPending, "pause": model.StatusDoing, "resume": model.StatusHeld}[action]
			if t.Status != source {
				return response.NewError(response.CodeConflict, "任务执行状态已变化，请刷新")
			}
			t.Status = target
			action = ""
		case "submit":
			if stage != "execution" || t.Status != model.StatusDoing || t.AssigneeID == nil || *t.AssigneeID != actor {
				return response.NewError(response.CodeForbidden, "仅进行中任务的执行人可以提交交付")
			}
			children, err := b.tasks.ListChildren(ctx, id)
			if err != nil {
				return err
			}
			for _, child := range children {
				if !model.IsClosed(child.Status) {
					return response.NewError(response.CodeConflict, "请先完成或取消全部子任务")
				}
			}
			t.Submission = strings.TrimSpace(req.Comment)
			t.Progress = 80
			action = "approve"
			if err := b.tasks.Update(ctx, t); err != nil {
				return err
			}
		case "approve", "return":
			if stage != "review" && stage != "acceptance" {
				return response.NewError(response.CodeConflict, "当前不是审核或验收环节")
			}
		default:
			return response.NewError(response.CodeBadRequest, "无效的流程操作")
		}
		if action != "" {
			if err := b.flow.TaskCheckpoint(ctx, *t.WorkflowInstanceID, stage, actor, action, req.Comment); err != nil {
				return err
			}
		}
		if req.Action == "return" {
			t.Progress = 50
		}
		if stage == "review" && req.Action == "approve" {
			t.Progress = 90
		}
		if err := b.projectWorkflow(ctx, t); err != nil {
			return err
		}
		if err := b.addLog(ctx, id, actor, "workflow_"+req.Action, stage, t.WorkflowStage, req.Comment); err != nil {
			return err
		}
		b.notifyWorkflow(ctx, t, req.Comment)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetWorkflow(ctx, id, actor)
}
func (s *taskService) notifyWorkflow(ctx context.Context, t *model.Task, comment string) {
	ids := []uuid.UUID{t.CreatorID}
	switch t.WorkflowStage {
	case "execution":
		if t.AssigneeID != nil {
			ids = append(ids, *t.AssigneeID)
		}
	case "review":
		ids = append(ids, *t.ReviewerID)
	case "acceptance":
		ids = append(ids, *t.AcceptorID)
	}
	seen := map[uuid.UUID]bool{}
	unique := []uuid.UUID{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	s.notifyUsers(ctx, unique, tplTaskAssigned, t, comment)
}
