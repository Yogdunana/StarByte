package ratelimit

import (
	"hash/fnv"
	"net"
	"strings"

	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
)

const (
	ContextTrafficColor = "traffic_color"
	ContextGrayHit      = "gray_hit"
	HeaderTrafficColor  = "X-Traffic-Color"
)

func viewerID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return strings.TrimSpace(auth.GetUserID(c))
}

func routeKey(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return "unknown"
	}
	return c.Request.Method + ":" + c.FullPath()
}

func paintTraffic(c *gin.Context, cfg Config) {
	color := strings.TrimSpace(c.GetHeader(HeaderTrafficColor))
	if color == "" {
		color = "prod"
		seed := viewerID(c)
		if seed == "" {
			seed = clientIP(c)
		}
		if grayByPercent(seed, cfg.GrayPercent) || inSet(cfg.GrayAllow, c.GetHeader(cfg.GrayHeader)) {
			color = "gray"
			c.Set(ContextGrayHit, true)
		}
	} else if color == "gray" {
		c.Set(ContextGrayHit, true)
	}
	c.Set(ContextTrafficColor, color)
	c.Header(HeaderTrafficColor, color)
}

func grayByPercent(seed string, pct int) bool {
	if pct <= 0 {
		return false
	}
	if pct >= 100 {
		return true
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(seed))
	return int(h.Sum32()%100) < pct
}

func deniedByACL(c *gin.Context, cfg Config) bool {
	return ipListed(cfg.IPBlacklist, clientIP(c)) || inSet(cfg.UserBlacklist, viewerID(c))
}

func skipLimit(c *gin.Context, cfg Config) bool {
	return ipListed(cfg.IPWhitelist, clientIP(c)) || inSet(cfg.UserWhitelist, viewerID(c))
}

// clientIP uses gin.Context.ClientIP, which honors Engine trusted proxies.
// Production must call SetTrustedProxies(nil) unless TRUSTED_PROXIES is set;
// otherwise Gin v1.9 defaults to trusting 0.0.0.0/0 and X-Forwarded-For is spoofable.
func clientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.ClientIP()
}

func ipListed(set map[string]struct{}, ip string) bool {
	if inSet(set, ip) {
		return true
	}
	parsed := net.ParseIP(ip)
	if parsed == nil || set == nil {
		return false
	}
	for spec := range set {
		if !strings.Contains(spec, "/") {
			continue
		}
		_, n, err := net.ParseCIDR(spec)
		if err == nil && n.Contains(parsed) {
			return true
		}
	}
	return false
}
