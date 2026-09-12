package handler

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	monitorWriteTimeout = 10 * time.Second
	monitorReadIdle     = 60 * time.Second
	defaultPushInterval = 5 * time.Second
)

// WSMessage is a client frame on /ws/monitor.
type WSMessage struct {
	Type string `json:"type"`
}

// WSResponse is a server frame on /ws/monitor.
type WSResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// WSHandler pushes monitor snapshots after JWT + monitor:read.
type WSHandler struct {
	svc      service.Service
	jwt      *config.JWTConfig
	cache    rbacService.PermissionCacheService
	upgrader websocket.Upgrader
	interval time.Duration
}

// NewWSHandler builds /ws/monitor. Empty allowedOrigins permits any Origin (dev).
func NewWSHandler(svc service.Service, jwt *config.JWTConfig, cache rbacService.PermissionCacheService, allowedOrigins []string) *WSHandler {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		originSet[origin] = true
	}
	return &WSHandler{
		svc:   svc,
		jwt:   jwt,
		cache: cache,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return checkMonitorWSOrigin(r, originSet)
			},
		},
		interval: defaultPushInterval,
	}
}

// RegisterWSRoute mounts GET /ws/monitor on the engine (auth is inside the handler).
func RegisterWSRoute(r gin.IRouter, h *WSHandler) {
	r.GET("/ws/monitor", h.HandleConnection)
}

func checkMonitorWSOrigin(r *http.Request, allowed map[string]bool) bool {
	if len(allowed) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if allowed[origin] {
		return true
	}
	return originMatchesRequestHost(origin, r.Host)
}

func originMatchesRequestHost(origin, reqHost string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.EqualFold(stripHostPort(u.Host), stripHostPort(reqHost))
}

func stripHostPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host[1 : len(host)-1]
	}
	return host
}

func extractWSToken(c *gin.Context) string {
	if token := c.Query("token"); token != "" {
		return token
	}
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return authHeader[7:]
	}
	return ""
}

func wsAuthFail(msg string) *response.AppError {
	return &response.AppError{
		Code:       response.CodeMonitorWSAuthFail,
		Message:    msg,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func (h *WSHandler) allowMonitorRead(ctx context.Context, userID uuid.UUID) error {
	if h.cache == nil {
		return &response.AppError{
			Code:       response.CodeMonitorWSForbidden,
			Message:    "权限缓存不可用",
			HTTPStatus: http.StatusForbidden,
		}
	}
	perms, isSuper, err := h.cache.GetUserPermissionsAndSuperAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if isSuper {
		return nil
	}
	for _, p := range perms {
		if p == "monitor:read" || p == "*" {
			return nil
		}
	}
	return &response.AppError{
		Code:       response.CodeMonitorWSForbidden,
		Message:    "权限不足: monitor:read",
		HTTPStatus: http.StatusForbidden,
	}
}

// HandleConnection GET /ws/monitor
func (h *WSHandler) HandleConnection(c *gin.Context) {
	token := extractWSToken(c)
	if token == "" {
		response.Error(c, wsAuthFail("WebSocket 认证失败：缺少 Token"))
		return
	}
	if h.jwt == nil {
		response.Error(c, wsAuthFail("WebSocket 认证失败：未配置签发密钥"))
		return
	}
	claims, err := authmiddleware.ParseToken(token, h.jwt)
	if err != nil || claims == nil {
		response.Error(c, wsAuthFail("WebSocket 认证失败：Token 无效或已过期"))
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.Error(c, wsAuthFail("WebSocket 认证失败：无效的用户标识"))
		return
	}
	if err := h.allowMonitorRead(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("monitor websocket upgrade failed",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := &safeWS{conn: conn}
	if err := writer.writeJSON(WSResponse{Type: "auth_result", Data: map[string]bool{"success": true}}); err != nil {
		_ = conn.Close()
		return
	}

	go func() {
		defer cancel()
		h.readLoop(ctx, conn, writer)
	}()
	go h.pushLoop(ctx, writer)

	<-ctx.Done()
	if err := conn.Close(); err != nil {
		logger.Error("monitor websocket close failed",
			zap.String("user_id", userID.String()),
			zap.Error(err))
	}
}

type safeWS struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (s *safeWS) writeJSON(v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.conn.SetWriteDeadline(time.Now().Add(monitorWriteTimeout))
	return s.conn.WriteJSON(v)
}

func (h *WSHandler) readLoop(ctx context.Context, conn *websocket.Conn, writer *safeWS) {
	_ = conn.SetReadDeadline(time.Now().Add(monitorReadIdle))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(monitorReadIdle))
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
					logger.Error("monitor websocket read error", zap.Error(err))
				}
				return
			}
			_ = conn.SetReadDeadline(time.Now().Add(monitorReadIdle))
			if msg.Type == "ping" {
				_ = writer.writeJSON(WSResponse{Type: "pong"})
			}
		}
	}
}

func (h *WSHandler) pushLoop(ctx context.Context, writer *safeWS) {
	interval := h.interval
	if interval <= 0 {
		interval = defaultPushInterval
	}
	h.pushSnapshot(ctx, writer)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.pushSnapshot(ctx, writer)
		}
	}
}

func (h *WSHandler) pushSnapshot(ctx context.Context, writer *safeWS) {
	if h.svc == nil {
		return
	}
	snap := h.svc.Snapshot(ctx)
	if err := writer.writeJSON(WSResponse{Type: "snapshot", Data: snap}); err != nil {
		logger.Error("monitor websocket push failed", zap.Error(err))
	}
}
