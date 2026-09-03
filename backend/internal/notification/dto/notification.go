package dto

import (
	"time"

	"github.com/google/uuid"
)

// ========== Notification DTOs ==========

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Category  string     `json:"category"`
	Priority  string     `json:"priority"`
	IsRead    bool       `json:"is_read"`
	ActionURL string     `json:"action_url"`
	Sender    SenderInfo `json:"sender"`
	CreatedAt time.Time  `json:"created_at"`
}

// SenderInfo 发送者信息
type SenderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListNotificationsRequest 通知列表请求
type ListNotificationsRequest struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	Category   string `form:"category" json:"category"`
	UnreadOnly bool   `form:"unread_only" json:"unread_only"`
}

// UnreadCountResponse 未读计数响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// MarkAllReadRequest 全部已读请求
type MarkAllReadRequest struct {
	Category string `json:"category"` // 可选：按分类标记已读
}

// ========== Notification Template DTOs ==========

// NotificationTemplateResponse 模板响应
type NotificationTemplateResponse struct {
	ID              string            `json:"id"`
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	TitleTemplate   string            `json:"title_template"`
	BodyTemplate    string            `json:"body_template"`
	Channels        []string          `json:"channels"`
	Category        string            `json:"category"`
	VariablesSchema map[string]string `json:"variables_schema"`
	Status          int               `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Code            string            `json:"code" binding:"required"`
	Name            string            `json:"name" binding:"required"`
	TitleTemplate   string            `json:"title_template" binding:"required"`
	BodyTemplate    string            `json:"body_template" binding:"required"`
	Channels        []string          `json:"channels" binding:"required"`
	Category        string            `json:"category"`
	VariablesSchema map[string]string `json:"variables_schema"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name            *string           `json:"name"`
	TitleTemplate   *string           `json:"title_template"`
	BodyTemplate    *string           `json:"body_template"`
	Channels        []string          `json:"channels"`
	Category        *string           `json:"category"`
	VariablesSchema map[string]string `json:"variables_schema"`
	Status          *int              `json:"status"`
}

// TestTemplateRequest 测试模板请求
type TestTemplateRequest struct {
	Variables map[string]interface{} `json:"variables"`
}

// TestTemplateResponse 测试模板响应
type TestTemplateResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ========== Admin Send Notification DTOs ==========

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	UserIDs      []uuid.UUID            `json:"user_ids" binding:"required,min=1"`
	TemplateCode string                 `json:"template_code" binding:"required"`
	Variables    map[string]interface{} `json:"variables"`
	Channels     []string               `json:"channels"` // 可选：覆盖模板渠道
}

// BroadcastNotificationRequest 广播通知请求
type BroadcastNotificationRequest struct {
	Title    string   `json:"title" binding:"required"`
	Content  string   `json:"content" binding:"required"`
	Category string   `json:"category"`
	Priority string   `json:"priority"`
	Channels []string `json:"channels"`
}

// ListTemplatesRequest 模板列表请求
type ListTemplatesRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Keyword  string `form:"keyword" json:"keyword"`
}
