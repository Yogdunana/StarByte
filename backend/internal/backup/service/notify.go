package service

import (
	"context"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const tplBackupFailed = "backup_failed"

// Alerter surfaces backup failures via in-app notification when cheap.
type Alerter interface {
	Failed(ctx context.Context, userID *uuid.UUID, filename, errText string)
}

func NewNotifier(inner notifsvc.NotificationService) Alerter {
	if inner == nil {
		return noopAlerter{}
	}
	return &notifAlerter{inner: inner}
}

type notifAlerter struct{ inner notifsvc.NotificationService }

func (a *notifAlerter) Failed(ctx context.Context, userID *uuid.UUID, filename, errText string) {
	if userID == nil || *userID == uuid.Nil {
		logger.Error("backup job failed", zap.String("filename", filename), zap.String("error", errText))
		return
	}
	err := a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{*userID},
		TemplateCode: tplBackupFailed,
		Variables:    map[string]interface{}{"filename": filename, "error": errText},
		Channels:     []string{"in_app", "websocket"},
	})
	if err != nil {
		logger.Warn("backup failure notify failed", zap.Error(err))
	}
}

type noopAlerter struct{}

func (noopAlerter) Failed(_ context.Context, _ *uuid.UUID, filename, errText string) {
	logger.Error("backup job failed", zap.String("filename", filename), zap.String("error", errText))
}
