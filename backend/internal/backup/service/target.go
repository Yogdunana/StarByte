package service

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
)

// Compose service names that address the same Postgres as DB_HOST=postgres.
var composeClusterHosts = map[string]struct{}{
	"postgres":          {},
	"starbyte-postgres": {},
}

var loopbackHosts = map[string]struct{}{
	"localhost": {},
	"127.0.0.1": {},
	"::1":       {},
	"[::1]":     {},
}

// ResolveDrillTarget builds a Postgres target from DSN / field overlays.
// It refuses to point at the live application database.
// Live user/password/ssl are reused only for the same cluster (host aliases + port)
// with a different database name. A different host never inherits the live password.
func ResolveDrillTarget(live config.DatabaseConfig, req *dto.DrillRequest) (config.DatabaseConfig, error) {
	if req == nil {
		return config.DatabaseConfig{}, errNotReady("缺少演练目标")
	}
	var over config.DatabaseConfig
	passwordFromRequest := false
	if dsn := strings.TrimSpace(req.TargetDSN); dsn != "" {
		parsed, err := ParsePostgresTarget(dsn)
		if err != nil {
			return config.DatabaseConfig{}, err
		}
		over = overlayDB(over, parsed)
		if dsnPasswordProvided(dsn) {
			passwordFromRequest = true
		}
	}
	if v := strings.TrimSpace(req.TargetHost); v != "" {
		over.Host = v
	}
	if req.TargetPort > 0 {
		over.Port = req.TargetPort
	}
	if v := strings.TrimSpace(req.TargetUser); v != "" {
		over.User = v
	}
	if req.TargetPassword != "" {
		over.Password = req.TargetPassword
		passwordFromRequest = true
	}
	if v := strings.TrimSpace(req.TargetDBName); v != "" {
		over.DBName = v
	}
	if v := strings.TrimSpace(req.TargetSSLMode); v != "" {
		over.SSLMode = v
	}

	out := config.DatabaseConfig{
		Host:     firstNonEmpty(over.Host, live.Host),
		Port:     firstPort(over.Port, live.Port),
		User:     over.User,
		Password: over.Password,
		DBName:   over.DBName,
		SSLMode:  over.SSLMode,
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
	if SameCluster(live, out) {
		if out.User == "" {
			out.User = live.User
		}
		if out.Password == "" {
			out.Password = live.Password
		}
		if out.SSLMode == "" {
			out.SSLMode = live.SSLMode
		}
		return out, nil
	}
	if !passwordFromRequest {
		return config.DatabaseConfig{}, errNotReady("跨主机演练须提供 target_password 或 DSN 密码，不会复用生产库凭据")
	}
	if out.User == "" {
		out.User = live.User
	}
	if out.SSLMode == "" {
		out.SSLMode = live.SSLMode
	}
	return out, nil
}

// SameCluster reports whether two configs address the same Postgres instance.
func SameCluster(a, b config.DatabaseConfig) bool {
	return sameHost(a.Host, b.Host) && normPort(a.Port) == normPort(b.Port)
}

// SameDatabase reports whether two configs address the same Postgres database.
func SameDatabase(a, b config.DatabaseConfig) bool {
	return SameCluster(a, b) && strings.TrimSpace(a.DBName) == strings.TrimSpace(b.DBName)
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
	if u.User != nil {
		if pass, ok := u.User.Password(); ok {
			out.Password = pass
		}
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

func dsnPasswordProvided(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil || u.User == nil {
			return false
		}
		_, ok := u.User.Password()
		return ok
	}
	for _, part := range strings.Fields(raw) {
		key, _, ok := strings.Cut(part, "=")
		if ok && strings.EqualFold(strings.TrimSpace(key), "password") {
			return true
		}
	}
	return false
}

func sameHost(a, b string) bool {
	x, y := normHost(a), normHost(b)
	if x == y {
		return true
	}
	_, ac := composeClusterHosts[x]
	_, bc := composeClusterHosts[y]
	return ac && bc
}

func normHost(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	if _, ok := loopbackHosts[h]; ok {
		return "localhost"
	}
	return h
}

func normPort(p int) int {
	if p == 0 {
		return 5432
	}
	return p
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func firstPort(over, live int) int {
	if over > 0 {
		return over
	}
	return live
}
