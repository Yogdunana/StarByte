package middleware

import (
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
	// Verify no response envelope fields are present.
	_, hasCode := resp["code"]
	assert.False(t, hasCode, "health check should not include envelope 'code' field")
	_, hasMessage := resp["message"]
	assert.False(t, hasMessage, "health check should not include envelope 'message' field")
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

	// Verify the response is plain JSON without envelope.
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "not ready", resp["status"])
	assert.NotNil(t, resp["checks"])
	// Verify no response envelope fields are present.
	_, hasCode := resp["code"]
	assert.False(t, hasCode, "readiness check should not include envelope 'code' field")
}
