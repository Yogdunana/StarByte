package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationService 通知服务接口
type NotificationService interface {
	Send(ctx context.Context, req *dto.SendNotificationRequest) error
	BatchSend(ctx context.Context, reqs []*dto.SendNotificationRequest) error
	ListByUser(ctx context.Context, userID uuid.UUID, req *dto.ListNotificationsRequest) ([]*model.Notification, int64, error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID, category string) error
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	Broadcast(ctx context.Context, req *dto.BroadcastNotificationRequest, onlineUserIDs []uuid.UUID) error
}

type notificationService struct {
	notificationRepo repo.NotificationRepo
	templateRepo     repo.NotificationTemplateRepo
	templateEngine   TemplateEngine
	channelRegistry  *ChannelRegistry
}

// NewNotificationService 创建通知服务
func NewNotificationService(
	notificationRepo repo.NotificationRepo,
	templateRepo repo.NotificationTemplateRepo,
	templateEngine TemplateEngine,
	channelRegistry *ChannelRegistry,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		templateEngine:   templateEngine,
		channelRegistry:  channelRegistry,
	}
}

// Send 发送通知（通过模板渲染 + 多渠道分发）
func (s *notificationService) Send(ctx context.Context, req *dto.SendNotificationRequest) error {
	// 1. 渲染模板
	rendered, err := s.templateEngine.Render(ctx, req.TemplateCode, req.Variables)
	if err != nil {
		return err
	}

	// 2. 获取模板配置（渠道 + 分类），仅在渠道未指定时查询
	category := "system"
	channels := req.Channels
	if len(channels) == 0 {
		tpl, err := s.templateRepo.GetByCode(ctx, req.TemplateCode)
		if err == nil {
			if tpl.Category != "" {
				category = tpl.Category
			}
			if tpl.Channels != "" {
				channels = tpl.GetChannels()
			}
		}
		if len(channels) == 0 {
			channels = []string{"in_app", "websocket"}
		}
	}

	// 3. 为每个用户发送通知（收集所有错误，不因单个用户失败而中断）
	var sendErrs []error
	for _, userID := range req.UserIDs {
		msg := &NotificationMessage{
			UserID:   userID,
			Title:    rendered.Title,
			Content:  rendered.Content,
			Category: category,
			Priority: "normal",
		}
		if errs := s.channelRegistry.SendViaChannels(ctx, msg, channels); len(errs) > 0 {
			sendErrs = append(sendErrs, fmt.Errorf("user %s: %w", userID, errs[0]))
		}
	}
	if len(sendErrs) > 0 {
		return fmt.Errorf("send notification failed, %d/%d users failed: %w",
			len(sendErrs), len(req.UserIDs), sendErrs[0])
	}
	return nil
}

// BatchSend 批量发送通知
func (s *notificationService) BatchSend(ctx context.Context, reqs []*dto.SendNotificationRequest) error {
	for _, req := range reqs {
		if err := s.Send(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ListByUser 查询用户通知列表
func (s *notificationService) ListByUser(ctx context.Context, userID uuid.UUID, req *dto.ListNotificationsRequest) ([]*model.Notification, int64, error) {
	return s.notificationRepo.ListByUser(ctx, userID, req.Page, req.PageSize, req.Category, req.UnreadOnly)
}

// MarkAsRead 标记通知为已读
func (s *notificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	return s.notificationRepo.MarkAsRead(ctx, userID, ids)
}

// MarkAllAsRead 标记用户全部通知为已读
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID, category string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID, category)
}

// Delete 删除通知
func (s *notificationService) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	// 先检查通知是否属于该用户
	n, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return response.NewError(response.CodeNotificationNotFound, "通知不存在")
	}
	if n.UserID != userID {
		return response.NewError(response.CodeNotificationNoAccess, "无权操作该通知")
	}
	return s.notificationRepo.Delete(ctx, userID, id)
}

// GetUnreadCount 获取用户未读通知数
func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationRepo.GetUnreadCount(ctx, userID)
}

// Broadcast 广播通知（给所有在线用户）
func (s *notificationService) Broadcast(ctx context.Context, req *dto.BroadcastNotificationRequest, onlineUserIDs []uuid.UUID) error {
	channels := req.Channels
	if len(channels) == 0 {
		channels = []string{"in_app", "websocket"}
	}

	// 批量创建站内通知（仅在 channels 包含 in_app 时）
	hasInApp := false
	for _, ch := range channels {
		if ch == "in_app" {
			hasInApp = true
			break
		}
	}
	if hasInApp && len(onlineUserIDs) > 0 {
		var notifications []*model.Notification
		for _, userID := range onlineUserIDs {
			notifications = append(notifications, &model.Notification{
				ID:       uuid.New(),
				UserID:   userID,
				Title:    req.Title,
				Content:  req.Content,
				Category: req.Category,
				Priority: req.Priority,
				IsRead:   false,
			})
		}
		if err := s.notificationRepo.BatchCreate(ctx, notifications); err != nil {
			return fmt.Errorf("batch create notifications: %w", err)
		}
	}

	// 通过 channelRegistry 发送非 in_app 渠道（websocket、email 等）
	// 过滤掉 in_app，因为已经通过 BatchCreate 批量创建了
	var otherChannels []string
	for _, ch := range channels {
		if ch != "in_app" {
			otherChannels = append(otherChannels, ch)
		}
	}
	for _, userID := range onlineUserIDs {
		msg := &NotificationMessage{
			UserID:   userID,
			Title:    req.Title,
			Content:  req.Content,
			Category: req.Category,
			Priority: req.Priority,
		}
		if len(otherChannels) > 0 {
			if errs := s.channelRegistry.SendViaChannels(ctx, msg, otherChannels); len(errs) > 0 {
				// 广播场景下不因个别渠道失败而中断
				logger.Error("broadcast notification via channels failed",
					zap.String("user_id", userID.String()),
					zap.Int("error_count", len(errs)))
			}
		}
	}

	return nil
}
