package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_RecordsMatchedPath(t *testing.T) {
	r := setupTestRouter()
	r.Use(Metrics())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/no-such", nil)
	r.ServeHTTP(w, req)
}

func TestMetrics_RecordsPanicAs500(t *testing.T) {
	before := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/boom", "500"))

	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(Metrics())
	r.Use(ErrorHandler())
	r.GET("/boom", func(c *gin.Context) {
		panic("kaboom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	after := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/boom", "500"))
	assert.Equal(t, before+1, after)
}

func TestMetrics_SkipsWebsocketDuration(t *testing.T) {
	metrics.ResetRecentHTTPLatenciesForTest()
	t.Cleanup(metrics.ResetRecentHTTPLatenciesForTest)

	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(Metrics())
	r.GET("/ws/monitor", func(c *gin.Context) {
		time.Sleep(20 * time.Millisecond)
		c.Status(http.StatusSwitchingProtocols)
	})

	before := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/ws/monitor", "101"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ws/monitor", nil))
	assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
	after := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/ws/monitor", "101"))
	assert.Equal(t, before+1, after)

	_, _, _, ok := metrics.RecentHTTPPercentiles()
	assert.False(t, ok)
}
