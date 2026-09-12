package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *announcementService) Create(ctx context.Context, viewer Viewer, req *dto.CreateAnnouncementRequest) (*dto.AnnouncementResponse, error) {
	if req == nil {
		return nil, response.NewError(response.CodeBadRequest, "参数错误")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, response.NewError(response.CodeBadRequest, "标题不能为空")
	}
	if !model.ValidCategory(req.Category) {
		return nil, invalidCategory()
	}
	contentType := req.ContentType
	if contentType == "" {
		contentType = model.ContentMarkdown
	}
	if !model.ValidContentType(contentType) {
		return nil, response.NewError(response.CodeBadRequest, "正文类型不合法")
	}
	if len(req.Content) > model.MaxContentLen {
		return nil, response.NewError(response.CodeBadRequest, "正文过长")
	}
	if err := validateSchedule(req.ScheduledAt, viewer.CanPublish, s.clock()); err != nil {
		return nil, err
	}

	now := s.clock()
	pinned := false
	if viewer.CanManage {
		pinned = req.Pinned
	}
	a := &model.Announcement{
		ID:          uuid.New(),
		Title:       title,
		Content:     req.Content,
		ContentType: contentType,
		Category:    req.Category,
		Pinned:      pinned,
		Required:    req.Required,
		Status:      model.StatusDraft,
		ScheduledAt: req.ScheduledAt,
		AuthorID:    viewer.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.rows.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("create announcement: %w", err)
	}
	return s.getResponse(ctx, a.ID, viewer.UserID)
}

func (s *announcementService) Update(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpdateAnnouncementRequest) (*dto.AnnouncementResponse, error) {
	if req == nil {
		return nil, response.NewError(response.CodeBadRequest, "参数错误")
	}
	a, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == model.StatusArchived {
		return nil, invalidState("已归档公告不能修改")
	}
	if !canEdit(viewer, a) {
		return nil, noAccess("无权修改该公告")
	}
	if err := applyUpdate(a, req, viewer.CanPublish, s.clock()); err != nil {
		return nil, err
	}
	a.UpdatedAt = s.clock()
	if err := s.rows.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update announcement: %w", err)
	}
	return s.getResponse(ctx, a.ID, viewer.UserID)
}

func applyUpdate(a *model.Announcement, req *dto.UpdateAnnouncementRequest, canPublish bool, now time.Time) error {
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return response.NewError(response.CodeBadRequest, "标题不能为空")
		}
		if len(title) > 200 {
			return response.NewError(response.CodeBadRequest, "标题过长")
		}
		a.Title = title
	}
	if req.Content != nil {
		if len(*req.Content) > model.MaxContentLen {
			return response.NewError(response.CodeBadRequest, "正文过长")
		}
		a.Content = *req.Content
	}
	if req.ContentType != nil {
		if !model.ValidContentType(*req.ContentType) {
			return response.NewError(response.CodeBadRequest, "正文类型不合法")
		}
		a.ContentType = *req.ContentType
	}
	if req.Category != nil {
		if !model.ValidCategory(*req.Category) {
			return invalidCategory()
		}
		a.Category = *req.Category
	}
	if req.Required != nil {
		a.Required = *req.Required
	}
	if req.ClearSched || (req.ScheduledAt != nil && a.Status != model.StatusDraft) {
		if !canPublish {
			return nil
		}
		a.ScheduledAt = nil
		return nil
	}
	if req.ScheduledAt != nil {
		if err := validateSchedule(req.ScheduledAt, canPublish, now); err != nil {
			return err
		}
		a.ScheduledAt = req.ScheduledAt
	}
	return nil
}

func validateSchedule(at *time.Time, canPublish bool, now time.Time) error {
	if at == nil {
		return nil
	}
	if !canPublish {
		return noAccess("无权设置定时发布")
	}
	if !at.After(now) {
		return response.NewError(response.CodeBadRequest, "定时发布时间必须晚于当前时间")
	}
	return nil
}

func (s *announcementService) Delete(ctx context.Context, viewer Viewer, id uuid.UUID) error {
	a, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	if !canDelete(viewer, a) {
		return noAccess("无权删除该公告")
	}
	return s.rows.Delete(ctx, id)
}

func (s *announcementService) Get(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.AnnouncementResponse, error) {
	row, err := s.rows.GetByIDNamed(ctx, id, viewer.UserID)
	if err != nil {
		return nil, fmt.Errorf("get announcement: %w", err)
	}
	if row == nil {
		return nil, notFound()
	}
	if !canView(viewer, &row.Announcement) {
		return nil, noAccess("无权查看该公告")
	}
	return toResponse(row), nil
}

func (s *announcementService) List(ctx context.Context, viewer Viewer, req *dto.ListAnnouncementRequest) ([]*dto.AnnouncementResponse, int64, error) {
	if req == nil {
		req = &dto.ListAnnouncementRequest{}
	}
	if req.Category != "" && !model.ValidCategory(req.Category) {
		return nil, 0, invalidCategory()
	}
	if req.Status != nil && !model.ValidStatus(*req.Status) {
		return nil, 0, response.NewError(response.CodeBadRequest, "状态不合法")
	}
	rows, total, err := s.rows.List(ctx, viewer.UserID, viewer.Staff, viewer.CanManage, req)
	if err != nil {
		return nil, 0, fmt.Errorf("list announcements: %w", err)
	}
	list := make([]*dto.AnnouncementResponse, 0, len(rows))
	for i := range rows {
		list = append(list, toResponse(&rows[i]))
	}
	return list, total, nil
}

func (s *announcementService) getResponse(ctx context.Context, id, viewer uuid.UUID) (*dto.AnnouncementResponse, error) {
	row, err := s.rows.GetByIDNamed(ctx, id, viewer)
	if err != nil {
		return nil, fmt.Errorf("get announcement: %w", err)
	}
	if row == nil {
		return nil, notFound()
	}
	return toResponse(row), nil
}

func isAuthor(v Viewer, a *model.Announcement) bool {
	return v.UserID == a.AuthorID
}

func canView(v Viewer, a *model.Announcement) bool {
	if a.Status == model.StatusPublished || a.Status == model.StatusArchived {
		return true
	}
	return v.CanManage || isAuthor(v, a)
}

func canEdit(v Viewer, a *model.Announcement) bool {
	if a.Status == model.StatusArchived {
		return false
	}
	return v.CanManage || isAuthor(v, a)
}

func canDelete(v Viewer, a *model.Announcement) bool {
	return v.CanManage || isAuthor(v, a)
}

func canPublishDraft(v Viewer, a *model.Announcement) bool {
	if isAuthor(v, a) {
		return v.CanPublish
	}
	return v.CanManage
}
