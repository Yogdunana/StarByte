package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
