package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEntity(t *testing.T) {
	typ, id := ParseEntity("/api/v1/users/11111111-1111-1111-1111-111111111111")
	assert.Equal(t, "user", typ)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", id)

	typ, id = ParseEntity("/api/v1/system/roles/22222222-2222-2222-2222-222222222222/permissions")
	assert.Equal(t, "role", typ)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", id)

	typ, id = ParseEntity("/api/v1/system/departments")
	assert.Equal(t, "department", typ)
	assert.Equal(t, "", id)
}

func TestComplianceFlags(t *testing.T) {
	assert.Equal(t, "delete", ComplianceFlags("DELETE", "/api/v1/users/x", ""))
	assert.Equal(t, "export", ComplianceFlags("GET", "/api/v1/system/audit-logs/export", ""))
	assert.Equal(t, "permission", ComplianceFlags("PUT", "/api/v1/system/roles/x/permissions", `{"permission_ids":[]}`))
	assert.Contains(t, ComplianceFlags("PUT", "/api/v1/users/x", `{"role_ids":["a"]}`), "permission")
}

func TestShouldAudit_GETExport(t *testing.T) {
	assert.True(t, shouldAudit("GET", "/api/v1/system/audit-logs/export"))
	assert.True(t, shouldAudit("GET", "/api/v1/system/audit-logs/reports"))
	assert.False(t, shouldAudit("GET", "/api/v1/system/audit-logs"))
	assert.False(t, shouldAudit("POST", "/api/v1/auth/login"))
}

func TestFillTraceFields_UsesSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/11111111-1111-1111-1111-111111111111", strings.NewReader(`{"real_name":"新"}`))
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
	SetAuditSnapshot(c, "user", "11111111-1111-1111-1111-111111111111", map[string]any{
		"real_name": "旧",
		"username":  "keep",
		"password":  "secret123",
	})

	entry := &AuditLogEntry{Method: "PUT", Path: "/api/v1/users/11111111-1111-1111-1111-111111111111"}
	fillTraceFields(c, entry, `{"real_name":"新"}`, "")
	assert.Equal(t, "user", entry.EntityType)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", entry.EntityID)
	assert.NotContains(t, entry.BeforeJSON, "secret123")
	assert.Contains(t, entry.AfterJSON, "新")
	assert.Contains(t, entry.AfterJSON, "keep")
	require.Contains(t, entry.DiffJSON, "real_name")
	assert.NotContains(t, entry.DiffJSON, "username")
}

func TestFillTraceFields_PrefersResponseAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/11111111-1111-1111-1111-111111111111", nil)
	c.Params = gin.Params{{Key: "id", Value: "11111111-1111-1111-1111-111111111111"}}
	SetAuditSnapshot(c, "user", "11111111-1111-1111-1111-111111111111", map[string]any{
		"real_name": "旧",
		"username":  "keep",
	})
	entry := &AuditLogEntry{Method: "PUT", Path: "/api/v1/users/11111111-1111-1111-1111-111111111111"}
	fillTraceFields(c, entry, `{"real_name":"新"}`, `{"code":0,"data":{"id":"11111111-1111-1111-1111-111111111111","real_name":"新","username":"keep"}}`)
	assert.Contains(t, entry.AfterJSON, `"username":"keep"`)
	assert.NotContains(t, entry.DiffJSON, `"path":"username"`)
}

func TestFillTraceFields_CreateEntityIDFromResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"username":"newuser"}`))
	entry := &AuditLogEntry{Method: "POST", Path: "/api/v1/users"}
	created := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	fillTraceFields(c, entry, `{"username":"newuser"}`, `{"code":0,"data":{"id":"`+created+`","username":"newuser"}}`)
	assert.Equal(t, "user", entry.EntityType)
	assert.Equal(t, created, entry.EntityID)
	assert.Contains(t, entry.AfterJSON, created)
}

func TestFillTraceFields_ExportAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/audit-logs/export", nil)
	entry := &AuditLogEntry{Method: "GET", Path: "/api/v1/system/audit-logs/export"}
	fillTraceFields(c, entry, "", "")
	assert.Equal(t, "EXPORT", entry.Action)
	assert.Equal(t, "export", entry.ComplianceFlags)
}

func TestExtractDataJSON(t *testing.T) {
	body := `{"code":0,"data":{"id":"abc","password":"secret123"}}`
	got := extractDataJSON(body)
	assert.Contains(t, got, "abc")
	assert.NotContains(t, got, "secret123")
}
