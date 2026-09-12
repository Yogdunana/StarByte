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

func (s *taskService) RequestHandover(ctx context.Context, id, actor uuid.UUID, req *dto.HandoverRequest) (*dto.HandoverResponse, error) {
	return taskMutation(ctx, s, id, func(b *taskService) (*dto.HandoverResponse, error) {
		t, err := b.workflowTask(ctx, id, actor)
		if err != nil {
			return nil, err
		}
		return b.requestHandover(ctx, t, actor, req)
	})
}

func (s *taskService) GetHandover(ctx context.Context, id, actor uuid.UUID) (*dto.HandoverResponse, error) {
	t, request, err := s.handoverVisible(ctx, id, actor)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, response.NewError(response.CodeConflict, "当前没有进行中的转办")
	}
	return s.handoverSnapshot(ctx, t, request, actor)
}

func (s *taskService) GetTransfer(ctx context.Context, transferID, actor uuid.UUID) (*dto.HandoverResponse, error) {
	if s.transfers == nil {
		return nil, response.NewError(response.CodeConflict, "转办服务未启用")
	}
	request, err := s.transfers.Get(ctx, transferID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, response.NewError(response.CodeTaskNotFound, "转办请求不存在")
	}
	t, _, err := s.handoverVisible(ctx, request.TaskID, actor)
	if err != nil {
		return nil, err
	}
	return s.handoverSnapshot(ctx, t, request, actor)
}

func (s *taskService) DecideHandover(ctx context.Context, id, actor uuid.UUID, req *dto.HandoverDecision) (*dto.HandoverResponse, error) {
	return taskMutation(ctx, s, id, func(b *taskService) (*dto.HandoverResponse, error) {
		t, request, err := b.handoverVisible(ctx, id, actor)
		if err != nil {
			return nil, err
		}
		return b.decideHandover(ctx, t, request, actor, req)
	})
}

func (s *taskService) DecideTransfer(ctx context.Context, transferID, actor uuid.UUID, req *dto.HandoverDecision) (*dto.HandoverResponse, error) {
	if s.transfers == nil {
		return nil, response.NewError(response.CodeConflict, "转办服务未启用")
	}
	request, err := s.transfers.Get(ctx, transferID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, response.NewError(response.CodeTaskNotFound, "转办请求不存在")
	}
	return taskMutation(ctx, s, request.TaskID, func(b *taskService) (*dto.HandoverResponse, error) {
		t, _, err := b.handoverVisible(ctx, request.TaskID, actor)
		if err != nil {
			return nil, err
		}
		current, err := b.transfers.Lock(ctx, transferID)
		if err != nil {
			return nil, err
		}
		if current == nil || current.TaskID != request.TaskID {
			return nil, response.NewError(response.CodeTaskNotFound, "转办请求不存在")
		}
		return b.decideHandover(ctx, t, current, actor, req)
	})
}

func (s *taskService) requestHandover(ctx context.Context, t *model.Task, actor uuid.UUID, req *dto.HandoverRequest) (*dto.HandoverResponse, error) {
	if req == nil || strings.TrimSpace(req.Reason) == "" || utf8.RuneCountInString(req.Reason) > 2000 {
		return nil, response.NewError(response.CodeBadRequest, "请填写 1 至 2000 字转办或委托原因")
	}
	if t.WorkflowInstanceID == nil || model.IsClosed(t.Status) {
		return nil, response.NewError(response.CodeConflict, "任务流程当前不可转办")
	}
	if t.WorkflowStage != "execution" || t.AssigneeID == nil || *t.AssigneeID != actor {
		return nil, response.NewError(response.CodeForbidden, "仅执行环节的当前执行人可委托或转办")
	}
	if req.Revision != 0 && req.Revision != t.WorkflowRevision {
		return nil, response.NewError(response.CodeConflict, "任务审批已更新，请刷新后再处理")
	}
	if s.transfers == nil || s.flow == nil {
		return nil, response.NewError(response.CodeConflict, "转办服务未启用")
	}
	pending, err := s.transfers.Pending(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return nil, response.NewError(response.CodeConflict, "已有进行中的转办，请等待负责人签字")
	}
	target, err := s.mustUser(ctx, req.TargetID)
	if err != nil {
		return nil, err
	}
	if target.ID == actor {
		return nil, response.NewError(response.CodeBadRequest, "请选择其他处理人")
	}
	if err := validateWorkflowAssignee(t, &target.ID); err != nil {
		return nil, err
	}
	kind, sourceDept, targetDept, sourceCenter, targetCenter, supervisor, err := s.classifyHandover(ctx, t, target)
	if err != nil {
		return nil, err
	}
	if kind == "internal" {
		return s.delegateExecution(ctx, t, actor, target.ID, strings.TrimSpace(req.Reason))
	}
	row := &model.TaskTransfer{
		ID: uuid.New(), TaskID: t.ID, FromUserID: actor, ToUserID: target.ID, InitiatorID: actor,
		SourceDepartmentID: sourceDept, TargetDepartmentID: targetDept,
		SourceCenterID: sourceCenter, TargetCenterID: targetCenter,
		SupervisorRole: supervisor, Kind: kind, Status: "pending", Revision: 1,
		Reason: strings.TrimSpace(req.Reason), PreviousStatus: t.Status, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.transfers.Create(ctx, row); err != nil {
		return nil, err
	}
	inst, err := s.flow.Start(ctx, engine.TaskTransferDefinitionKey(kind), row.ID.String(), engine.TaskTransferBusinessType, actor, map[string]interface{}{
		"task_id": t.ID.String(), "transfer_id": row.ID.String(),
	})
	if err != nil {
		return nil, err
	}
	row.WorkflowInstanceID = &inst.ID
	row.UpdatedAt = time.Now()
	if err := s.transfers.Save(ctx, row); err != nil {
		return nil, err
	}
	t.WorkflowRevision++
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	if err := s.addLog(ctx, t.ID, actor, "workflow_handover", actor.String(), target.ID.String(), row.Reason); err != nil {
		return nil, err
	}
	return s.handoverSnapshot(ctx, t, row, actor)
}

func (s *taskService) delegateExecution(ctx context.Context, t *model.Task, actor, target uuid.UUID, reason string) (*dto.HandoverResponse, error) {
	if err := s.flow.TransferTaskExecution(ctx, *t.WorkflowInstanceID, actor, target, reason); err != nil {
		return nil, err
	}
	old := t.AssigneeID.String()
	t.AssigneeID = &target
	t.WorkflowRevision++
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	if err := s.addLog(ctx, t.ID, actor, "workflow_delegate", old, target.String(), reason); err != nil {
		return nil, err
	}
	s.notifyUsers(ctx, []uuid.UUID{target, t.CreatorID}, tplTaskTransferred, t, reason)
	now := time.Now()
	row := &model.TaskTransfer{
		ID: uuid.New(), TaskID: t.ID, FromUserID: actor, ToUserID: target, InitiatorID: actor,
		Kind: "internal", Status: "completed", Revision: 1, Reason: reason,
		PreviousStatus: t.Status, CompletedAt: &now, CreatedAt: now, UpdatedAt: now,
	}
	if t.DepartmentID != nil {
		row.SourceDepartmentID, row.TargetDepartmentID = *t.DepartmentID, *t.DepartmentID
		row.SourceCenterID, row.TargetCenterID = *t.DepartmentID, *t.DepartmentID
	}
	if s.transfers != nil {
		if err := s.transfers.Create(ctx, row); err != nil {
			return nil, err
		}
	}
	return s.handoverSnapshot(ctx, t, row, actor)
}

func (s *taskService) decideHandover(ctx context.Context, t *model.Task, request *model.TaskTransfer, actor uuid.UUID, req *dto.HandoverDecision) (*dto.HandoverResponse, error) {
	if request == nil || request.Status != "pending" || request.WorkflowInstanceID == nil {
		return nil, response.NewError(response.CodeConflict, "当前没有待签字的转办")
	}
	if req == nil || strings.TrimSpace(req.Comment) == "" || utf8.RuneCountInString(req.Comment) > 2000 {
		return nil, response.NewError(response.CodeBadRequest, "请填写 1 至 2000 字签字意见")
	}
	if req.Revision != 0 && req.Revision != request.Revision {
		return nil, response.NewError(response.CodeConflict, "转办已更新，请刷新后再签字")
	}
	signer, err := s.transfers.Actor(ctx, actor)
	if err != nil {
		return nil, err
	}
	actual, waived := transferAuthority(signer, request, req.Requirement)
	if actual == "" {
		return nil, response.NewError(response.CodeForbidden, "当前账号不能签署该转办环节")
	}
	signatures, err := s.transfers.Signatures(ctx, request.ID)
	if err != nil {
		return nil, err
	}
	for _, item := range signatures {
		if item.Requirement == req.Requirement && item.Decision == "approve" {
			return nil, response.NewError(response.CodeConflict, "该转办环节已经签字")
		}
	}
	signature := &model.TransferSignature{
		ID: uuid.New(), TransferID: request.ID, Requirement: req.Requirement,
		SignerID: actor, SignerRole: actual, Waived: waived, Decision: req.Decision,
		Comment: strings.TrimSpace(req.Comment), CreatedAt: time.Now(),
	}
	if err := s.transfers.Sign(ctx, signature); err != nil {
		return nil, err
	}
	request.Revision++
	request.UpdatedAt = time.Now()
	if req.Decision == "reject" {
		if err := s.flow.Terminate(ctx, *request.WorkflowInstanceID, actor, signature.Comment); err != nil {
			return nil, err
		}
		request.Status = "rejected"
		if err := s.transfers.Save(ctx, request); err != nil {
			return nil, err
		}
		t.WorkflowRevision++
		t.UpdatedAt = time.Now()
		if err := s.tasks.Update(ctx, t); err != nil {
			return nil, err
		}
		if err := s.addLog(ctx, t.ID, actor, "workflow_handover_reject", request.FromUserID.String(), request.ToUserID.String(), signature.Comment); err != nil {
			return nil, err
		}
		return s.handoverSnapshot(ctx, t, request, actor)
	}
	if err := s.flow.CompleteTaskTransferApproval(ctx, *request.WorkflowInstanceID, req.Requirement, actor, signature.ID, waived); err != nil {
		return nil, err
	}
	_, done, err := s.flow.BusinessStage(ctx, *request.WorkflowInstanceID)
	if err != nil {
		return nil, err
	}
	if done {
		if err := s.finishSignedHandover(ctx, t, request, actor); err != nil {
			return nil, err
		}
	}
	if err := s.transfers.Save(ctx, request); err != nil {
		return nil, err
	}
	return s.handoverSnapshot(ctx, t, request, actor)
}

func (s *taskService) finishSignedHandover(ctx context.Context, t *model.Task, request *model.TaskTransfer, actor uuid.UUID) error {
	now := time.Now()
	if t.WorkflowStage != "execution" || t.AssigneeID == nil || *t.AssigneeID != request.FromUserID {
		if err := s.flow.Terminate(ctx, *request.WorkflowInstanceID, actor, "任务已离开执行环节，转办无法完成交接"); err != nil {
			return err
		}
		request.Status = "cancelled"
		request.CompletedAt = &now
		t.WorkflowRevision++
		t.UpdatedAt = now
		if err := s.tasks.Update(ctx, t); err != nil {
			return err
		}
		return s.addLog(ctx, t.ID, actor, "workflow_handover_cancel", request.FromUserID.String(), request.ToUserID.String(), "任务已离开执行环节，转办取消")
	}
	if err := s.flow.ReassignTaskExecution(ctx, *t.WorkflowInstanceID, request.FromUserID, request.ToUserID, actor, request.Reason); err != nil {
		return err
	}
	request.Status = "completed"
	request.CompletedAt = &now
	t.AssigneeID = &request.ToUserID
	t.WorkflowRevision++
	t.UpdatedAt = now
	if err := s.tasks.Update(ctx, t); err != nil {
		return err
	}
	if err := s.addLog(ctx, t.ID, actor, "workflow_handover_approve", request.FromUserID.String(), request.ToUserID.String(), request.Reason); err != nil {
		return err
	}
	s.notifyUsers(ctx, []uuid.UUID{request.ToUserID, t.CreatorID}, tplTaskTransferred, t, request.Reason)
	return nil
}

func (s *taskService) classifyHandover(ctx context.Context, t *model.Task, target *model.NamedUser) (kind string, sourceDept, targetDept, sourceCenter, targetCenter uuid.UUID, supervisor string, err error) {
	source := t.DepartmentID
	if source == nil {
		if actor, lookupErr := s.transfers.Actor(ctx, *t.AssigneeID); lookupErr != nil {
			return "", uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, "", lookupErr
		} else if actor != nil {
			source = actor.DepartmentID
		}
	}
	destination := target.DepartmentID
	if source == nil || destination == nil || *source == uuid.Nil || *destination == uuid.Nil {
		return "", uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, "", response.NewError(response.CodeConflict, "转办双方须有明确部门，无法按同部门直接委托")
	}
	src, err := s.transfers.Department(ctx, *source)
	if err != nil {
		return "", uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, "", err
	}
	dst, err := s.transfers.Department(ctx, *destination)
	if err != nil {
		return "", uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, "", err
	}
	if src == nil || dst == nil {
		return "", uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, "", response.NewError(response.CodeConflict, "转办双方部门不可用")
	}
	srcCenter, dstCenter := src.ID, dst.ID
	if src.ParentID != nil {
		srcCenter = *src.ParentID
	}
	if dst.ParentID != nil {
		dstCenter = *dst.ParentID
	}
	if src.ID == dst.ID {
		return "internal", src.ID, dst.ID, srcCenter, dstCenter, "minister", nil
	}
	if srcCenter == dstCenter {
		return "department", src.ID, dst.ID, srcCenter, dstCenter, "minister", nil
	}
	return "center", src.ID, dst.ID, srcCenter, dstCenter, "president", nil
}

func (s *taskService) pendingTransfer(ctx context.Context, id uuid.UUID) (*model.TaskTransfer, error) {
	if s.transfers == nil {
		return nil, nil
	}
	return s.transfers.Pending(ctx, id)
}

func (s *taskService) handoverVisible(ctx context.Context, id, actor uuid.UUID) (*model.Task, *model.TaskTransfer, error) {
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if t == nil {
		return nil, nil, response.NewError(response.CodeTaskNotFound, "任务不存在")
	}
	request, err := s.pendingTransfer(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if workflowParticipant(t, actor) {
		return t, request, nil
	}
	if request != nil && s.canSignHandover(ctx, request, actor) {
		return t, request, nil
	}
	if t.WorkflowInstanceID != nil {
		if _, err := s.workflowTask(ctx, id, actor); err == nil {
			return t, request, nil
		}
	}
	return nil, nil, response.NewError(response.CodeForbidden, "无权查看该转办")
}

func (s *taskService) canSignHandover(ctx context.Context, request *model.TaskTransfer, actor uuid.UUID) bool {
	if request == nil || request.Status != "pending" {
		return false
	}
	signer, err := s.transfers.Actor(ctx, actor)
	if err != nil || signer == nil {
		return false
	}
	for _, role := range engine.TaskTransferRoles(request.Kind) {
		if actual, _ := transferAuthority(signer, request, role); actual != "" {
			return true
		}
	}
	return false
}
