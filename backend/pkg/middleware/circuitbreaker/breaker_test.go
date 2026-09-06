package circuitbreaker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBreaker_ErrorRateTripsAndHalfOpen(t *testing.T) {
	b := New(Settings{MinRequests: 4, ErrorRate: 0.5, Window: 10, OpenFor: 20 * time.Millisecond, HalfOpenProbes: 2, P99: time.Hour})
	fixed := time.Unix(1_700_000_000, 0)
	b.now = func() time.Time { return fixed }

	for i := 0; i < 4; i++ {
		assert.True(t, b.Allow("r"))
		b.Record("r", true, time.Millisecond)
	}
	assert.Equal(t, StateOpen, b.State("r"))
	assert.False(t, b.Allow("r"))

	b.now = func() time.Time { return fixed.Add(30 * time.Millisecond) }
	assert.True(t, b.Allow("r"))
	assert.Equal(t, StateHalfOpen, b.State("r"))
	b.Record("r", false, time.Millisecond)
	assert.True(t, b.Allow("r"))
	b.Record("r", false, time.Millisecond)
	assert.Equal(t, StateClosed, b.State("r"))
}

func TestBreaker_HalfOpenFailureReopens(t *testing.T) {
	b := New(Settings{MinRequests: 2, ErrorRate: 0.5, Window: 10, OpenFor: time.Millisecond, HalfOpenProbes: 2, P99: time.Hour})
	assert.True(t, b.Allow("x"))
	b.Record("x", true, 0)
	assert.True(t, b.Allow("x"))
	b.Record("x", true, 0)
	assert.Equal(t, StateOpen, b.State("x"))
	time.Sleep(2 * time.Millisecond)
	assert.True(t, b.Allow("x"))
	b.Record("x", true, 0)
	assert.Equal(t, StateOpen, b.State("x"))
}

func TestBreaker_P99Trips(t *testing.T) {
	b := New(Settings{MinRequests: 5, ErrorRate: 1, Window: 10, P99: 10 * time.Millisecond, OpenFor: time.Minute})
	for i := 0; i < 5; i++ {
		assert.True(t, b.Allow("slow"))
		b.Record("slow", false, 20*time.Millisecond)
	}
	assert.Equal(t, StateClosed, b.State("slow"), "P99 must not trip before the window is full")
	for i := 0; i < 5; i++ {
		assert.True(t, b.Allow("slow"))
		b.Record("slow", false, 20*time.Millisecond)
	}
	assert.Equal(t, StateOpen, b.State("slow"))
}

func TestMiddleware_OpensAndDegrades(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b := New(Settings{MinRequests: 2, ErrorRate: 0.5, Window: 10, OpenFor: time.Hour, P99: time.Hour})
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(b, CacheDegrade(map[string]string{"fallback": "1"})))
	r.GET("/boom", func(c *gin.Context) { c.Status(500) })

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
		assert.Equal(t, 500, w.Code)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "1", w.Header().Get("X-Degraded"))
	assert.Contains(t, w.Body.String(), "fallback")
	assert.Contains(t, w.Body.String(), "21004")
}

func TestMiddleware_OpenWithoutDegrade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b := New(Settings{MinRequests: 2, ErrorRate: 0.5, Window: 8, OpenFor: time.Hour, P99: time.Hour})
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(b, nil))
	r.GET("/boom", func(c *gin.Context) { c.Status(http.StatusBadGateway) })
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
		assert.Equal(t, http.StatusBadGateway, w.Code)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "21002")
}

func TestMiddleware_PingDegrade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b := New(Settings{MinRequests: 2, ErrorRate: 0.5, Window: 8, OpenFor: time.Hour, P99: time.Hour})
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(b, PingDegrade))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusBadGateway) })
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
		assert.Equal(t, http.StatusBadGateway, w.Code)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "21004")
	assert.Contains(t, w.Body.String(), "pong")
	assert.Equal(t, "1", w.Header().Get("X-Degraded"))
}

func TestMiddleware_PingDegradeDeclinesOtherRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b := New(Settings{MinRequests: 2, ErrorRate: 0.5, Window: 8, OpenFor: time.Hour, P99: time.Hour})
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(Middleware(b, PingDegrade))
	r.GET("/boom", func(c *gin.Context) { c.Status(http.StatusBadGateway) })
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
		assert.Equal(t, http.StatusBadGateway, w.Code)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Empty(t, w.Header().Get("X-Degraded"))
	assert.Contains(t, w.Body.String(), "21002")
}

func TestMiddleware_PanicCountsAsFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b := New(Settings{MinRequests: 2, ErrorRate: 0.5, Window: 8, OpenFor: time.Hour, P99: time.Hour})
	r := gin.New()
	r.Use(gin.CustomRecovery(func(c *gin.Context, _ any) {
		c.Status(http.StatusInternalServerError)
	}))
	r.Use(Middleware(b, nil))
	r.GET("/boom", func(c *gin.Context) { panic("boom") })
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
	assert.Equal(t, StateOpen, b.State("GET:/boom"))
}

func TestBreaker_HalfOpenProbesExhausted(t *testing.T) {
	b := New(Settings{MinRequests: 2, ErrorRate: 0.5, Window: 10, OpenFor: time.Millisecond, HalfOpenProbes: 1, P99: time.Hour})
	assert.True(t, b.Allow("z"))
	b.Record("z", true, 0)
	assert.True(t, b.Allow("z"))
	b.Record("z", true, 0)
	assert.Equal(t, StateOpen, b.State("z"))
	time.Sleep(2 * time.Millisecond)
	assert.True(t, b.Allow("z"))
	assert.False(t, b.Allow("z"))
}

func TestNew_Defaults(t *testing.T) {
	b := New(Settings{})
	assert.Equal(t, StateClosed, b.State("n"))
	assert.True(t, b.Allow("n"))
}

func TestNilBreakerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(nil, nil))
	r.GET("/ok", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ok", nil))
	assert.Equal(t, 200, w.Code)
}

func TestPercentile(t *testing.T) {
	assert.Equal(t, time.Duration(0), percentile(nil, 0.99))
	got := percentile([]time.Duration{time.Second, 2 * time.Second, 3 * time.Second}, 0.99)
	assert.Equal(t, 3*time.Second, got)
}
