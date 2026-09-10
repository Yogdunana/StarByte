package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func validateTitle(title string) error {
	n := utf8.RuneCountInString(strings.TrimSpace(title))
	if n < 1 || n > 200 {
		return response.NewError(response.CodeBadRequest, "任务标题须为 1 至 200 字")
	}
	return nil
}
func validateTags(tags []string) error {
	if utf8.RuneCountInString(encodeTags(tags)) > 500 {
		return response.NewError(response.CodeBadRequest, "任务标签总长度超出限制")
	}
	return nil
}
func validateCreate(req *dto.CreateTaskRequest) error {
	if req == nil {
		return response.NewError(response.CodeBadRequest, "缺少任务内容")
	}
	if err := validateTitle(req.Title); err != nil {
		return err
	}
	if req.Priority < 0 || req.Priority > 3 {
		return response.NewError(response.CodeBadRequest, "无效优先级")
	}
	for _, raw := range []string{req.AssigneeID, req.DepartmentID, req.ParentID} {
		if raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil || id == uuid.Nil {
				return response.NewError(response.CodeBadRequest, "无效的用户、部门或父任务 ID")
			}
		}
	}
	return validateTags(req.Tags)
}
func validateUpdate(req *dto.UpdateTaskRequest) error {
	if req == nil {
		return response.NewError(response.CodeBadRequest, "缺少任务内容")
	}
	if req.ClearDueDate && req.DueDate != nil {
		return response.NewError(response.CodeBadRequest, "不能同时清空和设置截止时间")
	}
	if req.Title != nil {
		if err := validateTitle(*req.Title); err != nil {
			return err
		}
	}
	return validateTags(req.Tags)
}
func (s *taskService) validateDepartment(ctx context.Context, operator uuid.UUID, t *model.Task) error {
	v, ok := model.ViewerFromContext(ctx)
	if !ok {
		return nil
	}
	if v.ID != operator {
		return response.NewError(response.CodeTaskNoAccess, "用户身份不匹配")
	}
	u, err := s.tasks.GetUser(ctx, operator)
	if err != nil {
		return err
	}
	if u == nil {
		return response.NewError(response.CodeTaskTargetGone, "当前用户不可用")
	}
	if t.DepartmentID == nil {
		if t.ParentID != nil {
			return nil
		}
		t.DepartmentID = u.DepartmentID
		return nil
	}
	if u.DepartmentID != nil && *u.DepartmentID == *t.DepartmentID {
		return nil
	}
	// A creator cannot use their own ownership to bypass the department scope.
	probe := &model.Task{DepartmentID: t.DepartmentID}
	if !canViewTask(probe, operator, v.Scope) {
		return response.NewError(response.CodeTaskNoAccess, "无权在该部门创建任务")
	}
	return nil
}
