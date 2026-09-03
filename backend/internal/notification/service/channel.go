package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/go-mail/mail"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// NotificationChannel 通知渠道接口（插件化，未来可扩展短信、企业微信等）
type NotificationChannel interface {
	Type() string // "in_app", "email", "websocket"
	Send(ctx context.Context, msg *NotificationMessage) error
	IsAvailable() bool
}

// NotificationMessage 通知消息体
type NotificationMessage struct {
	UserID     uuid.UUID
	Username   string
	Email      string
	Title      string
	Content    string
	Category   string
	Priority   string
	ActionURL  string
	SenderID   *uuid.UUID
	SenderName string
}

// InAppChannel 站内消息渠道
type InAppChannel struct {
	notificationRepo repo.NotificationRepo
}

// NewInAppChannel 创建站内消息渠道
func NewInAppChannel(repo repo.NotificationRepo) *InAppChannel {
	return &InAppChannel{notificationRepo: repo}
}

func (c *InAppChannel) Type() string      { return "in_app" }
func (c *InAppChannel) IsAvailable() bool { return true }

func (c *InAppChannel) Send(ctx context.Context, msg *NotificationMessage) error {
	n := &model.Notification{
		ID:         uuid.New(),
		UserID:     msg.UserID,
		Title:      msg.Title,
		Content:    msg.Content,
		Category:   msg.Category,
		Priority:   msg.Priority,
		ActionURL:  msg.ActionURL,
		SenderID:   msg.SenderID,
		SenderName: msg.SenderName,
		IsRead:     false,
	}
	return c.notificationRepo.Create(ctx, n)
}

// EmailChannel 邮件渠道
type EmailChannel struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
}

// NewEmailChannel 创建邮件渠道
func NewEmailChannel(host string, port int, username, password, from string) *EmailChannel {
	return &EmailChannel{
		smtpHost: host,
		smtpPort: port,
		username: username,
		password: password,
		from:     from,
	}
}

func (c *EmailChannel) Type() string { return "email" }

func (c *EmailChannel) IsAvailable() bool {
	return c.smtpHost != "" && c.smtpPort > 0 && c.from != ""
}

func (c *EmailChannel) Send(ctx context.Context, msg *NotificationMessage) error {
	if msg.Email == "" {
		return fmt.Errorf("email address is empty for user %s", msg.UserID)
	}

	m := mail.NewMessage()
	m.SetHeader("From", c.from)
	m.SetHeader("To", msg.Email)
	m.SetHeader("Subject", msg.Title)
	m.SetBody("text/plain", msg.Content)

	d := mail.NewDialer(c.smtpHost, c.smtpPort, c.username, c.password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

// WebSocketChannel WebSocket 实时推送渠道
type WebSocketChannel struct {
	hub HubManager
}

// NewWebSocketChannel 创建 WebSocket 渠道
func NewWebSocketChannel(hub HubManager) *WebSocketChannel {
	return &WebSocketChannel{hub: hub}
}

func (c *WebSocketChannel) Type() string      { return "websocket" }
func (c *WebSocketChannel) IsAvailable() bool { return c.hub != nil }

func (c *WebSocketChannel) Send(ctx context.Context, msg *NotificationMessage) error {
	notification := map[string]interface{}{
		"type": "notification",
		"data": map[string]interface{}{
			"title":      msg.Title,
			"content":    msg.Content,
			"category":   msg.Category,
			"priority":   msg.Priority,
			"action_url": msg.ActionURL,
			"sender": map[string]string{
				"id":   "",
				"name": msg.SenderName,
			},
		},
	}

	// Hub 的 PushToUser 会处理 JSON 序列化
	return c.hub.PushToUser(msg.UserID, notification)
}

// ChannelRegistry 渠道注册表
type ChannelRegistry struct {
	channels map[string]NotificationChannel
}

// NewChannelRegistry 创建渠道注册表
func NewChannelRegistry() *ChannelRegistry {
	return &ChannelRegistry{
		channels: make(map[string]NotificationChannel),
	}
}

// Register 注册渠道
func (r *ChannelRegistry) Register(channel NotificationChannel) {
	r.channels[channel.Type()] = channel
}

// Get 获取渠道
func (r *ChannelRegistry) Get(channelType string) (NotificationChannel, bool) {
	ch, ok := r.channels[channelType]
	return ch, ok
}

// SendViaChannels 通过指定渠道发送通知
func (r *ChannelRegistry) SendViaChannels(ctx context.Context, msg *NotificationMessage, channels []string) []error {
	var errs []error
	for _, chType := range channels {
		ch, ok := r.Get(chType)
		if !ok {
			logger.Warn("notification channel not found",
				zap.String("channel", chType))
			errs = append(errs, fmt.Errorf("unsupported channel: %s", chType))
			continue
		}
		if !ch.IsAvailable() {
			logger.Warn("notification channel not available",
				zap.String("channel", chType))
			continue
		}
		if err := ch.Send(ctx, msg); err != nil {
			logger.Error("send notification via channel failed",
				zap.String("channel", chType),
				zap.String("user_id", msg.UserID.String()),
				zap.Error(err))
			errs = append(errs, err)
		}
	}
	return errs
}

// HubManager WebSocket 连接管理器接口
type HubManager interface {
	RegisterClient(userID uuid.UUID, conn *websocket.Conn)
	UnregisterClient(userID uuid.UUID, conn *websocket.Conn)
	PushToUser(userID uuid.UUID, message interface{}) error
	PushToChannel(channel string, message interface{}) error
	GetOnlineUsers() []uuid.UUID
}

// Hub WebSocket Hub 实现
type Hub struct {
	mu      sync.RWMutex
	writeMu sync.Mutex // 序列化 WebSocket 写入，防止并发写同一连接
	clients map[uuid.UUID]map[*websocket.Conn]bool // 一个用户可能有多个连接
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*websocket.Conn]bool),
	}
}

// RegisterClient 注册客户端连接
func (h *Hub) RegisterClient(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[userID]; !ok {
		h.clients[userID] = make(map[*websocket.Conn]bool)
	}
	h.clients[userID][conn] = true
	logger.Info("websocket client registered",
		zap.String("user_id", userID.String()),
		zap.Int("connections", len(h.clients[userID])))
}

// UnregisterClient 注销客户端连接
func (h *Hub) UnregisterClient(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[userID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, userID)
		}
		logger.Info("websocket client unregistered",
			zap.String("user_id", userID.String()))
	}
}

// PushToUser 向指定用户推送消息（所有连接）
func (h *Hub) PushToUser(userID uuid.UUID, message interface{}) error {
	h.mu.RLock()
	conns, ok := h.clients[userID]
	if !ok || len(conns) == 0 {
		h.mu.RUnlock()
		return nil // 用户不在线，静默跳过
	}
	// 复制连接列表，避免在 I/O 期间持有锁
	connList := make([]*websocket.Conn, 0, len(conns))
	for conn := range conns {
		connList = append(connList, conn)
	}
	h.mu.RUnlock()

	// 序列化写入：gorilla/websocket 不支持并发写同一连接
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	for _, conn := range connList {
		if err := conn.WriteJSON(message); err != nil {
			logger.Error("websocket push failed",
				zap.String("user_id", userID.String()),
				zap.Error(err))
		}
	}
	return nil
}

// PushToChannel 向频道推送消息（目前不支持，预留）
func (h *Hub) PushToChannel(channel string, message interface{}) error {
	// 预留：未来实现频道订阅功能
	return nil
}

// GetOnlineUsers 获取在线用户列表
func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()
	users := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}
