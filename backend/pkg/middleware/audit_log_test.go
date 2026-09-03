package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuditLog_NilDB_NoOp(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLog_GETNotAudited(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// GET requests should not be audited — handler should run normally.
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLog_POSTAudited(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	body := `{"name":"test","value":123}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// POST should be audited — response should still be returned.
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLog_PUTAudited(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.PUT("/test/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/test/123", strings.NewReader(`{"name":"updated"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLog_DELETEAudited(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.DELETE("/test/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/test/123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLog_PATCHAudited(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.PATCH("/test/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "patched"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/test/123", strings.NewReader(`{"field":"value"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLog_RequestBodyRestored(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.POST("/echo", func(c *gin.Context) {
		body := make([]byte, 1024)
		n, _ := c.Request.Body.Read(body)
		c.Data(http.StatusOK, "text/plain", body[:n])
	})

	body := `{"message":"hello world"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// The downstream handler should be able to read the body even though
	// the audit middleware already read it.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, body, w.Body.String())
}

func TestAuditLog_ResponseCaptured(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "data": "response"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// The actual response should still be written to the client.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "response")
}

func TestAuditLog_LargeResponseBodyTruncated(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.POST("/large", func(c *gin.Context) {
		// Generate a response larger than maxResponseBodySize
		large := strings.Repeat("x", maxResponseBodySize+1000)
		c.Data(http.StatusOK, "text/plain", []byte(large))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/large", nil)
	r.ServeHTTP(w, req)

	// The client should receive the full response.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, maxResponseBodySize+1000, w.Body.Len())
}

func TestWriteMethods(t *testing.T) {
	assert.True(t, writeMethods["POST"])
	assert.True(t, writeMethods["PUT"])
	assert.True(t, writeMethods["PATCH"])
	assert.True(t, writeMethods["DELETE"])
	assert.False(t, writeMethods["GET"])
	assert.False(t, writeMethods["OPTIONS"])
	assert.False(t, writeMethods["HEAD"])
}

func TestSanitizeRequestBody_SensitivePath(t *testing.T) {
	body := `{"username":"admin","password":"secret123"}`
	result := sanitizeRequestBody("/api/v1/auth/login", body)
	assert.Equal(t, "[redacted: sensitive endpoint]", result)
}

func TestSanitizeRequestBody_SensitiveFieldRedacted(t *testing.T) {
	body := `{"name":"test","password":"secret123","email":"test@example.com"}`
	result := sanitizeRequestBody("/api/v1/user/profile", body)
	assert.Contains(t, result, `"[redacted]"`)
	assert.NotContains(t, result, "secret123")
	assert.Contains(t, result, "test@example.com")
}

func TestSanitizeRequestBody_NoSensitiveData(t *testing.T) {
	body := `{"name":"test","email":"test@example.com"}`
	result := sanitizeRequestBody("/api/v1/user/profile", body)
	assert.Equal(t, body, result)
}

func TestSanitizeRequestBody_MultipleSensitiveFields(t *testing.T) {
	body := `{"old_password":"old123","new_password":"new456","secret":"abc"}`
	result := sanitizeRequestBody("/api/v1/user/profile", body)
	assert.Contains(t, result, `"[redacted]"`)
	assert.NotContains(t, result, "old123")
	assert.NotContains(t, result, "new456")
	assert.NotContains(t, result, "abc")
}

func TestCloseAuditWriter_NoWriterExists_NoOp(t *testing.T) {
	// Ensure no writer is initialized.
	CloseAuditWriter()
	// Should not panic.
}

func TestCloseAuditWriter_AfterUse_DrainsAndResets(t *testing.T) {
	// Initialize the writer with nil db (no-op writes).
	r := setupTestRouter()
	r.Use(AuditLog(nil))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Close the writer — should drain and reset.
	CloseAuditWriter()

	// After closing, AuditLog should still work (creates a new writer).
	r2 := setupTestRouter()
	r2.Use(AuditLog(nil))
	r2.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req2.Header.Set("Content-Type", "application/json")
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Clean up.
	CloseAuditWriter()
}
