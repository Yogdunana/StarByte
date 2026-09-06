package middleware

import (
	"context"
	"encoding/json"
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

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
	assert.NotEmpty(t, resp["uptime"])
	assert.NotEmpty(t, resp["version"])
	assert.NotEmpty(t, resp["go_version"])
	_, hasCode := resp["code"]
	assert.False(t, hasCode, "health check should not include envelope 'code' field")
}

func TestReadinessCheck_NilDeps_Returns503(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health/ready", ReadinessCheck(nil, nil, nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "not_ready", resp["status"])
	checks, ok := resp["checks"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, checks, "database")
	assert.Contains(t, checks, "redis")
	assert.Contains(t, checks, "minio")
	_, hasCode := resp["code"]
	assert.False(t, hasCode)
}

func TestReadinessCheck_MinioOKOthersFail(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health/ready", ReadinessCheck(nil, nil, func(ctx context.Context) error { return nil }))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/ready", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"minio"`)
	assert.Contains(t, w.Body.String(), `"ok"`)
}
