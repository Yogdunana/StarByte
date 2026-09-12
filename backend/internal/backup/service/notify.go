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

// RecipientLookup finds ops users who should receive backup failure alerts.
type RecipientLookup interface {
	ListOpsUserIDs(ctx context.Context) ([]uuid.UUID, error)
}

// Alerter surfaces backup failures via in-app / websocket / email when cheap.
type Alerter interface {
	Failed(ctx context.Context, userID *uuid.UUID, filename, errText string)
}

func NewNotifier(inner notifsvc.NotificationService, lookup RecipientLookup) Alerter {
	if inner == nil {
		return noopAlerter{}
	}
	return &notifAlerter{inner: inner, lookup: lookup}
}

type notifAlerter struct {
	inner  notifsvc.NotificationService
	lookup RecipientLookup
}

func collectAlertUserIDs(created *uuid.UUID, ops []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ops)+1)
	var out []uuid.UUID
	add := func(id uuid.UUID) {
		if id == uuid.Nil {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if created != nil {
		add(*created)
	}
	for _, id := range ops {
		add(id)
	}
	return out
}

func (a *notifAlerter) Failed(ctx context.Context, userID *uuid.UUID, filename, errText string) {
	var ops []uuid.UUID
	if a.lookup != nil {
		extra, err := a.lookup.ListOpsUserIDs(ctx)
		if err != nil {
			logger.Warn("backup alert resolve ops users failed", zap.Error(err))
		} else {
			ops = extra
		}
	}
	ids := collectAlertUserIDs(userID, ops)
	if len(ids) == 0 {
		logger.Error("backup job failed", zap.String("filename", filename), zap.String("error", errText))
		return
	}
	// NotificationService.Send 会按 users.email 填充邮件渠道；无邮箱则跳过 email，站内信仍发。
	err := a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      ids,
		TemplateCode: tplBackupFailed,
		Variables:    map[string]interface{}{"filename": filename, "error": errText},
		Channels:     []string{"in_app", "websocket", "email"},
	})
	if err != nil {
		logger.Warn("backup failure notify failed", zap.Error(err))
	}
}

type noopAlerter struct{}

func (noopAlerter) Failed(_ context.Context, _ *uuid.UUID, filename, errText string) {
	logger.Error("backup job failed", zap.String("filename", filename), zap.String("error", errText))
}
