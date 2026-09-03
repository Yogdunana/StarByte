package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EventListener 事件总线监听器，监听业务事件并自动发送通知
type EventListener struct {
	templateRepo    repo.NotificationTemplateRepo
	templateEngine  TemplateEngine
	channelRegistry *ChannelRegistry
}

// NewEventListener 创建事件监听器
func NewEventListener(
	templateRepo repo.NotificationTemplateRepo,
	templateEngine TemplateEngine,
	channelRegistry *ChannelRegistry,
) *EventListener {
	return &EventListener{
		templateRepo:    templateRepo,
		templateEngine:  templateEngine,
		channelRegistry: channelRegistry,
	}
}

// RegisterAll 注册所有事件监听
func (l *EventListener) RegisterAll(eventBus *events.EventBus) {
	eventBus.Subscribe("task.created", l.onTaskCreated)
	eventBus.Subscribe("task.assigned", l.onTaskAssigned)
}

// onTaskCreated 处理流程任务创建事件
func (l *EventListener) onTaskCreated(ctx context.Context, event events.Event) error {
	taskEvent, ok := event.(events.TaskCreatedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task.created")
	}

	if taskEvent.AssigneeID == uuid.Nil {
		return nil
	}

	templateCode := "FLOW_TASK_CREATED"
	variables := map[string]interface{}{
		"task_name": taskEvent.NodeName,
		"node_name": taskEvent.NodeName,
		"task_type": taskEvent.TaskType,
	}

	return l.sendNotification(ctx, taskEvent.AssigneeID, templateCode, variables, "task")
}

// onTaskAssigned 处理任务转办事件
func (l *EventListener) onTaskAssigned(ctx context.Context, event events.Event) error {
	taskEvent, ok := event.(events.TaskAssignedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task.assigned")
	}

	if taskEvent.NewAssigneeID == uuid.Nil {
		return nil
	}

	templateCode := "TASK_ASSIGNED"
	variables := map[string]interface{}{
		"task_id":     taskEvent.TaskID.String(),
		"instance_id": taskEvent.InstanceID.String(),
	}

	return l.sendNotification(ctx, taskEvent.NewAssigneeID, templateCode, variables, "task")
}

// sendNotification 渲染模板并通过渠道发送通知
func (l *EventListener) sendNotification(ctx context.Context, userID uuid.UUID, templateCode string, variables map[string]interface{}, category string) error {
	// 查询模板一次，同时用于渲染和获取渠道配置
	tpl, err := l.templateRepo.GetByCode(ctx, templateCode)

	var rendered *dto.TestTemplateResponse
	channels := []string{"in_app", "websocket"}

	if err != nil {
		// 模板不存在，使用默认内容
		logger.Warn("template not found, using fallback",
			zap.String("template_code", templateCode),
			zap.Error(err))
		rendered = &dto.TestTemplateResponse{
			Title:   "新通知",
			Content: fmt.Sprintf("您有一条新的 %s 通知", category),
		}
	} else {
		// 从已查出的模板渲染，避免重复查库
		rendered, err = l.templateEngine.RenderTemplate(tpl, variables)
		if err != nil {
			logger.Warn("template render failed, using fallback",
				zap.String("template_code", templateCode),
				zap.Error(err))
			rendered = &dto.TestTemplateResponse{
				Title:   "新通知",
				Content: fmt.Sprintf("您有一条新的 %s 通知", category),
			}
		}
		// 使用模板配置的渠道
		if tpl.Channels != "" {
			channels = tpl.GetChannels()
		}
	}

	// 通过渠道发送通知（in_app 渠道会自动创建站内通知，避免重复创建）
	msg := &NotificationMessage{
		UserID:   userID,
		Title:    rendered.Title,
		Content:  rendered.Content,
		Category: category,
		Priority: "normal",
	}
	if errs := l.channelRegistry.SendViaChannels(ctx, msg, channels); len(errs) > 0 {
		logger.Error("send notification via channels failed",
			zap.String("template_code", templateCode),
			zap.String("user_id", userID.String()),
			zap.Int("error_count", len(errs)))
	}

	return nil
}
