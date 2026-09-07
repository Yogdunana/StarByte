package service

import (
	"context"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	wfsvc "github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

type notificationAdapter struct{ inner notifsvc.NotificationService }

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
		UserIDs: userIDs, TemplateCode: template, Variables: vars, Channels: []string{"in_app", "websocket"},
	})
}

type FlowStarter interface {
	Start(ctx context.Context, recordID, initiator uuid.UUID, vars map[string]interface{}) (*uuid.UUID, error)
}

type definitionLookup interface {
	GetByKey(ctx context.Context, key string) (*wfmodel.FlowDefinition, error)
}

type engineStarter struct {
	defs definitionLookup
	inst wfsvc.InstanceService
}

func NewFlowStarter(defs definitionLookup, inst wfsvc.InstanceService) FlowStarter {
	if defs == nil || inst == nil {
		return nil
	}
	return &engineStarter{defs: defs, inst: inst}
}

func (s *engineStarter) Start(ctx context.Context, recordID, initiator uuid.UUID, vars map[string]interface{}) (*uuid.UUID, error) {
	def, err := s.defs.GetByKey(ctx, "discipline_approve")
	if err != nil {
		logger.Warn("discipline flow lookup failed", zap.Error(err))
		return nil, nil
	}
	if def == nil || def.Status != 1 {
		return nil, nil
	}
	inst, err := s.inst.Start(ctx, def.ID, recordID.String(), "discipline", initiator, vars)
	if err != nil {
		logger.Warn("discipline flow start failed", zap.Error(err))
		return nil, nil
	}
	return &inst.ID, nil
}
