package ratelimit

import (
	"os"
	"strconv"
	"strings"
)

// LoadFromEnv starts from DefaultConfig and overlays TRAFFIC_* environment vars.
func LoadFromEnv() Config {
	cfg := DefaultConfig()
	if v := os.Getenv("TRAFFIC_IP_BLACKLIST"); v != "" {
		cfg.IPBlacklist = parseSet(v)
	}
	if v := os.Getenv("TRAFFIC_IP_WHITELIST"); v != "" {
		cfg.IPWhitelist = parseSet(v)
	}
	if v := os.Getenv("TRAFFIC_USER_BLACKLIST"); v != "" {
		cfg.UserBlacklist = parseSet(v)
	}
	if v := os.Getenv("TRAFFIC_USER_WHITELIST"); v != "" {
		cfg.UserWhitelist = parseSet(v)
	}
	if v := os.Getenv("TRAFFIC_GRAY_GROUPS"); v != "" {
		cfg.GrayAllow = parseSet(v)
	}
	if v := os.Getenv("TRAFFIC_GRAY_PERCENT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.GrayPercent = n
		}
	}
	overlayBucket(&cfg.IP, "TRAFFIC_IP_RATE", "TRAFFIC_IP_BURST")
	overlayBucket(&cfg.User, "TRAFFIC_USER_RATE", "TRAFFIC_USER_BURST")
	overlayBucket(&cfg.Route, "TRAFFIC_ROUTE_RATE", "TRAFFIC_ROUTE_BURST")
	return cfg
}

func overlayBucket(b *Bucket, rateKey, burstKey string) {
	if v := os.Getenv(rateKey); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			b.Rate = f
		}
	}
	if v := os.Getenv(burstKey); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			b.Burst = f
		}
	}
}

// TrustedProxiesFromEnv returns CIDRs/IPs allowed to set X-Forwarded-For.
// Empty / unset means trust none (ClientIP uses the socket peer).
func TrustedProxiesFromEnv() []string {
	v := os.Getenv("TRUSTED_PROXIES")
	if v == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseSet(csv string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out[p] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
