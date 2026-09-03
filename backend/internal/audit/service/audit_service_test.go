package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/stretchr/testify/assert"
)

func TestDesensitize_EmptyString(t *testing.T) {
	result := Desensitize("")
	assert.Equal(t, "", result)
}

func TestDesensitize_NoSensitiveData(t *testing.T) {
	input := `{"name":"test","email":"test@example.com","age":25}`
	result := Desensitize(input)
	assert.Equal(t, input, result)
}

func TestDesensitize_PasswordField(t *testing.T) {
	input := `{"username":"admin","password":"secret123"}`
	result := Desensitize(input)
	assert.Contains(t, result, `"[redacted]"`)
	assert.NotContains(t, result, "secret123")
	assert.Contains(t, result, "admin")
}

func TestDesensitize_TokenField(t *testing.T) {
	input := `{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9","user_id":"123"}`
	result := Desensitize(input)
	assert.Contains(t, result, `"[redacted]"`)
	assert.NotContains(t, result, "eyJhbGci")
	assert.Contains(t, result, "123")
}

func TestDesensitize_MultipleSensitiveFields(t *testing.T) {
	input := `{"old_password":"old123","new_password":"new456","secret":"abc"}`
	result := Desensitize(input)
	assert.Contains(t, result, `"[redacted]"`)
	assert.NotContains(t, result, "old123")
	assert.NotContains(t, result, "new456")
	assert.NotContains(t, result, "abc")
}

func TestDesensitize_CaseInsensitive(t *testing.T) {
	input := `{"Password":"secret123"}`
	result := Desensitize(input)
	assert.Contains(t, result, `"[redacted]"`)
	assert.NotContains(t, result, "secret123")
}

func TestDesensitize_RefreshToken(t *testing.T) {
	input := `{"refresh_token":"tok_abc123xyz","user":"admin"}`
	result := Desensitize(input)
	assert.Contains(t, result, `"[redacted]"`)
	assert.NotContains(t, result, "tok_abc123xyz")
	assert.Contains(t, result, "admin")
}

func TestExportCSV_BasicFormat(t *testing.T) {
	s := &auditService{}
	logs := []model.AuditLog{
		{
			Username:       "admin",
			Operation:      "POST /api/v1/users",
			Method:         "POST",
			Path:           "/api/v1/users",
			IP:             "192.168.1.1",
			ResponseStatus: 200,
			DurationMs:     45,
			RequestID:      "req-123",
		},
	}

	data := s.exportCSV(logs)
	str := string(data)

	// Should contain BOM
	assert.Equal(t, byte(0xEF), data[0])
	assert.Equal(t, byte(0xBB), data[1])
	assert.Equal(t, byte(0xBF), data[2])

	// Should contain headers
	assert.Contains(t, str, "ID")
	assert.Contains(t, str, "用户名")
	assert.Contains(t, str, "操作")

	// Should contain data
	assert.Contains(t, str, "admin")
	assert.Contains(t, str, "POST /api/v1/users")
	assert.Contains(t, str, "192.168.1.1")
	assert.Contains(t, str, "req-123")
}

func TestExportCSV_EmptyLogs(t *testing.T) {
	s := &auditService{}
	data := s.exportCSV([]model.AuditLog{})

	// Should still have BOM and headers
	assert.Equal(t, byte(0xEF), data[0])
	str := string(data)
	assert.Contains(t, str, "用户名")
}

func TestExportCSV_MultipleRows(t *testing.T) {
	s := &auditService{}
	logs := []model.AuditLog{
		{Username: "admin", Operation: "POST /users", Method: "POST"},
		{Username: "user1", Operation: "DELETE /users/123", Method: "DELETE"},
	}

	data := s.exportCSV(logs)
	str := string(data)

	assert.Contains(t, str, "admin")
	assert.Contains(t, str, "user1")
	assert.Contains(t, str, "POST /users")
	assert.Contains(t, str, "DELETE /users/123")
}

func TestExportJSON_BasicFormat(t *testing.T) {
	s := &auditService{}
	logs := []model.AuditLog{
		{
			Username:       "admin",
			Operation:      "POST /api/v1/users",
			Method:         "POST",
			Path:           "/api/v1/users",
			IP:             "192.168.1.1",
			ResponseStatus: 200,
			DurationMs:     45,
			RequestID:      "req-123",
		},
	}

	data, err := s.exportJSON(logs)
	assert.NoError(t, err)
	str := string(data)

	assert.Contains(t, str, "admin")
	assert.Contains(t, str, "POST /api/v1/users")
	assert.Contains(t, str, "192.168.1.1")
	assert.Contains(t, str, "req-123")
}

func TestExportJSON_EmptyLogs(t *testing.T) {
	s := &auditService{}
	data, err := s.exportJSON([]model.AuditLog{})
	assert.NoError(t, err)
	str := string(data)

	// Should produce a valid JSON array (empty)
	assert.Contains(t, str, "[]")
}

func TestExportJSON_DesensitizesRequestParams(t *testing.T) {
	s := &auditService{}
	logs := []model.AuditLog{
		{
			RequestParams: `{"password":"secret123","name":"test"}`,
		},
	}

	data, err := s.exportJSON(logs)
	assert.NoError(t, err)
	str := string(data)

	assert.Contains(t, str, "[redacted]")
	assert.NotContains(t, str, "secret123")
	assert.Contains(t, str, "test")
}

func TestSha256Hex(t *testing.T) {
	result := sha256Hex([]byte(""))
	// SHA256 of empty string
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", result)
}

func TestHmacSha256(t *testing.T) {
	key := []byte("test-key")
	data := []byte("test-data")
	result := hmacSha256(key, data)
	assert.NotEmpty(t, result)
	assert.Len(t, result, 32) // SHA256 produces 32 bytes
}

func TestGetSigningKey(t *testing.T) {
	key := getSigningKey("secret", "20240101", "us-east-1", "s3")
	assert.NotEmpty(t, key)
	assert.Len(t, key, 32) // SHA256 produces 32 bytes
}

func TestDesensitize_NonJSONString(t *testing.T) {
	// Non-JSON strings should be returned unchanged
	input := "plain text without json"
	result := Desensitize(input)
	assert.Equal(t, input, result)
}

func TestDesensitize_PartialJSON(t *testing.T) {
	// Malformed JSON should be returned without changes to non-sensitive parts
	input := `{"name":"test","password":"secret"`
	result := Desensitize(input)
	// Even with malformed JSON, the regex should still work
	assert.Contains(t, result, "[redacted]")
	assert.NotContains(t, result, "secret")
}

func TestModelTableName(t *testing.T) {
	assert.Equal(t, "audit_logs", model.AuditLog{}.TableName())
	assert.Equal(t, "audit_log_archives", model.AuditLogArchive{}.TableName())
}
