package service

import (
	"context"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	tplTaskAssigned    = "task_assigned"
	tplTaskTransferred = "task_transferred"
	tplTaskUrged       = "task_urged"
	tplTaskMention     = "task_mention"
	tplTaskDueSoon     = "task_due_soon"
	tplTaskOverdue     = "task_overdue"
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

func (s *taskService) notifyUsers(ctx context.Context, userIDs []uuid.UUID, template string, t *model.Task, extra string) {
	if s.notify == nil || len(userIDs) == 0 || t == nil {
		return
	}
	due := ""
	if t.DueDate != nil {
		due = t.DueDate.Format("2006-01-02 15:04")
	}
	for _, uid := range userIDs {
		name := ""
		if u, err := s.tasks.GetUser(ctx, uid); err == nil && u != nil {
			name = u.RealName
			if name == "" {
				name = u.Username
			}
		}
		vars := map[string]interface{}{
			"title":     t.Title,
			"real_name": name,
			"message":   extra,
			"due_date":  due,
		}
		if err := s.notify.Send(ctx, []uuid.UUID{uid}, template, vars); err != nil {
			logger.Warn("send task notify failed", zap.Error(err), zap.String("tpl", template))
		}
	}
}
