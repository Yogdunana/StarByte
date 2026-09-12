package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testJWT() *config.JWTConfig {
	return &config.JWTConfig{Secret: "monitor-ws-test-secret", Issuer: "starbyte", AccessTokenExp: 3600}
}

func TestCheckMonitorWSOrigin(t *testing.T) {
	allowed := map[string]bool{"https://starbyte.smbu.edu.cn": true}
	r := httptest.NewRequest(http.MethodGet, "/ws/monitor", nil)
	r.Host = "10.0.0.8"
	r.Header.Set("Origin", "http://10.0.0.8")
	assert.True(t, checkMonitorWSOrigin(r, allowed))

	r = httptest.NewRequest(http.MethodGet, "/ws/monitor", nil)
	r.Host = "10.0.0.8"
	r.Header.Set("Origin", "https://evil.example")
	assert.False(t, checkMonitorWSOrigin(r, allowed))

	r = httptest.NewRequest(http.MethodGet, "/ws/monitor", nil)
	assert.True(t, checkMonitorWSOrigin(r, map[string]bool{}))
}

func TestMonitorWSAuthFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewWSHandler(&stubSvc{app: &dto.AppHealth{}}, testJWT(), stubCache{perms: []string{"monitor:read"}}, nil)
	r := gin.New()
	RegisterWSRoute(r, h)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ws/monitor", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ws/monitor?token=bad", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	denied := NewWSHandler(&stubSvc{}, testJWT(), stubCache{perms: []string{"leave:read"}}, nil)
	token, _, err := auth.GenerateAccessToken(uuid.NewString(), "u", nil, nil, testJWT())
	require.NoError(t, err)
	rd := gin.New()
	RegisterWSRoute(rd, denied)
	w = httptest.NewRecorder()
	rd.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ws/monitor?token="+token, nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	var env response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, response.CodeMonitorWSForbidden, env.Code)
}

func TestMonitorWSPushAndPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uid := uuid.New()
	token, _, err := auth.GenerateAccessToken(uid.String(), "ops", nil, nil, testJWT())
	require.NoError(t, err)

	h := NewWSHandler(&stubSvc{
		server: &dto.ServerStatus{CPUPercent: 9},
		app:    &dto.AppHealth{Goroutines: 4},
	}, testJWT(), stubCache{perms: []string{"monitor:read"}}, nil)
	h.interval = 40 * time.Millisecond

	engine := gin.New()
	RegisterWSRoute(engine, h)
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/monitor?token=" + token
	conn, resp, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	t.Cleanup(func() { _ = conn.Close() })

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var authFrame WSResponse
	require.NoError(t, conn.ReadJSON(&authFrame))
	assert.Equal(t, "auth_result", authFrame.Type)

	var snapFrame WSResponse
	require.NoError(t, conn.ReadJSON(&snapFrame))
	assert.Equal(t, "snapshot", snapFrame.Type)

	require.NoError(t, conn.WriteJSON(WSMessage{Type: "ping"}))
	deadline := time.Now().Add(2 * time.Second)
	sawPong := false
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		var frame WSResponse
		if err := conn.ReadJSON(&frame); err != nil {
			continue
		}
		if frame.Type == "pong" {
			sawPong = true
			break
		}
	}
	assert.True(t, sawPong)
}

func TestExtractWSTokenBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ws/monitor", nil)
	c.Request.Header.Set("Authorization", "Bearer abc.def")
	assert.Equal(t, "abc.def", extractWSToken(c))
}

func TestAllowMonitorReadSuperAdmin(t *testing.T) {
	h := NewWSHandler(&stubSvc{}, testJWT(), stubCache{super: true}, nil)
	require.NoError(t, h.allowMonitorRead(context.Background(), uuid.New()))
	h.cache = nil
	err := h.allowMonitorRead(context.Background(), uuid.New())
	require.Error(t, err)
}
