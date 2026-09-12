package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const tplAnnouncementPublished = "announcement_published"

type notificationAdapter struct {
	inner notifsvc.NotificationService
}

func NewNotifier(inner notifsvc.NotificationService) Notifier {
	if inner == nil {
		return nil
	}
	return &notificationAdapter{inner: inner}
}

func (a *notificationAdapter) Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error {
	if len(userIDs) == 0 {
		return nil
	}
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      userIDs,
		TemplateCode: template,
		Variables:    vars,
		// 复用 #4/#49：站内 + WS + 邮件。SMTP 未配置时 email 渠道自动跳过。
		Channels: []string{"in_app", "websocket", "email"},
	})
}

func (s *announcementService) notifyPublished(ctx context.Context, a *model.Announcement) {
	if s.notify == nil || a == nil {
		return
	}
	ids, err := s.recipients(ctx, a)
	if err != nil {
		logger.Warn("list announcement recipients failed", zap.Error(err))
		return
	}
	if len(ids) == 0 {
		return
	}
	vars := map[string]interface{}{
		"title":    a.Title,
		"category": a.Category,
		"id":       a.ID.String(),
	}
	if err := s.notify.Send(ctx, ids, tplAnnouncementPublished, vars); err != nil {
		logger.Warn("send announcement notify failed", zap.Error(err), zap.String("id", a.ID.String()))
	}
}
