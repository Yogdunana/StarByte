package service

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func assignmentPolicy(config *dto.AutoAssignmentConfig, department *uuid.UUID) (*model.AssignmentPolicy, error) {
	if config == nil || config.Mode == "" || config.Mode == "manual" {
		return nil, nil
	}
	if config.Mode != "department" && config.Mode != "role" && config.Mode != "round_robin" {
		return nil, response.NewError(response.CodeBadRequest, "未知自动分配方式")
	}
	if department == nil || *department == uuid.Nil {
		return nil, response.NewError(response.CodeBadRequest, "自动分配需要明确的任务归属部门")
	}
	policy := &model.AssignmentPolicy{Mode: config.Mode, DepartmentID: *department}
	if config.RoleID != "" {
		id, err := uuid.Parse(config.RoleID)
		if err != nil || id == uuid.Nil {
			return nil, response.NewError(response.CodeBadRequest, "无效的分配角色")
		}
		policy.RoleID = &id
	}
	if config.Mode == "role" && policy.RoleID == nil {
		return nil, response.NewError(response.CodeBadRequest, "按角色分配需要选择角色")
	}
	return policy, nil
}
func (s *taskService) autoAssign(ctx context.Context, t *model.Task, config *dto.AutoAssignmentConfig) error {
	policy, err := assignmentPolicy(config, t.DepartmentID)
	if err != nil || policy == nil {
		return err
	}
	if t.AssigneeID != nil {
		return response.NewError(response.CodeBadRequest, "自动分配和手动指定执行人不能同时使用")
	}
	if s.assignments == nil || s.afterCommit == nil {
		return response.NewError(response.CodeConflict, "自动分配服务未启用")
	}
	excluded := []uuid.UUID{}
	for _, id := range []*uuid.UUID{t.ReviewerID, t.AcceptorID} {
		if id != nil {
			excluded = append(excluded, *id)
		}
	}
	user, err := s.assignments.Pick(ctx, *policy, excluded)
	if err != nil {
		return err
	}
	if user == nil {
		return response.NewError(response.CodeConflict, "该部门或角色没有可用执行人；审核人、验收人和停用账号不会被分配")
	}
	encoded, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	t.AssignmentPolicy = string(encoded)
	t.AssigneeID = &user.ID
	return nil
}
func (s *taskService) AssignmentRoles(ctx context.Context, keyword string) ([]dto.Person, error) {
	keyword = strings.TrimSpace(keyword)
	if utf8.RuneCountInString(keyword) > 100 {
		return nil, response.NewError(response.CodeBadRequest, "搜索词过长")
	}
	if s.assignments == nil {
		return nil, response.NewError(response.CodeConflict, "分配规则服务未启用")
	}
	roles, err := s.assignments.Roles(ctx, keyword)
	if err != nil {
		return nil, err
	}
	out := make([]dto.Person, 0, len(roles))
	for _, r := range roles {
		out = append(out, dto.Person{ID: r.ID.String(), Name: r.Name})
	}
	return out, nil
}
