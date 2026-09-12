package service

import (
	"context"
	"fmt"
	"strings"

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
		"task_name":      taskEvent.NodeName,
		"node_name":      taskEvent.NodeName,
		"task_type":      taskEvent.TaskType,
		"task_id":        taskEvent.TaskID.String(),
		"instance_id":    taskEvent.InstanceID.String(),
		"business_type":  taskEvent.BusinessType,
		"business_key":   taskEvent.BusinessKey,
		"applicant_name": taskEvent.ApplicantName,
	}

	category := "approval"
	if taskEvent.BusinessType == "member_application" {
		category = "member"
	}
	return l.sendNotification(ctx, taskEvent.AssigneeID, templateCode, variables, category, workflowActionURL(taskEvent.TaskID), workflowFallback(taskEvent))
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

	return l.sendNotification(ctx, taskEvent.NewAssigneeID, templateCode, variables, "approval", workflowActionURL(taskEvent.TaskID), &dto.TestTemplateResponse{Title: "审批待办已转交给你", Content: "请打开待办详情查看流程事项与处理要求。"})
}

// sendNotification 渲染模板并通过渠道发送通知
func (l *EventListener) sendNotification(ctx context.Context, userID uuid.UUID, templateCode string, variables map[string]interface{}, category, actionURL string, fallback *dto.TestTemplateResponse) error {
	// 查询模板一次，同时用于渲染和获取渠道配置
	tpl, err := l.templateRepo.GetByCode(ctx, templateCode)

	var rendered *dto.TestTemplateResponse
	channels := []string{"in_app", "websocket"}

	if err != nil || tpl == nil {
		// 模板不存在，使用默认内容
		logger.Warn("template not found, using fallback",
			zap.String("template_code", templateCode),
			zap.Error(err))
		rendered = fallback
	} else {
		// 从已查出的模板渲染，避免重复查库
		rendered, err = l.templateEngine.RenderTemplate(tpl, variables)
		if err != nil {
			logger.Warn("template render failed, using fallback",
				zap.String("template_code", templateCode),
				zap.Error(err))
			rendered = fallback
		}
		// 使用模板配置的渠道
		if tpl.Channels != "" {
			channels = tpl.GetChannels()
		}
	}

	// 通过渠道发送通知（in_app 渠道会自动创建站内通知，避免重复创建）
	msg := &NotificationMessage{
		UserID:     userID,
		Title:      rendered.Title,
		Content:    rendered.Content,
		Category:   category,
		Priority:   "normal",
		ActionURL:  actionURL,
		SenderName: "工作流",
	}
	if errs := l.channelRegistry.SendViaChannels(ctx, msg, channels); len(errs) > 0 {
		logger.Error("send notification via channels failed",
			zap.String("template_code", templateCode),
			zap.String("user_id", userID.String()),
			zap.Int("error_count", len(errs)))
	}

	return nil
}

func workflowActionURL(taskID uuid.UUID) string {
	if taskID == uuid.Nil {
		return "/workflow/todo"
	}
	return "/workflow/todo?task_id=" + taskID.String()
}

func workflowFallback(event events.TaskCreatedEvent) *dto.TestTemplateResponse {
	node := strings.TrimSpace(event.NodeName)
	if node == "" {
		node = "流程审批"
	}
	if runes := []rune(node); len(runes) > 100 {
		node = string(runes[:100])
	}
	business := "审批流程"
	switch event.BusinessType {
	case "member_application":
		business = "入会申请"
	case "collaboration_task":
		business = "协作任务"
	case "task_transfer":
		business = "任务交接"
	case "leave_application":
		business = "请假申请"
	}
	content := fmt.Sprintf("事项：%s\n当前环节：%s", business, node)
	if event.ApplicantName != "" {
		content += "\n申请人：" + event.ApplicantName
	}
	if event.BusinessKey != "" {
		content += "\n事项编号：" + event.BusinessKey
	}
	content += "\n请打开待办详情查看资料并处理。"
	return &dto.TestTemplateResponse{Title: "待办：" + business + " · " + node, Content: content}
}
