package httpx

import (
	"net/http"
	"net/url"
	"strings"
)

// SanitizeOrigin returns scheme://host for http(s) URLs and rejects everything else
// (javascript:, protocol-relative, userinfo, empty host).
func SanitizeOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// RequestOrigin builds a public origin from Host plus the first X-Forwarded-Proto
// hop, falling back to the TLS state. Non-http(s) proto values are ignored.
// X-Forwarded-Host is ignored so a client cannot redirect codes or mail links off-site.
func RequestOrigin(r *http.Request) string {
	if r == nil {
		return ""
	}
	proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if i := strings.Index(proto, ","); i >= 0 {
		proto = strings.TrimSpace(proto[:i])
	}
	if proto != "http" && proto != "https" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	return SanitizeOrigin(proto + "://" + host)
}
