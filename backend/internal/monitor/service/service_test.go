package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/metrics"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/alicebob/miniredis/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeHost struct {
	cpuErr, memErr, diskErr, loadErr error
	cpu                              float64
	memUsed, memTotal                uint64
	memPct                           float64
	diskUsed, diskTotal              uint64
	diskPct                          float64
	l1, l5, l15                      float64
}

func (f fakeHost) CPUPercent(context.Context) (float64, error) {
	return f.cpu, f.cpuErr
}
func (f fakeHost) VirtualMemory() (uint64, uint64, float64, error) {
	return f.memUsed, f.memTotal, f.memPct, f.memErr
}
func (f fakeHost) DiskUsage(string) (uint64, uint64, float64, error) {
	return f.diskUsed, f.diskTotal, f.diskPct, f.diskErr
}
func (f fakeHost) LoadAvg() (float64, float64, float64, error) {
	return f.l1, f.l5, f.l15, f.loadErr
}

func TestCollectServer(t *testing.T) {
	now := time.Date(2026, 9, 12, 8, 0, 0, 0, time.UTC)
	out, err := collectServer(context.Background(), fakeHost{
		cpu: 12.345, memUsed: 100, memTotal: 400, memPct: 25.4,
		diskUsed: 50, diskTotal: 200, diskPct: 25, l1: 0.41, l5: 0.5, l15: 0.62,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 12.35, out.CPUPercent)
	assert.Equal(t, uint64(100), out.MemUsed)
	assert.Equal(t, "/", out.DiskPath)
	assert.Equal(t, 0.41, out.Load1)
	assert.Equal(t, "2026-09-12T08:00:00Z", out.CollectedAt)

	_, err = collectServer(context.Background(), fakeHost{cpuErr: assert.AnError}, now)
	require.Error(t, err)
	app, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeMonitorCollectFail, app.Code)
}

func TestCollectApp(t *testing.T) {
	out := collectApp(time.Now())
	assert.NotEmpty(t, out.StartedAt)
	assert.GreaterOrEqual(t, out.UptimeSeconds, int64(0))
	assert.Greater(t, out.Goroutines, 0)
	assert.NotEmpty(t, out.GoVersion)
	assert.Equal(t, "1.0.0", out.Version)
}

func TestCollectDatabase(t *testing.T) {
	now := time.Now()
	empty := collectDatabase(nil, now)
	assert.False(t, empty.Available)

	pool := stubPool{st: sql.DBStats{OpenConnections: 4, InUse: 1, Idle: 3, WaitCount: 2, MaxOpenConnections: 10}}
	out := collectDatabase(pool, now)
	assert.True(t, out.Available)
	assert.Equal(t, 4, out.OpenConnections)
	assert.Equal(t, 1, out.InUse)
	assert.Equal(t, 3, out.Idle)
	assert.Equal(t, int64(2), out.WaitCount)
	assert.Equal(t, 10, out.MaxOpenConnections)
}

type stubPool struct{ st sql.DBStats }

func (s stubPool) Stats() sql.DBStats { return s.st }

func TestSQLPoolFromGormNil(t *testing.T) {
	_, err := sqlPoolFromGorm(nil)
	require.Error(t, err)
	app, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeMonitorDBDown, app.Code)
}

func TestCollectRedis(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	empty, err := collectRedis(ctx, nil, now)
	require.NoError(t, err)
	assert.False(t, empty.Available)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	require.NoError(t, rdb.Set(ctx, "k1", "v", 0).Err())

	out, err := collectRedis(ctx, rdb, now)
	require.NoError(t, err)
	assert.True(t, out.Available)
	assert.GreaterOrEqual(t, out.ConnectedClients, int64(0))
	assert.GreaterOrEqual(t, out.UsedMemory, int64(0))
	assert.GreaterOrEqual(t, out.Keys, int64(0))

	_ = rdb.Close()
	_, err = collectRedis(ctx, rdb, now)
	require.Error(t, err)
	app, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeMonitorRedisDown, app.Code)
}

func TestParseRedisInfoAndHitRate(t *testing.T) {
	raw := `# Server
redis_version:7.2.4
# Clients
connected_clients:3
# Memory
used_memory:2048
maxmemory:0
# Stats
keyspace_hits:80
keyspace_misses:20
# Keyspace
db0:keys=5,expires=1,avg_ttl=0
db1:keys=2,expires=0,avg_ttl=0
`
	fields := parseRedisInfo(raw)
	assert.Equal(t, int64(3), fields["connected_clients"])
	assert.Equal(t, int64(80), fields["keyspace_hits"])
	assert.Equal(t, "7.2.4", infoString(raw, "redis_version"))
	assert.Equal(t, int64(7), sumKeyspaceKeys(raw))
	assert.Equal(t, 80.0, hitRate(80, 20))
	assert.Equal(t, 0.0, hitRate(0, 0))
}

func TestCollectAPIStats(t *testing.T) {
	metrics.ResetRecentHTTPLatenciesForTest()
	t.Cleanup(metrics.ResetRecentHTTPLatenciesForTest)

	reg := prometheus.NewRegistry()
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "starbyte_http_requests_total",
		Help: "test",
	}, []string{"method", "path", "status"})
	require.NoError(t, reg.Register(c))
	c.WithLabelValues("GET", "/api/v1/x", "200").Add(10)
	c.WithLabelValues("GET", "/api/v1/x", "500").Add(2)

	out := collectAPIStats(reg, time.Now())
	assert.True(t, out.Available)
	assert.Equal(t, 12.0, out.RequestTotal)
	assert.Equal(t, 2.0, out.ErrorTotal)
	assert.Equal(t, 16.67, out.ErrorRate)
	assert.Nil(t, out.P50Seconds)
	assert.Contains(t, out.PercentilesNote, "尚无请求样本")

	empty := collectAPIStats(nil, time.Now())
	assert.False(t, empty.Available)
}

func TestCollectAPIStatsRecentPercentiles(t *testing.T) {
	metrics.ResetRecentHTTPLatenciesForTest()
	t.Cleanup(metrics.ResetRecentHTTPLatenciesForTest)
	for i := 0; i < 20; i++ {
		metrics.RecordHTTPLatency(0.02)
	}
	out := collectAPIStats(nil, time.Now())
	require.NotNil(t, out.P50Seconds)
	assert.InDelta(t, 0.02, *out.P50Seconds, 0.0001)
	assert.Contains(t, out.Source, "recent_requests")
}

func TestLiveHostAndNilDeps(t *testing.T) {
	svc := New(nil, nil)
	ctx := context.Background()

	server, err := svc.Server(ctx)
	require.NoError(t, err)
	assert.Equal(t, "/", server.DiskPath)
	assert.Greater(t, server.MemTotal, uint64(0))
	assert.Greater(t, server.DiskTotal, uint64(0))
	assert.GreaterOrEqual(t, server.CPUPercent, 0.0)

	app, err := svc.App(ctx)
	require.NoError(t, err)
	assert.Greater(t, app.Goroutines, 0)

	_, err = svc.Database(ctx)
	require.Error(t, err)

	redisOut, err := svc.Redis(ctx)
	require.NoError(t, err)
	assert.False(t, redisOut.Available)

	slow, err := svc.SlowQueries(ctx)
	require.NoError(t, err)
	assert.True(t, slow.Available)
	assert.Equal(t, "in_memory", slow.Source)
	assert.Empty(t, slow.RedisCommands)

	snap := svc.Snapshot(ctx)
	require.NotNil(t, snap.Server)
	require.NotNil(t, snap.App)
	assert.Contains(t, snap.Errors, "database")
}

func TestServiceAppAndAPIStats(t *testing.T) {
	svc := &monitorService{
		host:     fakeHost{cpu: 1, memTotal: 1, diskTotal: 1},
		gatherer: prometheus.DefaultGatherer,
		now:      time.Now,
	}
	app, err := svc.App(context.Background())
	require.NoError(t, err)
	assert.Greater(t, app.Goroutines, 0)

	metrics.HTTPRequestsTotal.WithLabelValues("GET", "/health", "200").Inc()
	stats, err := svc.APIStats(context.Background())
	require.NoError(t, err)
	assert.True(t, stats.Available)

	server, err := svc.Server(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "/", server.DiskPath)

	slow, err := svc.SlowQueries(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "in_memory", slow.Source)
}

func TestRound2(t *testing.T) {
	assert.Equal(t, 1.24, round2(1.235))
	assert.Equal(t, 1.24, round2(1.244))
}

func TestDTOShapeHasNoSecrets(t *testing.T) {
	raw := dto.RedisStatus{Version: "7"}
	assert.Empty(t, raw.Version[:0])
	assert.False(t, strings.Contains(strings.ToLower(dto.SlowQuery{Query: "SELECT 1"}.Query), "password"))
}

func TestQuantileFromHistogram(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "starbyte_http_request_duration_seconds",
		Help:    "t",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1},
	}, []string{"method", "path"})
	require.NoError(t, reg.Register(h))
	for i := 0; i < 90; i++ {
		h.WithLabelValues("GET", "/a").Observe(0.02)
	}
	for i := 0; i < 10; i++ {
		h.WithLabelValues("GET", "/a").Observe(0.4)
	}
	metrics.ResetRecentHTTPLatenciesForTest()
	t.Cleanup(metrics.ResetRecentHTTPLatenciesForTest)
	p50, p95, p99, src, note := resolvePercentiles(reg)
	require.NotNil(t, p50)
	require.NotNil(t, p95)
	require.NotNil(t, p99)
	assert.Equal(t, "prometheus_histogram", src)
	assert.Contains(t, note, "直方图")
	assert.Greater(t, *p95, *p50)
}

func TestMapRedisSlow(t *testing.T) {
	entries := []redis.SlowLog{{
		Args:     []string{"GET", "session:1"},
		Duration: 1500 * time.Microsecond,
	}}
	got := mapRedisSlow(entries, time.Date(2026, 9, 12, 8, 0, 0, 0, time.UTC))
	require.Len(t, got, 1)
	assert.Equal(t, "GET session:1", got[0].Query)
	assert.Equal(t, "redis_slowlog", got[0].Source)
	assert.InDelta(t, 1.5, got[0].MeanTimeMs, 0.01)
	assert.Empty(t, collectRedisSlow(context.Background(), nil, time.Now(), 0))
}
