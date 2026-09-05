package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *taskService) Create(ctx context.Context, operator uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	now := time.Now()
	t := &model.Task{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      model.StatusPending,
		Priority:    req.Priority,
		CreatorID:   operator,
		DueDate:     req.DueDate,
		Tags:        encodeTags(req.Tags),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if t.Priority < 0 || t.Priority > 3 {
		t.Priority = model.PriorityMedium
	}
	if id := parseUUIDPtr(req.AssigneeID); id != nil {
		if u, err := s.tasks.GetUser(ctx, *id); err != nil {
			return nil, fmt.Errorf("lookup assignee: %w", err)
		} else if u == nil {
			return nil, response.NewError(response.CodeTaskTargetGone, "转办目标用户不存在")
		}
		t.AssigneeID = id
	}
	if id := parseUUIDPtr(req.DepartmentID); id != nil {
		t.DepartmentID = id
	}
	if id := parseUUIDPtr(req.ParentID); id != nil {
		parent, err := s.mustTask(ctx, *id)
		if err != nil {
			return nil, err
		}
		if parent.ParentID != nil {
			return nil, response.NewError(response.CodeBadRequest, "仅支持一层子任务")
		}
		t.ParentID = id
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	s.addLog(ctx, t.ID, operator, model.ActionCreate, "", t.Title, "")
	if t.AssigneeID != nil {
		s.notifyUsers(ctx, []uuid.UUID{*t.AssigneeID}, tplTaskAssigned, t, "")
		s.addLog(ctx, t.ID, operator, model.ActionAssign, "", t.AssigneeID.String(), "")
	}
	return s.Get(ctx, operator, t.ID, nil)
}

func (s *taskService) List(ctx context.Context, viewer uuid.UUID, req *dto.ListTaskRequest, scope *rbacModel.DataScopeCondition) ([]*dto.TaskResponse, int64, error) {
	rows, total, err := s.tasks.List(ctx, req, rewriteTaskScope(scope, viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	out := make([]*dto.TaskResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapTask(&rows[i], nil))
	}
	return out, total, nil
}

func (s *taskService) Get(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.TaskResponse, error) {
	if _, err := s.mustVisible(ctx, id, viewer, scope); err != nil {
		return nil, err
	}
	row, err := s.tasks.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeTaskNotFound, "任务不存在")
	}
	children, err := s.tasks.ListChildren(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list children: %w", err)
	}
	return mapTask(row, children), nil
}

func (s *taskService) Update(ctx context.Context, id, operator uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMutable(t, operator); err != nil {
		return nil, err
	}
	if req.Title != nil {
		t.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Priority != nil {
		if *req.Priority < 0 || *req.Priority > 3 {
			return nil, response.NewError(response.CodeBadRequest, "无效优先级")
		}
		t.Priority = *req.Priority
	}
	if req.DueDate != nil {
		t.DueDate = req.DueDate
		t.DueRemindedAt = nil
		t.OverdueRemindedAt = nil
	}
	if req.Tags != nil {
		t.Tags = encodeTags(req.Tags)
	}
	t.UpdatedAt = time.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return s.Get(ctx, operator, id, nil)
}

func (s *taskService) Delete(ctx context.Context, id, operator uuid.UUID) error {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return err
	}
	if t.Status == model.StatusDoing {
		return response.NewError(response.CodeTaskInvalidState, "进行中任务不可删除")
	}
	if t.CreatorID != operator && (t.AssigneeID == nil || *t.AssigneeID != operator) {
		return response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	if err := s.tasks.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (s *taskService) mustTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if t == nil {
		return nil, response.NewError(response.CodeTaskNotFound, "任务不存在")
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
	if model.IsClosed(t.Status) {
		return response.NewError(response.CodeTaskClosed, "任务已关闭，无法操作")
	}
	if t.CreatorID != operator && (t.AssigneeID == nil || *t.AssigneeID != operator) {
		return response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	return nil
}

func rewriteTaskScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	return rewriteTaskScopeAlias(scope, userID, "t")
}

func rewriteTaskScopeAlias(scope *rbacModel.DataScopeCondition, userID uuid.UUID, alias string) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return scope
	}
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	if scope.Query == "1 = 0" {
		return &rbacModel.DataScopeCondition{
			Query: prefix + "creator_id = ? OR " + prefix + "assignee_id = ?",
			Args:  []interface{}{userID, userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", prefix+"department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func canViewTask(t *model.Task, viewer uuid.UUID, scope *rbacModel.DataScopeCondition) bool {
	rewritten := rewriteTaskScopeAlias(scope, viewer, "")
	if rewritten == nil || rewritten.IsEmpty() {
		return true
	}
	if strings.Contains(rewritten.Query, "creator_id = ? OR") {
		if t.CreatorID == viewer {
			return true
		}
		return t.AssigneeID != nil && *t.AssigneeID == viewer
	}
	if t.DepartmentID == nil {
		return false
	}
	for _, arg := range rewritten.Args {
		switch v := arg.(type) {
		case uuid.UUID:
			if v == *t.DepartmentID {
				return true
			}
		case []uuid.UUID:
			for _, id := range v {
				if id == *t.DepartmentID {
					return true
				}
			}
		}
	}
	return false
}
