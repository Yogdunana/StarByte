package httpx

import (
	"net"
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

// MailOrigin picks the origin written into activation emails.
// A configured public base always wins. Request Host is only used when that
// base is empty and the host is loopback/private (campus IP leak-test).
func MailOrigin(configured, request string) string {
	if base := SanitizeOrigin(configured); base != "" {
		return base
	}
	if base := SanitizeOrigin(request); base != "" && privateHTTPHost(base) {
		return base
	}
	return "http://127.0.0.1"
}

func privateHTTPHost(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
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
