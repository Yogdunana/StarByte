package metrics

import (
	"context"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const namespace = "starbyte"

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration in seconds",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	DBConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "db_connections",
		Help:      "Database connection pool status",
	}, []string{"state"})

	RedisUp = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "redis_up",
		Help:      "Redis connection status (1=up, 0=down)",
	})

	Goroutines = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "goroutines",
		Help:      "Number of goroutines",
	})

	MemAllocBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_alloc_bytes",
		Help:      "Bytes of allocated heap objects",
	})
)

func init() {
	_ = prometheus.Register(collectors.NewGoCollector())
	_ = prometheus.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// CollectOnce snapshots runtime, DB pool, and Redis connectivity.
func CollectOnce(db *gorm.DB, rdb *redis.Client) {
	Goroutines.Set(float64(runtime.NumGoroutine()))
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	MemAllocBytes.Set(float64(ms.Alloc))

	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			st := sqlDB.Stats()
			DBConnections.WithLabelValues("active").Set(float64(st.InUse))
			DBConnections.WithLabelValues("idle").Set(float64(st.Idle))
			DBConnections.WithLabelValues("max").Set(float64(st.MaxOpenConnections))
		}
	}

	if rdb == nil {
		RedisUp.Set(0)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		RedisUp.Set(0)
		return
	}
	RedisUp.Set(1)
}

// StartCollectors refreshes gauges until ctx is done.
func StartCollectors(ctx context.Context, db *gorm.DB, rdb *redis.Client) {
	CollectOnce(db, rdb)
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				CollectOnce(db, rdb)
			}
		}
	}()
}
