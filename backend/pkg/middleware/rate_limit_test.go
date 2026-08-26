package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	return r
}

func TestRateLimit_NilRedis_FailsOpen(t *testing.T) {
	r := setupTestRouter()
	r.Use(RateLimit(nil, GlobalRateLimit))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitConfig_PredefinedConfigs(t *testing.T) {
	tests := []struct {
		name      string
		cfg       RateLimitConfig
		wantRate  int
		wantBurst int
	}{
		{
			name:      "GlobalRateLimit",
			cfg:       GlobalRateLimit,
			wantRate:  1000,
			wantBurst: 1200,
		},
		{
			name:      "PerIPRateLimit",
			cfg:       PerIPRateLimit,
			wantRate:  100,
			wantBurst: 120,
		},
		{
			name:      "LoginRateLimit",
			cfg:       LoginRateLimit,
			wantRate:  5,
			wantBurst: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantRate, tt.cfg.Rate)
			assert.Equal(t, tt.wantBurst, tt.cfg.Burst)
			assert.NotNil(t, tt.cfg.KeyFunc)
		})
	}
}

func TestGlobalRateLimit_KeyFunc(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	key := GlobalRateLimit.KeyFunc(c)
	assert.Equal(t, "ratelimit:global", key)
}

func TestPerIPRateLimit_KeyFunc(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = "192.168.1.100:12345"
	key := PerIPRateLimit.KeyFunc(c)
	assert.Contains(t, key, "ratelimit:ip:")
}

func TestLoginRateLimit_KeyFunc(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/auth/login", nil)
	c.Request.RemoteAddr = "10.0.0.1:54321"
	key := LoginRateLimit.KeyFunc(c)
	assert.Contains(t, key, "ratelimit:login:")
}

func TestRateLimitWithFallback_PanicRecovery(t *testing.T) {
	r := setupTestRouter()

	panicCfg := RateLimitConfig{
		Rate:    1,
		Burst:   1,
		Window:  1,
		KeyFunc: func(c *gin.Context) string { return "test" },
	}

	r.Use(RateLimitWithFallback(nil, panicCfg))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_ResponseHeaders(t *testing.T) {
	r := setupTestRouter()
	r.Use(func(c *gin.Context) {
		c.Header("X-RateLimit-Limit", "5")
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("Retry-After", "60")
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"code":       1006,
			"message":    "请求过于频繁，请 60 秒后重试",
			"data":       nil,
			"request_id": "test-request-id",
		})
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "60", w.Header().Get("Retry-After"))

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1006), resp["code"])
	assert.Contains(t, resp["message"].(string), "请求过于频繁")
}

func TestRateLimit_FailingOpenOnRedisError(t *testing.T) {
	r := setupTestRouter()
	r.Use(RateLimit(nil, GlobalRateLimit))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
