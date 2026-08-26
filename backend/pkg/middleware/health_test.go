package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck_ReturnsOK(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health", HealthCheck())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestReadinessCheck_NilDB_NilRedis_Returns503(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health/ready", ReadinessCheck(nil, nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "not ready")
	assert.Contains(t, body, "db")
	assert.Contains(t, body, "redis")
}

func TestReadinessCheck_ResponseFormat(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health/ready", ReadinessCheck(nil, nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/ready", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)

	// Verify the response contains the checks structure.
	body := w.Body.String()
	assert.Contains(t, body, "checks")
	assert.Contains(t, body, "fail")
}
