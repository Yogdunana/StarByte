package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, rdb
}

func TestStore_AllowTokenBucket(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewStore(rdb)
	s.now = func() time.Time { return time.Unix(1_700_000_000, 0) }
	b := Bucket{Rate: 1, Burst: 2}
	r1, err := s.Allow(context.Background(), "k", b)
	require.NoError(t, err)
	assert.True(t, r1.Allowed)
	r2, err := s.Allow(context.Background(), "k", b)
	require.NoError(t, err)
	assert.True(t, r2.Allowed)
	r3, err := s.Allow(context.Background(), "k", b)
	require.NoError(t, err)
	assert.False(t, r3.Allowed)
	assert.GreaterOrEqual(t, r3.RetryAfter, 1)
}

func TestMiddleware_IPLimitAndHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.IP = Bucket{Rate: 1, Burst: 1}
	cfg.Route = Bucket{Rate: 100, Burst: 100}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(rdb, cfg))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "ok") })

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, 200, w1.Code)
	assert.NotEmpty(t, w1.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "prod", w1.Header().Get(HeaderTrafficColor))

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.NotEmpty(t, w2.Header().Get("Retry-After"))
	assert.Contains(t, w2.Body.String(), "21001")
}

func TestMiddleware_BlacklistAndWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.IPBlacklist = map[string]struct{}{"10.0.0.9": {}}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(rdb, cfg))
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "10.0.0.9:1"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "21003")

	cfg2 := DefaultConfig()
	cfg2.IP = Bucket{Rate: 0.0001, Burst: 1}
	cfg2.IPWhitelist = map[string]struct{}{"192.0.2.1": {}}
	r2 := gin.New()
	r2.Use(Middleware(rdb, cfg2))
	r2.GET("/x", func(c *gin.Context) { c.String(200, "ok") })
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r2.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		assert.Equal(t, 200, w.Code)
	}
}

func TestUserMiddleware_RequiresUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.User = Bucket{Rate: 1, Burst: 1}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(UserMiddleware(rdb, cfg))
	r.GET("/me", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	assert.Equal(t, 200, w.Code)

	r2 := gin.New()
	r2.Use(func(c *gin.Context) {
		c.Set("request_id", "rid")
		c.Set("user_id", "u1")
		c.Next()
	})
	r2.Use(UserMiddleware(rdb, cfg))
	r2.GET("/me", func(c *gin.Context) { c.String(200, "ok") })
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	assert.Equal(t, 200, w.Code)
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestUserMiddleware_Blacklist(t *testing.T) {
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.UserBlacklist = map[string]struct{}{"u-bad": {}}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("request_id", "rid")
		c.Set("user_id", "u-bad")
		c.Next()
	})
	r.Use(UserMiddleware(rdb, cfg))
	r.GET("/me", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "21003")
}

func TestUserMiddleware_BlacklistBeatsIPWhitelist(t *testing.T) {
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.IPWhitelist = map[string]struct{}{"192.0.2.1": {}}
	cfg.UserBlacklist = map[string]struct{}{"u-bad": {}}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("request_id", "rid")
		c.Set("user_id", "u-bad")
		c.Next()
	})
	r.Use(UserMiddleware(rdb, cfg))
	r.GET("/me", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "21003")
}

func TestMiddleware_CIDRBlacklist(t *testing.T) {
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.IPBlacklist = map[string]struct{}{"10.0.0.0/8": {}}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(rdb, cfg))
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "10.8.0.2:9"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_NilRedisStillLimitsLocally(t *testing.T) {
	cfg := DefaultConfig()
	cfg.IP = Bucket{Rate: 1, Burst: 1}
	cfg.Route = Bucket{Rate: 100, Burst: 100}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(nil, cfg))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "ok") })
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, 200, w1.Code)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestGrayPercent(t *testing.T) {
	assert.False(t, grayByPercent("x", 0))
	assert.True(t, grayByPercent("x", 100))
	hit := 0
	for i := 0; i < 100; i++ {
		if grayByPercent(string(rune('a'+i%26))+string(rune('0'+i%10)), 30) {
			hit++
		}
	}
	assert.Greater(t, hit, 5)
	assert.Less(t, hit, 70)
}
