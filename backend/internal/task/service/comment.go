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

func (s *taskService) ListComments(ctx context.Context, viewer, taskID uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.CommentResponse, error) {
	if _, err := s.mustVisible(ctx, taskID, viewer, scope); err != nil {
		return nil, err
	}
	rows, err := s.comments.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	out := make([]dto.CommentResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapComment(row))
	}
	return out, nil
}

func (s *taskService) addComment(ctx context.Context, taskID, operator uuid.UUID, req *dto.CommentRequest) (*dto.CommentResponse, error) {
	t, err := s.mustTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if _, ok := model.ViewerFromContext(ctx); !ok && !canViewTask(t, operator, nil) {
		return nil, response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	if model.IsClosed(t.Status) {
		return nil, response.NewError(response.CodeTaskClosed, "任务已关闭，无法操作")
	}
	if req == nil || strings.TrimSpace(req.Content) == "" {
		return nil, response.NewError(response.CodeBadRequest, "评论内容不能为空")
	}
	ids, err := s.resolveMentions(ctx, req.Content, req.Mentions)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	c := &model.TaskComment{
		ID:        uuid.New(),
		TaskID:    taskID,
		AuthorID:  operator,
		Content:   req.Content,
		Mentions:  encodeJSONList(ids),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.comments.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	if err := s.addLog(ctx, taskID, operator, model.ActionComment, "", c.ID.String(), ""); err != nil {
		return nil, err
	}
	mentionUUIDs := parseUUIDList(ids)
	for _, id := range mentionUUIDs {
		if canViewTask(t, id, nil) {
			s.notifyUsers(ctx, []uuid.UUID{id}, tplTaskMention, t, req.Content)
		} else {
			redacted := *t
			redacted.Title = "协作任务"
			redacted.DueDate = nil
			s.notifyUsers(ctx, []uuid.UUID{id}, tplTaskMention, &redacted, "有人在任务讨论中提及你。内容仅对有权限的成员开放。")
		}
	}
	return s.getComment(ctx, c.ID)
}

func (s *taskService) updateComment(ctx context.Context, taskID, commentID, operator uuid.UUID, content string) (*dto.CommentResponse, error) {
	task, err := s.mustTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if model.IsClosed(task.Status) {
		return nil, response.NewError(response.CodeTaskClosed, "任务已关闭，无法修改评论")
	}
	c, err := s.mustComment(ctx, commentID, taskID)
	if err != nil {
		return nil, err
	}
	if c.AuthorID != operator {
		return nil, response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	if strings.TrimSpace(content) == "" {
		return nil, response.NewError(response.CodeBadRequest, "评论内容不能为空")
	}
	ids, err := s.resolveMentions(ctx, content, nil)
	if err != nil {
		return nil, err
	}
	c.Content = content
	c.Mentions = encodeJSONList(ids)
	c.UpdatedAt = time.Now()
	if err := s.comments.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("update comment: %w", err)
	}
	return s.getComment(ctx, c.ID)
}

func (s *taskService) deleteComment(ctx context.Context, taskID, commentID, operator uuid.UUID) error {
	task, err := s.mustTask(ctx, taskID)
	if err != nil {
		return err
	}
	if model.IsClosed(task.Status) {
		return response.NewError(response.CodeTaskClosed, "任务已关闭，无法修改评论")
	}
	c, err := s.mustComment(ctx, commentID, taskID)
	if err != nil {
		return err
	}
	if c.AuthorID != operator {
		return response.NewError(response.CodeTaskNoAccess, "无权操作该任务")
	}
	if err := s.comments.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return nil
}

func (s *taskService) mustComment(ctx context.Context, id, taskID uuid.UUID) (*model.TaskComment, error) {
	c, err := s.comments.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}
	if c == nil || c.TaskID != taskID {
		return nil, response.NewError(response.CodeTaskCommentGone, "评论不存在")
	}
	return c, nil
}

func (s *taskService) getComment(ctx context.Context, id uuid.UUID) (*dto.CommentResponse, error) {
	c, err := s.comments.GetByID(ctx, id)
	if err != nil || c == nil {
		return nil, response.NewError(response.CodeTaskCommentGone, "评论不存在")
	}
	u, err := s.tasks.GetUser(ctx, c.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("lookup comment author: %w", err)
	}
	named := model.TaskCommentNamed{TaskComment: *c, AuthorName: displayName(u)}
	out := mapComment(named)
	return &out, nil
}

func (s *taskService) resolveMentions(ctx context.Context, content string, extra []string) ([]string, error) {
	names := parseMentionNames(content)
	users, err := s.tasks.FindUsersByUsername(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("resolve mentioned usernames: %w", err)
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(users)+len(extra))
	for _, u := range users {
		id := u.ID.String()
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, raw := range extra {
		id := parseUUIDPtr(raw)
		if id == nil {
			continue
		}
		key := id.String()
		if _, ok := seen[key]; ok {
			continue
		}
		u, err := s.tasks.GetUser(ctx, *id)
		if err != nil {
			return nil, fmt.Errorf("resolve mentioned user: %w", err)
		}
		if u != nil {
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return out, nil
}

func parseUUIDList(raw []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		if id := parseUUIDPtr(s); id != nil {
			out = append(out, *id)
		}
	}
	return out
}
