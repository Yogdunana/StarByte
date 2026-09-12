package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
)

func (s *taskService) EscalateOverdueWorkflows(ctx context.Context) (int, error) {
	if s.flow == nil {
		return 0, nil
	}
	todos, err := s.flow.ListOverdueCollaborationTodos(ctx, time.Now())
	if err != nil {
		return 0, err
	}
	applied := 0
	for _, todo := range todos {
		taskID, err := uuid.Parse(todo.BusinessKey)
		if err != nil || taskID == uuid.Nil {
			continue
		}
		err = s.transaction(ctx, taskID, func(b *taskService) error {
			return b.escalateTodo(ctx, todo)
		})
		if err != nil {
			logger.Warn("task workflow timeout escalate failed", zap.Error(err), zap.String("task_id", todo.BusinessKey), zap.String("stage", todo.Stage))
			continue
		}
		applied++
	}
	return applied, nil
}

func (s *taskService) escalateTodo(ctx context.Context, todo engine.OverdueCollaborationTodo) error {
	t, err := s.tasks.GetByID(ctx, uuid.MustParse(todo.BusinessKey))
	if err != nil {
		return err
	}
	if t == nil || t.WorkflowInstanceID == nil || *t.WorkflowInstanceID != todo.InstanceID || model.IsClosed(t.Status) {
		return nil
	}
	if s.alreadyEscalated(ctx, t.ID, todo.Stage) {
		return nil
	}
	if s.transfers != nil {
		if pending, err := s.transfers.Pending(ctx, t.ID); err != nil {
			return err
		} else if pending != nil {
			return nil
		}
	}
	switch todo.Stage {
	case "assignment":
		return s.escalateAssignment(ctx, t)
	case "execution":
		return s.escalateHolder(ctx, t, "execution", t.AssigneeID, func(next uuid.UUID) { t.AssigneeID = &next })
	case "review":
		return s.escalateHolder(ctx, t, "review", t.ReviewerID, func(next uuid.UUID) { t.ReviewerID = &next })
	case "acceptance":
		return s.escalateHolder(ctx, t, "acceptance", t.AcceptorID, func(next uuid.UUID) { t.AcceptorID = &next })
	default:
		return nil
	}
}

func (s *taskService) escalateAssignment(ctx context.Context, t *model.Task) error {
	if t.AssigneeID != nil {
		return nil
	}
	next, err := s.pickEscalationTarget(ctx, t, nil)
	if err != nil {
		return err
	}
	if next == uuid.Nil {
		return s.recordEscalate(ctx, t, "assignment", t.CreatorID, "超时仍无可用执行人，已通知发布人")
	}
	t.AssigneeID = &next
	if err := s.tasks.Update(ctx, t); err != nil {
		return err
	}
	if err := s.flow.ClaimTaskAssignment(ctx, *t.WorkflowInstanceID, next, "超时自动分配"); err != nil {
		return err
	}
	if err := s.projectWorkflow(ctx, t); err != nil {
		return err
	}
	return s.recordEscalate(ctx, t, "assignment", next, "超时自动分配执行人")
}

func (s *taskService) escalateHolder(ctx context.Context, t *model.Task, stage string, current *uuid.UUID, apply func(uuid.UUID)) error {
	if current == nil {
		return s.recordEscalate(ctx, t, stage, t.CreatorID, "超时缺少当前处理人，已通知发布人")
	}
	excluded := []uuid.UUID{*current}
	next, err := s.pickEscalationTarget(ctx, t, excluded)
	if err != nil {
		return err
	}
	if next == uuid.Nil && t.CreatorID != *current && canTakeEscalatedStage(t, stage, t.CreatorID) {
		next = t.CreatorID
	}
	if next == uuid.Nil || next == *current || !canTakeEscalatedStage(t, stage, next) {
		return s.recordEscalate(ctx, t, stage, t.CreatorID, "超时无法重新分配，已通知发布人")
	}
	if err := s.flow.ReassignStage(ctx, *t.WorkflowInstanceID, stage, *current, next, t.CreatorID, "超时升级重新分配"); err != nil {
		return err
	}
	apply(next)
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return err
	}
	return s.recordEscalate(ctx, t, stage, next, "超时升级重新分配")
}

func canTakeEscalatedStage(t *model.Task, stage string, id uuid.UUID) bool {
	if id == uuid.Nil {
		return false
	}
	switch stage {
	case "execution":
		return validateWorkflowAssignee(t, &id) == nil
	case "review", "acceptance":
		return t.AssigneeID == nil || *t.AssigneeID != id
	default:
		return true
	}
}

func (s *taskService) pickEscalationTarget(ctx context.Context, t *model.Task, extra []uuid.UUID) (uuid.UUID, error) {
	if s.assignments == nil || t.DepartmentID == nil {
		return uuid.Nil, nil
	}
	policy := model.AssignmentPolicy{Mode: "department", DepartmentID: *t.DepartmentID}
	if t.AssignmentPolicy != "" {
		parsed := model.AssignmentPolicy{}
		if err := json.Unmarshal([]byte(t.AssignmentPolicy), &parsed); err != nil {
			return uuid.Nil, err
		}
		if parsed.Mode != "" {
			policy.Mode = parsed.Mode
			policy.RoleID = parsed.RoleID
		}
	}
	policy.DepartmentID = *t.DepartmentID
	excluded := append([]uuid.UUID{}, extra...)
	for _, id := range []*uuid.UUID{t.ReviewerID, t.AcceptorID, t.AssigneeID} {
		if id != nil {
			excluded = append(excluded, *id)
		}
	}
	user, err := s.assignments.Pick(ctx, policy, excluded)
	if err != nil {
		return uuid.Nil, err
	}
	if user == nil {
		return uuid.Nil, nil
	}
	if err := validateWorkflowAssignee(t, &user.ID); err != nil {
		return uuid.Nil, nil
	}
	return user.ID, nil
}

func (s *taskService) alreadyEscalated(ctx context.Context, taskID uuid.UUID, stage string) bool {
	rows, err := s.logs.ListByTask(ctx, taskID)
	if err != nil {
		return false
	}
	for _, row := range rows {
		if row.ActionType == "workflow_escalate" && row.OldValue == stage {
			return true
		}
	}
	return false
}

func (s *taskService) recordEscalate(ctx context.Context, t *model.Task, stage string, next uuid.UUID, comment string) error {
	if err := s.addLog(ctx, t.ID, t.CreatorID, "workflow_escalate", stage, next.String(), comment); err != nil {
		return err
	}
	s.notifyUsers(ctx, reminderTargets(t), tplTaskOverdue, t, comment)
	if next != uuid.Nil && (t.AssigneeID == nil || *t.AssigneeID != next) {
		s.notifyUsers(ctx, []uuid.UUID{next}, tplTaskAssigned, t, comment)
	}
	return nil
}
