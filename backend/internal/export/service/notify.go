package service

import (
	"context"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const tplExportReady = "export_ready"

type notifReady struct {
	inner notifsvc.NotificationService
}

func NewNotifReady(inner notifsvc.NotificationService) ReadyNotifier {
	if inner == nil {
		return nil
	}
	return &notifReady{inner: inner}
}

func (n *notifReady) NotifyReady(ctx context.Context, userID, filename, fileID string) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return
	}
	err = n.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{uid},
		TemplateCode: tplExportReady,
		Variables: map[string]interface{}{
			"filename": filename,
			"file_id":  fileID,
		},
		Channels: []string{"in_app", "websocket"},
	})
	if err != nil {
		logger.Warn("export ready notify failed", zap.Error(err))
	}
}
