package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	tplActivityRegistered = "activity_registered"
	tplActivityWaitlist   = "activity_waitlist"
	tplActivityApproved   = "activity_approved"
	tplActivityRejected   = "activity_rejected"
	tplActivityStarting   = "activity_starting"
)

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
		Channels:     []string{"in_app", "websocket"},
	})
}

func (s *activityService) notifyActivity(ctx context.Context, userIDs []uuid.UUID, template string, a *model.Activity) {
	if s.notify == nil || len(userIDs) == 0 {
		return
	}
	vars := map[string]interface{}{
		"title":      a.Title,
		"start_time": a.StartTime.Format("2006-01-02 15:04"),
		"location":   a.Location,
	}
	if err := s.notify.Send(ctx, userIDs, template, vars); err != nil {
		logger.Warn("send activity notify failed", zap.Error(err), zap.String("tpl", template))
	}
}
