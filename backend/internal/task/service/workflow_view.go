package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func (s *taskService) GetWorkflow(ctx context.Context, id, actor uuid.UUID) (*dto.WorkflowResponse, error) {
	t, err := s.workflowTask(ctx, id, actor)
	if err != nil {
		return nil, err
	}
	person := func(id uuid.UUID) (dto.Person, error) {
		u, err := s.tasks.GetUser(ctx, id)
		return dto.Person{ID: id.String(), Name: displayName(u)}, err
	}
	out := &dto.WorkflowResponse{Revision: t.WorkflowRevision, TaskID: id.String(), Title: t.Title, InstanceID: t.WorkflowInstanceID.String(), Stage: t.WorkflowStage, Submission: t.Submission, UpdatedAt: t.UpdatedAt, History: []dto.LogResponse{}}
	policy := model.AssignmentPolicy{}
	if t.AssignmentPolicy != "" {
		if err := json.Unmarshal([]byte(t.AssignmentPolicy), &policy); err != nil {
			return nil, err
		}
	}
	out.AssignmentMode = policy.Mode
	if out.Creator, err = person(t.CreatorID); err != nil {
		return nil, err
	}
	if t.ReviewerID != nil {
		if out.Reviewer, err = person(*t.ReviewerID); err != nil {
			return nil, err
		}
	}
	if t.AcceptorID != nil {
		if out.Acceptor, err = person(*t.AcceptorID); err != nil {
			return nil, err
		}
	}
	if t.AssigneeID != nil {
		p, err := person(*t.AssigneeID)
		if err != nil {
			return nil, err
		}
		out.Assignee = &p
	}
	executable := t.WorkflowStage == "execution" && t.AssigneeID != nil && *t.AssigneeID == actor
	out.CanStart = executable && t.Status == model.StatusPending
	out.CanPause = executable && t.Status == model.StatusDoing
	out.CanResume = executable && t.Status == model.StatusHeld
	out.CanSubmit = t.WorkflowStage == "execution" && t.Status == model.StatusDoing && t.AssigneeID != nil && *t.AssigneeID == actor
	out.CanApprove = !model.IsClosed(t.Status) && ((t.WorkflowStage == "review" && t.ReviewerID != nil && *t.ReviewerID == actor) || (t.WorkflowStage == "acceptance" && t.AcceptorID != nil && *t.AcceptorID == actor))
	out.CanReturn = out.CanApprove
	out.CanReject = out.CanApprove
	out.CanClaim = t.WorkflowStage == "assignment" && t.AssigneeID == nil && !model.IsClosed(t.Status) && validateWorkflowAssignee(t, &actor) == nil
	out.CanDelegate = executable && !model.IsClosed(t.Status)
	if s.transfers != nil {
		if pending, err := s.transfers.Pending(ctx, id); err != nil {
			return nil, err
		} else if pending != nil {
			out.CanDelegate = false
			handover, err := s.handoverSnapshot(ctx, t, pending, actor)
			if err != nil {
				return nil, err
			}
			out.Handover = handover
		}
	}
	logs, err := s.logs.ListByTask(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, log := range logs {
		if strings.HasPrefix(log.ActionType, "workflow_") {
			out.History = append(out.History, mapLog(log))
		}
	}
	return out, nil
}
