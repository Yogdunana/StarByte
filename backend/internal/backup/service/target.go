package service

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
)

// ResolveDrillTarget builds a Postgres target from DSN / field overlays.
// It refuses to point at the live application database.
func ResolveDrillTarget(live config.DatabaseConfig, req *dto.DrillRequest) (config.DatabaseConfig, error) {
	if req == nil {
		return config.DatabaseConfig{}, errNotReady("缺少演练目标")
	}
	out := live
	if dsn := strings.TrimSpace(req.TargetDSN); dsn != "" {
		parsed, err := ParsePostgresTarget(dsn)
		if err != nil {
			return config.DatabaseConfig{}, err
		}
		out = overlayDB(out, parsed)
	}
	if v := strings.TrimSpace(req.TargetHost); v != "" {
		out.Host = v
	}
	if req.TargetPort > 0 {
		out.Port = req.TargetPort
	}
	if v := strings.TrimSpace(req.TargetUser); v != "" {
		out.User = v
	}
	if req.TargetPassword != "" {
		out.Password = req.TargetPassword
	}
	if v := strings.TrimSpace(req.TargetDBName); v != "" {
		out.DBName = v
	}
	if v := strings.TrimSpace(req.TargetSSLMode); v != "" {
		out.SSLMode = v
	}
	if strings.TrimSpace(out.DBName) == "" {
		return config.DatabaseConfig{}, errNotReady("演练须指定 target_dbname 或 DSN 中的库名")
	}
	if strings.TrimSpace(out.Host) == "" {
		return config.DatabaseConfig{}, errNotReady("演练须指定目标主机")
	}
	if SameDatabase(live, out) {
		return config.DatabaseConfig{}, errInvalidState("演练目标不能是当前应用库，请换库名或主机")
	}
	return out, nil
}

// SameDatabase reports whether two configs address the same Postgres database.
func SameDatabase(a, b config.DatabaseConfig) bool {
	return normHost(a.Host) == normHost(b.Host) &&
		normPort(a.Port) == normPort(b.Port) &&
		strings.TrimSpace(a.DBName) == strings.TrimSpace(b.DBName)
}

// ParsePostgresTarget accepts postgres:// URLs or libpq key=value strings.
func ParsePostgresTarget(raw string) (config.DatabaseConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return config.DatabaseConfig{}, errNotReady("目标 DSN 为空")
	}
	if strings.Contains(raw, "://") {
		return parsePostgresURL(raw)
	}
	return parseLibpq(raw)
}

func parsePostgresURL(raw string) (config.DatabaseConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return config.DatabaseConfig{}, errNotReady("目标 DSN 无效")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "postgres" && scheme != "postgresql" {
		return config.DatabaseConfig{}, errNotReady("目标 DSN 须为 postgres:// 或 postgresql://")
	}
	out := config.DatabaseConfig{
		Host:    u.Hostname(),
		User:    u.User.Username(),
		DBName:  strings.TrimPrefix(u.Path, "/"),
		SSLMode: u.Query().Get("sslmode"),
	}
	if pass, ok := u.User.Password(); ok {
		out.Password = pass
	}
	if u.Port() != "" {
		p, err := strconv.Atoi(u.Port())
		if err != nil {
			return config.DatabaseConfig{}, errNotReady("目标端口无效")
		}
		out.Port = p
	}
	if i := strings.Index(out.DBName, "/"); i >= 0 {
		out.DBName = out.DBName[:i]
	}
	return out, nil
}

func parseLibpq(raw string) (config.DatabaseConfig, error) {
	out := config.DatabaseConfig{}
	for _, part := range strings.Fields(raw) {
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			return config.DatabaseConfig{}, errNotReady("目标 DSN 须为 key=value")
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "host":
			out.Host = val
		case "port":
			p, err := strconv.Atoi(val)
			if err != nil {
				return config.DatabaseConfig{}, errNotReady("目标端口无效")
			}
			out.Port = p
		case "user":
			out.User = val
		case "password":
			out.Password = val
		case "dbname":
			out.DBName = val
		case "sslmode":
			out.SSLMode = val
		}
	}
	return out, nil
}

func overlayDB(base, over config.DatabaseConfig) config.DatabaseConfig {
	if over.Host != "" {
		base.Host = over.Host
	}
	if over.Port > 0 {
		base.Port = over.Port
	}
	if over.User != "" {
		base.User = over.User
	}
	if over.Password != "" {
		base.Password = over.Password
	}
	if over.DBName != "" {
		base.DBName = over.DBName
	}
	if over.SSLMode != "" {
		base.SSLMode = over.SSLMode
	}
	return base
}

func normHost(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	switch h {
	case "127.0.0.1", "::1":
		return "localhost"
	default:
		return h
	}
}

func normPort(p int) int {
	if p == 0 {
		return 5432
	}
	return p
}
