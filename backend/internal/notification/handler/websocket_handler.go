package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	heartbeatInterval = 30 * time.Second
	writeTimeout      = 10 * time.Second
)

// WSMessage WebSocket 消息协议
type WSMessage struct {
	Type     string          `json:"type"`
	Token    string          `json:"token,omitempty"`
	Channels []string        `json:"channels,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// WSResponse WebSocket 响应消息
type WSResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// WSHandler WebSocket 连接处理器
type WSHandler struct {
	hub       service.HubManager
	jwtConfig *config.JWTConfig
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(hub service.HubManager, jwtConfig *config.JWTConfig) *WSHandler {
	return &WSHandler{
		hub:       hub,
		jwtConfig: jwtConfig,
	}
}

// HandleConnection GET /ws/notifications
func (h *WSHandler) HandleConnection(c *gin.Context) {
	// 1. 认证：优先 query param，其次 Authorization header
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, response.Response{
			Code:    response.CodeNotificationWSAuthFail,
			Message: "WebSocket 认证失败：缺少 Token",
		})
		return
	}

	// 2. 验证 JWT
	claims, err := authmiddleware.ParseToken(token, h.jwtConfig)
	if err != nil || claims == nil {
		c.JSON(http.StatusUnauthorized, response.Response{
			Code:    response.CodeNotificationWSAuthFail,
			Message: "WebSocket 认证失败：Token 无效或已过期",
		})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Response{
			Code:    response.CodeNotificationWSAuthFail,
			Message: "WebSocket 认证失败：无效的用户标识",
		})
		return
	}

	// 3. 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("websocket upgrade failed",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return
	}

	// 4. 注册连接
	h.hub.RegisterClient(userID, conn)

	// 5. 发送认证成功消息
	authResp := WSResponse{
		Type: "auth_result",
		Data: map[string]bool{"success": true},
	}
	_ = conn.WriteJSON(authResp)

	// 6. 启动读写协程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 读协程：处理客户端消息（心跳/订阅）
	go func() {
		defer cancel()
		h.readLoop(ctx, conn, userID)
		h.hub.UnregisterClient(userID, conn)
	}()

	// 写协程：心跳检测
	go h.writeLoop(ctx, conn, userID)

	// 等待读协程退出
	<-ctx.Done()
	if err := conn.Close(); err != nil {
		logger.Error("websocket conn close failed",
			zap.String("user_id", userID.String()),
			zap.Error(err))
	}
}

// readLoop 读取客户端消息
func (h *WSHandler) readLoop(ctx context.Context, conn *websocket.Conn, userID uuid.UUID) {
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logger.Error("websocket read error",
						zap.String("user_id", userID.String()),
						zap.Error(err))
				}
				return
			}

			switch msg.Type {
			case "ping":
				pong := WSResponse{Type: "pong"}
				_ = conn.WriteJSON(pong)
			case "subscribe":
				resp := WSResponse{
					Type: "subscribe_result",
					Data: map[string]bool{"success": true},
				}
				_ = conn.WriteJSON(resp)
			}
		}
	}
}

// writeLoop 心跳检测
func (h *WSHandler) writeLoop(ctx context.Context, conn *websocket.Conn, userID uuid.UUID) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
