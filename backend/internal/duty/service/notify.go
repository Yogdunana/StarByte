package service

import (
	"context"
	"time"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	tplDutyReminder  = "duty_reminder"
	tplSwapApproved  = "swap_approved"
	tplSwapRejected  = "swap_rejected"
)

type dutyNotifierAdapter struct {
	inner notifsvc.NotificationService
}

func NewDutyNotifier(inner notifsvc.NotificationService) Notifier {
	if inner == nil {
		return nil
	}
	return &dutyNotifierAdapter{inner: inner}
}

func (a *dutyNotifierAdapter) NotifyDutyReminder(ctx context.Context, userID uuid.UUID, dutyDate time.Time, timeSlot, location string) error {
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{userID},
		TemplateCode: tplDutyReminder,
		Variables: map[string]interface{}{
			"duty_date": dutyDate.Format("2006-01-02"),
			"time_slot": timeSlot,
			"location":  location,
		},
		Channels: []string{"in_app", "websocket"},
	})
}

func (a *dutyNotifierAdapter) NotifySwapResult(ctx context.Context, userID uuid.UUID, approved bool, remark string) error {
	tpl := tplSwapApproved
	if !approved {
		tpl = tplSwapRejected
	}
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{userID},
		TemplateCode: tpl,
		Variables: map[string]interface{}{
			"remark": remark,
		},
		Channels: []string{"in_app", "websocket"},
	})
}

func (s *dutyService) notifyDutyReminder(ctx context.Context, userID uuid.UUID, dutyDate time.Time, timeSlot, location string) {
	if s.notifier == nil {
		return
	}
	if err := s.notifier.NotifyDutyReminder(ctx, userID, dutyDate, timeSlot, location); err != nil {
		logger.Warn("send duty reminder failed", zap.Error(err))
	}
}
