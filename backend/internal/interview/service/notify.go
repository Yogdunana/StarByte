package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	tplInvite   = "interview_invite"
	tplAssigned = "interview_assigned"
	tplResult   = "interview_result"
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

func (s *interviewService) notifyInvite(ctx context.Context, applicant uuid.UUID, sess *model.Session, iv *model.Interview) {
	if s.notify == nil {
		return
	}
	u, _ := s.records.GetUser(ctx, applicant)
	vars := map[string]interface{}{
		"real_name":    displayName(u),
		"round":        fmt.Sprintf("第%d轮", sess.Round),
		"scheduled_at": formatTime(iv.ScheduledAt),
		"location":     firstNonEmpty(iv.Location, sess.Location),
	}
	if err := s.notify.Send(ctx, []uuid.UUID{applicant}, tplInvite, vars); err != nil {
		logger.Warn("send interview invite failed", zap.Error(err))
	}
}

func (s *interviewService) notifyAssigned(ctx context.Context, evaluators []uuid.UUID, iv *model.Interview) {
	if s.notify == nil || len(evaluators) == 0 {
		return
	}
	u, _ := s.records.GetUser(ctx, iv.ApplicantID)
	vars := map[string]interface{}{
		"real_name":      "",
		"applicant_name": displayName(u),
		"scheduled_at":   formatTime(iv.ScheduledAt),
		"location":       iv.Location,
		"round":          fmt.Sprintf("第%d轮", iv.Round),
	}
	if err := s.notify.Send(ctx, evaluators, tplAssigned, vars); err != nil {
		logger.Warn("send interview assigned failed", zap.Error(err))
	}
}

func (s *interviewService) notifyResult(ctx context.Context, iv *model.Interview) {
	if s.notify == nil {
		return
	}
	u, _ := s.records.GetUser(ctx, iv.ApplicantID)
	vars := map[string]interface{}{
		"real_name": displayName(u),
		"result":    resultLabel(iv.ResultCode),
		"comment":   iv.ResultComment,
	}
	if err := s.notify.Send(ctx, []uuid.UUID{iv.ApplicantID}, tplResult, vars); err != nil {
		logger.Warn("send interview result failed", zap.Error(err))
	}
}

func formatTime(t *time.Time) string {
	if t == nil {
		return "待定"
	}
	return t.Format("2006-01-02 15:04")
}
