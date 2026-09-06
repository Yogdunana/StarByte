package middleware

import (
	"strings"
	"unicode/utf8"

	"github.com/Yogdunana/StarByte/backend/pkg/audit"
)

var sensitivePaths = map[string]bool{
	"/api/v1/auth/login":    true,
	"/api/v1/auth/register": true,
	"/api/v1/auth/refresh":  true,
	"/api/v1/user/password": true,
	"/api/v1/auth/password": true,
}

func sanitizeRequestBody(path, body string) string {
	if sensitivePaths[path] {
		return "[redacted: sensitive endpoint]"
	}
	return audit.DesensitizeJSON(body)
}

func sanitizeResponseBody(body string) string {
	if body == "" {
		return ""
	}
	if looksBinaryResponse(body) {
		return "[binary response omitted]"
	}
	return audit.DesensitizeJSON(body)
}

func looksBinaryResponse(body string) bool {
	if !utf8.ValidString(body) || strings.ContainsRune(body, 0) {
		return true
	}
	if strings.HasPrefix(body, "%PDF") || strings.HasPrefix(body, "PK\x03\x04") {
		return true
	}
	return false
}
