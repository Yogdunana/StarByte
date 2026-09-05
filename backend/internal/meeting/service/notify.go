package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	tplMeetingNotice  = "meeting_notice"
	tplMeetingStarted = "meeting_started"
	tplMeetingEnded   = "meeting_ended"
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

func (s *meetingService) notifyMeeting(ctx context.Context, userIDs []uuid.UUID, template string, m *model.Meeting) {
	if s.notify == nil || len(userIDs) == 0 {
		return
	}
	vars := map[string]interface{}{
		"title":      m.Title,
		"start_time": m.StartTime.Format("2006-01-02 15:04"),
		"location":   m.Location,
	}
	if err := s.notify.Send(ctx, userIDs, template, vars); err != nil {
		logger.Warn("send meeting notify failed", zap.Error(err), zap.String("tpl", template))
	}
}
