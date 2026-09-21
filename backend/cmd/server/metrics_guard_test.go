package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
)

// TestMetricsGuard 固定 /metrics 的准入规则。
//
// 重点是「未设 METRICS_TOKEN 必须 fail-closed」这条：早期实现只看
// cfg.Server.Mode，于是用二进制直跑或裸 `docker run`（只补齐密钥环境变量、
// 不导出 APP_ENV）时，生效的是 base config.yaml 的 mode=debug，/metrics
// 就会连同全部 API 路由标签一起暴露在监听网卡上（Addr 是 ":%d"）。
func TestMetricsGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const token = "s3cret-token"

	tests := []struct {
		name       string
		tokenEnv   string // METRICS_TOKEN
		appEnv     string // APP_ENV
		mode       string // cfg.Server.Mode
		query      string // ?token=
		header     string // Authorization
		wantStatus int
	}{
		{name: "令牌正确（query）", tokenEnv: token, appEnv: "prod", mode: "release", query: token, wantStatus: http.StatusOK},
		{name: "令牌正确（Bearer）", tokenEnv: token, appEnv: "prod", mode: "release", header: "Bearer " + token, wantStatus: http.StatusOK},
		{name: "令牌错误", tokenEnv: token, appEnv: "prod", mode: "release", query: "wrong", wantStatus: http.StatusNotFound},
		{name: "未提供令牌", tokenEnv: token, appEnv: "prod", mode: "release", wantStatus: http.StatusNotFound},
		{name: "无令牌+APP_ENV未设+debug（裸跑二进制）", appEnv: "", mode: "debug", wantStatus: http.StatusNotFound},
		{name: "无令牌+prod+debug", appEnv: "prod", mode: "debug", wantStatus: http.StatusNotFound},
		{name: "无令牌+prod+release", appEnv: "prod", mode: "release", wantStatus: http.StatusNotFound},
		{name: "无令牌+dev+release（配置自相矛盾）", appEnv: "dev", mode: "release", wantStatus: http.StatusNotFound},
		{name: "无令牌+dev+debug（本地放行）", appEnv: "dev", mode: "debug", wantStatus: http.StatusOK},
		{name: "无令牌+test+debug（本地放行）", appEnv: "test", mode: "debug", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METRICS_TOKEN", tt.tokenEnv)
			t.Setenv("APP_ENV", tt.appEnv)

			cfg := &config.Config{Server: config.ServerConfig{Mode: tt.mode}}
			r := gin.New()
			r.GET("/metrics", metricsGuard(cfg), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			target := "/metrics"
			if tt.query != "" {
				target += "?token=" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, target, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("状态码 = %d，期望 %d", w.Code, tt.wantStatus)
			}
		})
	}
}
