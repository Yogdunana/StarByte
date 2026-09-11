package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) create(ctx context.Context, operator uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}
	now := time.Now()
	t := &model.Task{
		ID:          uuid.New(),
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Status:      model.StatusPending,
		Priority:    req.Priority,
		CreatorID:   operator,
		DueDate:     req.DueDate,
		Tags:        encodeTags(req.Tags),
		CreatedAt:   now,
		UpdatedAt:   now,
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
		if model.IsClosed(parent.Status) {
			return nil, response.NewError(response.CodeTaskClosed, "不能向已关闭任务添加子任务")
		}
		if t.DepartmentID != nil && (parent.DepartmentID == nil || *t.DepartmentID != *parent.DepartmentID) {
			return nil, response.NewError(response.CodeBadRequest, "子任务必须沿用父任务部门")
		}
		t.DepartmentID = parent.DepartmentID
		t.ParentID = id
	}
	if err := s.validateDepartment(ctx, operator, t); err != nil {
		return nil, err
	}
	if req.Workflow != nil {
		if err := s.prepareWorkflow(ctx, t, req.Workflow); err != nil {
			return nil, err
		}
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	if err := s.addLog(ctx, t.ID, operator, model.ActionCreate, "", t.Title, ""); err != nil {
		return nil, err
	}
	if t.AssignmentPolicy != "" && t.AssignmentPolicy != "{}" {
		if err := s.addLog(ctx, t.ID, operator, "auto_assign", t.AssignmentPolicy, t.AssigneeID.String(), "根据发布时选定的自动分配规则指定执行人"); err != nil {
			return nil, err
		}
	}
	if t.AssigneeID != nil {
		s.notifyUsers(ctx, []uuid.UUID{*t.AssigneeID}, tplTaskAssigned, t, "")
		if err := s.addLog(ctx, t.ID, operator, model.ActionAssign, "", t.AssigneeID.String(), ""); err != nil {
			return nil, err
		}
	}
	if req.Workflow != nil {
		if err := s.startWorkflow(ctx, t); err != nil {
			return nil, err
		}
	}
	return s.taskResponse(ctx, operator, t.ID, nil)
}

func (s *taskService) List(ctx context.Context, viewer uuid.UUID, req *dto.ListTaskRequest, scope *rbacModel.DataScopeCondition) ([]*dto.TaskResponse, int64, error) {
	rows, total, err := s.tasks.List(ctx, req, rewriteTaskScope(scope, viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	out := make([]*dto.TaskResponse, 0, len(rows))
	for i := range rows {
		item := mapTask(&rows[i], nil)
		taskCapabilities(ctx, &rows[i].Task, item)
		out = append(out, item)
	}
	if err := s.filterParents(ctx, viewer, scope, out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *taskService) Get(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.TaskResponse, error) {
	if _, err := s.mustVisible(ctx, id, viewer, scope); err != nil {
		return nil, err
	}
	return s.taskResponse(ctx, viewer, id, scope)
}

func (s *taskService) taskResponse(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.TaskResponse, error) {
	if scope == nil {
		if v, ok := model.ViewerFromContext(ctx); ok {
			scope = v.Scope
		}
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
	visible := make([]model.Task, 0, len(children))
	for _, child := range children {
		if canViewTask(&child, viewer, scope) {
			visible = append(visible, child)
		}
	}
	out := mapTask(row, visible)
	taskCapabilities(ctx, &row.Task, out)
	if row.ParentID != nil {
		parent, err := s.tasks.GetByID(ctx, *row.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil || !canViewTask(parent, viewer, scope) {
			out.Parent = nil
		}
	}
	return out, nil
}

func (s *taskService) update(ctx context.Context, id, operator uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMutable(t, operator); err != nil {
		return nil, err
	}
	if err := validateUpdate(req); err != nil {
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
	if req.ClearDueDate {
		t.DueDate = nil
		t.DueRemindedAt = nil
		t.OverdueRemindedAt = nil
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
	return s.taskResponse(ctx, operator, id, nil)
}

func (s *taskService) delete(ctx context.Context, id, operator uuid.UUID) error {
	t, err := s.mustTask(ctx, id)
	if err != nil {
		return err
	}
	if t.WorkflowInstanceID != nil && !model.IsClosed(t.Status) {
		return response.NewError(response.CodeTaskInvalidState, "请先取消审批流程，再移除任务")
	}
	if t.Status == model.StatusDoing {
		return response.NewError(response.CodeTaskInvalidState, "进行中任务不可删除")
	}
	if t.CreatorID != operator && (t.AssigneeID == nil || *t.AssigneeID != operator) {
		return response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	children, err := s.tasks.ListChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return response.NewError(response.CodeTaskInvalidState, "请先处理子任务，再删除父任务")
	}
	if err := s.addLog(ctx, id, operator, "delete", "", "", "任务已移出日常列表，历史记录保留"); err != nil {
		return err
	}
	if err := s.tasks.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}
