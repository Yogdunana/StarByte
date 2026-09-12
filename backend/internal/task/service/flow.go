package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) assign(ctx context.Context, id, operator uuid.UUID, assigneeRaw string) (*dto.TaskResponse, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, ok := model.ViewerFromContext(ctx); !ok && !canViewTask(t, operator, nil) {
		return nil, response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	if !canMutate(t.Status) {
		return nil, response.NewError(response.CodeTaskClosed, "任务已关闭，无法操作")
	}
	assignee, err := s.mustUser(ctx, assigneeRaw)
	if err != nil {
		return nil, err
	}
	if t.AssigneeID != nil && *t.AssigneeID == assignee.ID {
		return s.taskResponse(ctx, operator, id, nil)
	}
	old := uuidPtrString(t.AssigneeID)
	if err := s.workflowAssignment(ctx, t, operator, assignee.ID); err != nil {
		return nil, err
	}
	t.AssigneeID = &assignee.ID
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("assign task: %w", err)
	}
	if err := s.addLog(ctx, t.ID, operator, model.ActionAssign, old, assignee.ID.String(), ""); err != nil {
		return nil, err
	}
	s.notifyUsers(ctx, []uuid.UUID{assignee.ID}, tplTaskAssigned, t, "")
	return s.taskResponse(ctx, operator, id, nil)
}

func (s *taskService) transfer(ctx context.Context, id, operator uuid.UUID, req *dto.TransferRequest) (*dto.TaskResponse, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMutable(t, operator); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Reason) == "" {
		return nil, response.NewError(response.CodeBadRequest, "请填写转办原因")
	}
	target, err := s.mustUser(ctx, req.NewAssigneeID)
	if err != nil {
		return nil, err
	}
	if t.AssigneeID != nil && *t.AssigneeID == target.ID {
		return s.taskResponse(ctx, operator, id, nil)
	}
	applied, err := s.workflowTransfer(ctx, t, operator, target.ID, req.Reason)
	if err != nil {
		return nil, err
	}
	if !applied {
		return s.taskResponse(ctx, operator, id, nil)
	}
	old := uuidPtrString(t.AssigneeID)
	t.AssigneeID = &target.ID
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("transfer task: %w", err)
	}
	if err := s.addLog(ctx, t.ID, operator, model.ActionTransfer, old, target.ID.String(), req.Reason); err != nil {
		return nil, err
	}
	s.notifyUsers(ctx, []uuid.UUID{target.ID}, tplTaskTransferred, t, req.Reason)
	return s.taskResponse(ctx, operator, id, nil)
}

func (s *taskService) changeStatus(ctx context.Context, id, operator uuid.UUID, req *dto.StatusRequest) (*dto.TaskResponse, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if model.IsClosed(t.Status) && t.Status != req.Status {
		return nil, response.NewError(response.CodeTaskClosed, "任务已关闭，无法操作")
	}
	if t.CreatorID != operator && (t.AssigneeID == nil || *t.AssigneeID != operator) {
		return nil, response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	if !CanTransit(t.Status, req.Status) {
		return nil, response.NewError(response.CodeTaskInvalidState, "任务状态不允许该操作")
	}
	if t.Status == req.Status {
		return s.taskResponse(ctx, operator, id, nil)
	}
	if err := s.workflowStatus(ctx, t, operator, req); err != nil {
		return nil, err
	}
	old := strconv.Itoa(int(t.Status))
	now := time.Now()
	t.Status = req.Status
	if req.Status == model.StatusDone {
		t.CompletedAt = &now
		t.Progress = 100
	} else {
		t.CompletedAt = nil
	}
	t.UpdatedAt = now
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("change status: %w", err)
	}
	if err := s.addLog(ctx, t.ID, operator, model.ActionStatusChange, old, strconv.Itoa(int(req.Status)), req.Comment); err != nil {
		return nil, err
	}
	return s.taskResponse(ctx, operator, id, nil)
}

func (s *taskService) urge(ctx context.Context, id, operator uuid.UUID, message string) error {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return err
	}
	if !canMutate(t.Status) {
		return response.NewError(response.CodeTaskClosed, "任务已关闭，无法操作")
	}
	if t.CreatorID != operator {
		return response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	if t.AssigneeID == nil {
		return response.NewError(response.CodeTaskInvalidState, "任务尚未分配负责人")
	}
	if err := s.addLog(ctx, t.ID, operator, model.ActionUrge, "", t.AssigneeID.String(), message); err != nil {
		return err
	}
	s.notifyUsers(ctx, []uuid.UUID{*t.AssigneeID}, tplTaskUrged, t, message)
	return nil
}

func (s *taskService) ListLogs(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.LogResponse, error) {
	if _, err := s.mustVisible(ctx, id, viewer, scope); err != nil {
		return nil, err
	}
	rows, err := s.logs.ListByTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}
	out := make([]dto.LogResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapLog(row))
	}
	return out, nil
}

func (s *taskService) ListMy(ctx context.Context, userID uuid.UUID, kind string, req *dto.MyTaskRequest) ([]*dto.TaskResponse, int64, error) {
	rows, total, err := s.tasks.ListMine(ctx, userID, kind, req)
	if err != nil {
		return nil, 0, fmt.Errorf("list my tasks: %w", err)
	}
	out := make([]*dto.TaskResponse, 0, len(rows))
	for i := range rows {
		item := mapTask(&rows[i], nil)
		taskCapabilities(ctx, &rows[i].Task, item)
		out = append(out, item)
	}
	if err := s.filterParents(ctx, userID, nil, out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *taskService) Stats(ctx context.Context, viewer uuid.UUID, req *dto.StatsRequest, scope *rbacModel.DataScopeCondition) (*dto.StatsResponse, error) {
	out, err := s.tasks.Stats(ctx, req, time.Now(), rewriteTaskScopeAlias(scope, viewer, ""))
	if err != nil {
		return nil, fmt.Errorf("task stats: %w", err)
	}
	return out, nil
}

func (s *taskService) mustUser(ctx context.Context, raw string) (*model.NamedUser, error) {
	id := parseUUIDPtr(raw)
	if id == nil {
		return nil, response.NewError(response.CodeTaskTargetGone, "转办目标用户不存在")
	}
	u, err := s.tasks.GetUser(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if u == nil {
		return nil, response.NewError(response.CodeTaskTargetGone, "转办目标用户不存在")
	}
	return u, nil
}

func (s *taskService) addLog(ctx context.Context, taskID, operator uuid.UUID, action, oldV, newV, comment string) error {
	return s.logs.Create(ctx, &model.TaskLog{
		ID:         uuid.New(),
		TaskID:     taskID,
		ActionType: action,
		OldValue:   oldV,
		NewValue:   newV,
		OperatorID: operator,
		Comment:    comment,
		CreatedAt:  time.Now(),
	})
}
