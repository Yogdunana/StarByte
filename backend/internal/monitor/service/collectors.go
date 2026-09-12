package service

import (
	"context"
	"database/sql"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/prometheus/client_golang/prometheus"
	dtoProm "github.com/prometheus/client_model/go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var processStart = time.Now()

func appVersion() string {
	if v := os.Getenv("APP_VERSION"); v != "" {
		return v
	}
	return "1.0.0"
}

func collectApp(now time.Time) *dto.AppHealth {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	lastGC := uint64(0)
	if ms.LastGC > 0 {
		lastGC = ms.LastGC / 1e6
	}
	return &dto.AppHealth{
		StartedAt:     processStart.UTC().Format(time.RFC3339),
		UptimeSeconds: int64(now.Sub(processStart).Seconds()),
		Version:       appVersion(),
		GoVersion:     runtime.Version(),
		Goroutines:    runtime.NumGoroutine(),
		MemAlloc:      ms.Alloc,
		MemSys:        ms.Sys,
		HeapAlloc:     ms.HeapAlloc,
		HeapSys:       ms.HeapSys,
		HeapInuse:     ms.HeapInuse,
		StackInuse:    ms.StackInuse,
		NumGC:         ms.NumGC,
		NextGC:        ms.NextGC,
		LastGCUnixMs:  lastGC,
		PauseTotalNs:  ms.PauseTotalNs,
		GCCPUFraction: ms.GCCPUFraction,
		CollectedAt:   now.UTC().Format(time.RFC3339),
	}
}

type dbPool interface {
	Stats() sql.DBStats
}

func collectDatabase(pool dbPool, now time.Time) *dto.DatabaseStatus {
	if pool == nil {
		return &dto.DatabaseStatus{Available: false, CollectedAt: now.UTC().Format(time.RFC3339)}
	}
	st := pool.Stats()
	return &dto.DatabaseStatus{
		Available:          true,
		OpenConnections:    st.OpenConnections,
		InUse:              st.InUse,
		Idle:               st.Idle,
		WaitCount:          st.WaitCount,
		WaitDurationMs:     st.WaitDuration.Milliseconds(),
		MaxOpenConnections: st.MaxOpenConnections,
		MaxIdleClosed:      st.MaxIdleClosed,
		MaxLifetimeClosed:  st.MaxLifetimeClosed,
		CollectedAt:        now.UTC().Format(time.RFC3339),
	}
}

func sqlPoolFromGorm(db *gorm.DB) (dbPool, error) {
	if db == nil {
		return nil, dbDown("数据库未配置")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, dbDown("获取数据库连接池失败")
	}
	return sqlDB, nil
}

func collectRedis(ctx context.Context, rdb *redis.Client, now time.Time) (*dto.RedisStatus, error) {
	stamp := now.UTC().Format(time.RFC3339)
	if rdb == nil {
		return &dto.RedisStatus{Available: false, CollectedAt: stamp}, nil
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, redisDown("Redis 不可用")
	}
	raw, err := rdb.Info(pingCtx).Result()
	if err != nil {
		return nil, redisDown("读取 Redis INFO 失败")
	}
	fields := parseRedisInfo(raw)
	hits := fields["keyspace_hits"]
	misses := fields["keyspace_misses"]
	out := &dto.RedisStatus{
		Available:        true,
		ConnectedClients: fields["connected_clients"],
		UsedMemory:       fields["used_memory"],
		MaxMemory:        fields["maxmemory"],
		KeyspaceHits:     hits,
		KeyspaceMisses:   misses,
		HitRate:          hitRate(hits, misses),
		Keys:             sumKeyspaceKeys(raw),
		Version:          infoString(raw, "redis_version"),
		CollectedAt:      stamp,
	}
	return out, nil
}

func parseRedisInfo(raw string) map[string]int64 {
	out := map[string]int64{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			continue
		}
		out[strings.TrimSpace(k)] = n
	}
	return out
}

func infoString(raw, key string) string {
	prefix := key + ":"
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func sumKeyspaceKeys(raw string) int64 {
	var total int64
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "db") || !strings.Contains(line, "keys=") {
			continue
		}
		_, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		for _, part := range strings.Split(rest, ",") {
			k, v, ok := strings.Cut(part, "=")
			if ok && strings.TrimSpace(k) == "keys" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err == nil {
					total += n
				}
			}
		}
	}
	return total
}

func hitRate(hits, misses int64) float64 {
	total := hits + misses
	if total <= 0 {
		return 0
	}
	return round2(float64(hits) / float64(total) * 100)
}

const apiStatsTODO = "TODO(phase-2): persist request histograms for P50/P95/P99; scrape /metrics starbyte_http_request_duration_seconds"

func collectAPIStats(gatherer prometheus.Gatherer, now time.Time) *dto.APIStats {
	stamp := now.UTC().Format(time.RFC3339)
	out := &dto.APIStats{
		Source:          "prometheus",
		PercentilesNote: apiStatsTODO,
		CollectedAt:     stamp,
	}
	if gatherer == nil {
		return out
	}
	families, err := gatherer.Gather()
	if err != nil {
		return out
	}
	var req, errs float64
	for _, mf := range families {
		if mf.GetName() != "starbyte_http_requests_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			val := counterValue(m)
			req += val
			if isErrorStatus(m.GetLabel()) {
				errs += val
			}
		}
	}
	out.Available = true
	out.RequestTotal = req
	out.ErrorTotal = errs
	if req > 0 {
		out.ErrorRate = round2(errs / req * 100)
	}
	return out
}

func counterValue(m *dtoProm.Metric) float64 {
	if m == nil || m.Counter == nil || m.Counter.Value == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

func isErrorStatus(labels []*dtoProm.LabelPair) bool {
	for _, l := range labels {
		if l.GetName() != "status" {
			continue
		}
		s := l.GetValue()
		return strings.HasPrefix(s, "5")
	}
	return false
}
