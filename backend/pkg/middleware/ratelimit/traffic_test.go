package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestStore_NilRedisUsesLocal(t *testing.T) {
	s := NewStore(nil)
	b := Bucket{Rate: 1, Burst: 1}
	r1, err := s.Allow(context.Background(), "local", b)
	require.NoError(t, err)
	assert.True(t, r1.Allowed)
	r2, err := s.Allow(context.Background(), "local", b)
	require.NoError(t, err)
	assert.False(t, r2.Allowed)
	assert.GreaterOrEqual(t, r2.RetryAfter, 1)
}

func TestStore_ZeroRateAllows(t *testing.T) {
	s := NewStore(nil)
	got, err := s.Allow(context.Background(), "k", Bucket{Rate: 0, Burst: 1})
	require.NoError(t, err)
	assert.True(t, got.Allowed)
}

func TestStore_RedisDownFallsBackToLocal(t *testing.T) {
	mr, rdb := testRedis(t)
	s := NewStore(rdb)
	mr.Close()
	b := Bucket{Rate: 1, Burst: 1}
	r1, err := s.Allow(context.Background(), "fb", b)
	require.NoError(t, err)
	assert.True(t, r1.Allowed)
	r2, err := s.Allow(context.Background(), "fb", b)
	require.NoError(t, err)
	assert.False(t, r2.Allowed)
}

func TestNilStoreAllows(t *testing.T) {
	var s *Store
	got, err := s.Allow(context.Background(), "k", Bucket{Rate: 1, Burst: 1})
	require.NoError(t, err)
	assert.True(t, got.Allowed)
}

func TestIPListedCIDR(t *testing.T) {
	set := map[string]struct{}{"10.0.0.0/8": {}}
	assert.True(t, ipListed(set, "10.1.2.3"))
	assert.False(t, ipListed(set, "11.0.0.1"))
	assert.True(t, ipListed(map[string]struct{}{"127.0.0.1": {}}, "127.0.0.1"))
	assert.False(t, ipListed(nil, "127.0.0.1"))
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("TRAFFIC_IP_BLACKLIST", "10.0.0.1, 10.0.0.0/8")
	t.Setenv("TRAFFIC_IP_WHITELIST", "192.0.2.1")
	t.Setenv("TRAFFIC_USER_BLACKLIST", "u-bad")
	t.Setenv("TRAFFIC_USER_WHITELIST", "u-ok")
	t.Setenv("TRAFFIC_GRAY_GROUPS", "canary")
	t.Setenv("TRAFFIC_GRAY_PERCENT", "15")
	t.Setenv("TRAFFIC_IP_RATE", "3")
	t.Setenv("TRAFFIC_IP_BURST", "9")
	cfg := LoadFromEnv()
	assert.True(t, ipListed(cfg.IPBlacklist, "10.9.9.9"))
	assert.True(t, inSet(cfg.IPWhitelist, "192.0.2.1"))
	assert.True(t, inSet(cfg.UserBlacklist, "u-bad"))
	assert.True(t, inSet(cfg.UserWhitelist, "u-ok"))
	assert.True(t, inSet(cfg.GrayAllow, "canary"))
	assert.Equal(t, 15, cfg.GrayPercent)
	assert.Equal(t, 3.0, cfg.IP.Rate)
	assert.Equal(t, 9.0, cfg.IP.Burst)
}

func TestPaintTrafficGrayHeader(t *testing.T) {
	cfg := DefaultConfig()
	cfg.GrayAllow = map[string]struct{}{"canary": {}}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c.Request.Header.Set("X-Gray-Group", "canary")
	paintTraffic(c, cfg)
	assert.Equal(t, "gray", c.GetString(ContextTrafficColor))
	assert.Equal(t, "gray", w.Header().Get(HeaderTrafficColor))
}

func TestPaintTrafficIncomingColor(t *testing.T) {
	cfg := DefaultConfig()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c.Request.Header.Set(HeaderTrafficColor, "gray")
	paintTraffic(c, cfg)
	assert.Equal(t, "gray", c.GetString(ContextTrafficColor))
}

func TestRouteKey(t *testing.T) {
	assert.Equal(t, "unknown", routeKey(nil))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ping", nil)
	assert.Equal(t, "POST:", routeKey(c))
}

func TestViewerIDEmpty(t *testing.T) {
	assert.Equal(t, "", viewerID(nil))
}

func TestParseSetEmpty(t *testing.T) {
	assert.Nil(t, parseSet("  ,  "))
	assert.Equal(t, 2, len(parseSet("a, b")))
}

func TestInSet(t *testing.T) {
	assert.False(t, inSet(nil, "x"))
	assert.False(t, inSet(map[string]struct{}{"a": {}}, ""))
	assert.True(t, inSet(map[string]struct{}{"a": {}}, "a"))
}

func TestClientIPNil(t *testing.T) {
	assert.Equal(t, "", clientIP(nil))
}

func TestTrustedProxiesFromEnv(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "")
	assert.Nil(t, TrustedProxiesFromEnv())
	t.Setenv("TRUSTED_PROXIES", "10.0.0.1, 10.0.0.0/8")
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.0/8"}, TrustedProxiesFromEnv())
}

func TestLocalLimiterCapsKeys(t *testing.T) {
	l := newLocalLimiter()
	l.max = 2
	b := Bucket{Rate: 100, Burst: 100}
	l.allow("a", b)
	l.allow("b", b)
	l.allow("c", b)
	l.mu.Lock()
	n := len(l.m)
	l.mu.Unlock()
	assert.LessOrEqual(t, n, 2)
}

func TestForwardedForCannotBypassBlacklist(t *testing.T) {
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.IPBlacklist = map[string]struct{}{"192.0.2.1": {}}
	r := gin.New()
	require.NoError(t, r.SetTrustedProxies(nil))
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(rdb, cfg))
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestForwardedForCannotSkipViaWhitelist(t *testing.T) {
	_, rdb := testRedis(t)
	cfg := DefaultConfig()
	cfg.IP = Bucket{Rate: 1, Burst: 1}
	cfg.Route = Bucket{Rate: 100, Burst: 100}
	cfg.IPWhitelist = map[string]struct{}{"8.8.8.8": {}}
	r := gin.New()
	require.NoError(t, r.SetTrustedProxies(nil))
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(rdb, cfg))
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })
	mk := func() *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("X-Forwarded-For", "8.8.8.8")
		return req
	}
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, mk())
	assert.Equal(t, 200, w1.Code)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, mk())
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}
