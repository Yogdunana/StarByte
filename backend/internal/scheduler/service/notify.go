package service

import (
	"context"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const tplSchedulerFailed = "scheduler_task_failed"

type Alerter interface {
	TaskFailed(ctx context.Context, userID uuid.UUID, taskName, errText string)
}

type notifAlerter struct{ inner notifsvc.NotificationService }

func NewNotifAlerter(inner notifsvc.NotificationService) Alerter {
	if inner == nil {
		return nil
	}
	return &notifAlerter{inner: inner}
}

func (a *notifAlerter) TaskFailed(ctx context.Context, userID uuid.UUID, taskName, errText string) {
	if userID == uuid.Nil {
		return
	}
	err := a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{userID},
		TemplateCode: tplSchedulerFailed,
		Variables: map[string]interface{}{
			"task_name": taskName,
			"error":     errText,
		},
		Channels: []string{"in_app", "websocket"},
	})
	if err != nil {
		logger.Warn("scheduler fail notify failed", zap.Error(err))
	}
}
