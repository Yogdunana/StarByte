package middleware

import "github.com/Yogdunana/StarByte/backend/pkg/audit"

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
	return audit.DesensitizeJSON(body)
}
