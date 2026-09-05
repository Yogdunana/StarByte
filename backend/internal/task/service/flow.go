package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *taskService) Assign(ctx context.Context, id, operator uuid.UUID, assigneeRaw string) (*dto.TaskResponse, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canMutate(t.Status) {
		return nil, response.NewError(response.CodeTaskClosed, "任务已关闭，无法操作")
	}
	assignee, err := s.mustUser(ctx, assigneeRaw)
	if err != nil {
		return nil, err
	}
	old := uuidPtrString(t.AssigneeID)
	t.AssigneeID = &assignee.ID
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("assign task: %w", err)
	}
	s.addLog(ctx, t.ID, operator, model.ActionAssign, old, assignee.ID.String(), "")
	s.notifyUsers(ctx, []uuid.UUID{assignee.ID}, tplTaskAssigned, t, "")
	return s.Get(ctx, id)
}

func (s *taskService) Transfer(ctx context.Context, id, operator uuid.UUID, req *dto.TransferRequest) (*dto.TaskResponse, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMutable(t, operator); err != nil {
		return nil, err
	}
	target, err := s.mustUser(ctx, req.NewAssigneeID)
	if err != nil {
		return nil, err
	}
	old := uuidPtrString(t.AssigneeID)
	t.AssigneeID = &target.ID
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("transfer task: %w", err)
	}
	s.addLog(ctx, t.ID, operator, model.ActionTransfer, old, target.ID.String(), req.Reason)
	s.notifyUsers(ctx, []uuid.UUID{target.ID}, tplTaskTransferred, t, req.Reason)
	return s.Get(ctx, id)
}

func (s *taskService) ChangeStatus(ctx context.Context, id, operator uuid.UUID, req *dto.StatusRequest) (*dto.TaskResponse, error) {
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
	s.addLog(ctx, t.ID, operator, model.ActionStatusChange, old, strconv.Itoa(int(req.Status)), req.Comment)
	return s.Get(ctx, id)
}

func (s *taskService) Urge(ctx context.Context, id, operator uuid.UUID, message string) error {
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
	s.addLog(ctx, t.ID, operator, model.ActionUrge, "", t.AssigneeID.String(), message)
	s.notifyUsers(ctx, []uuid.UUID{*t.AssigneeID}, tplTaskUrged, t, message)
	return nil
}

func (s *taskService) ListLogs(ctx context.Context, id uuid.UUID) ([]dto.LogResponse, error) {
	if _, err := s.mustTask(ctx, id); err != nil {
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
		out = append(out, mapTask(&rows[i], nil))
	}
	return out, total, nil
}

func (s *taskService) Stats(ctx context.Context, req *dto.StatsRequest) (*dto.StatsResponse, error) {
	out, err := s.tasks.Stats(ctx, req, time.Now())
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

func (s *taskService) addLog(ctx context.Context, taskID, operator uuid.UUID, action, oldV, newV, comment string) {
	_ = s.logs.Create(ctx, &model.TaskLog{
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
