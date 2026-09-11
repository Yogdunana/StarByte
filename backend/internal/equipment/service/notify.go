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
	tplBorrowApproved  = "equipment_borrow_approved"
	tplBorrowRejected  = "equipment_borrow_rejected"
	tplBorrowOverdue   = "equipment_borrow_overdue"
	tplMaintenanceStart = "equipment_maintenance_start"
)

type equipmentNotifierAdapter struct {
	inner notifsvc.NotificationService
}

func NewEquipmentNotifier(inner notifsvc.NotificationService) EquipmentNotifier {
	if inner == nil {
		return nil
	}
	return &equipmentNotifierAdapter{inner: inner}
}

func (a *equipmentNotifierAdapter) NotifyBorrowApproved(ctx context.Context, userID uuid.UUID, equipmentName string) error {
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{userID},
		TemplateCode: tplBorrowApproved,
		Variables: map[string]interface{}{
			"equipment_name": equipmentName,
		},
		Channels: []string{"in_app", "websocket"},
	})
}

func (a *equipmentNotifierAdapter) NotifyBorrowRejected(ctx context.Context, userID uuid.UUID, equipmentName, remark string) error {
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{userID},
		TemplateCode: tplBorrowRejected,
		Variables: map[string]interface{}{
			"equipment_name": equipmentName,
			"remark":         remark,
		},
		Channels: []string{"in_app", "websocket"},
	})
}

func (a *equipmentNotifierAdapter) NotifyOverdue(ctx context.Context, userID uuid.UUID, equipmentName string, expectedReturn time.Time) error {
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{userID},
		TemplateCode: tplBorrowOverdue,
		Variables: map[string]interface{}{
			"equipment_name":  equipmentName,
			"expected_return": expectedReturn.Format("2006-01-02 15:04"),
		},
		Channels: []string{"in_app", "websocket"},
	})
}

func (a *equipmentNotifierAdapter) NotifyMaintenanceStart(ctx context.Context, userID uuid.UUID, equipmentName string) error {
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{userID},
		TemplateCode: tplMaintenanceStart,
		Variables: map[string]interface{}{
			"equipment_name": equipmentName,
		},
		Channels: []string{"in_app", "websocket"},
	})
}

func (s *equipmentService) notifyBorrowApproved(ctx context.Context, userID uuid.UUID, equipmentName string) {
	if s.notifier == nil {
		return
	}
	if err := s.notifier.NotifyBorrowApproved(ctx, userID, equipmentName); err != nil {
		logger.Warn("send borrow approved notify failed", zap.Error(err))
	}
}
