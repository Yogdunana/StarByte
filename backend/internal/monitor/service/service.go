package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Service collects ops snapshots for the monitor dashboard.
type Service interface {
	Server(ctx context.Context) (*dto.ServerStatus, error)
	App(ctx context.Context) (*dto.AppHealth, error)
	Database(ctx context.Context) (*dto.DatabaseStatus, error)
	Redis(ctx context.Context) (*dto.RedisStatus, error)
	APIStats(ctx context.Context) (*dto.APIStats, error)
}

type monitorService struct {
	db       *gorm.DB
	rdb      *redis.Client
	host     HostSampler
	gatherer prometheus.Gatherer
	now      func() time.Time
}

// New builds the default collector (gopsutil + Prometheus default gatherer).
func New(db *gorm.DB, rdb *redis.Client) Service {
	return &monitorService{
		db:       db,
		rdb:      rdb,
		host:     newGopsutilHost(),
		gatherer: prometheus.DefaultGatherer,
		now:      time.Now,
	}
}

func (s *monitorService) Server(ctx context.Context) (*dto.ServerStatus, error) {
	return collectServer(ctx, s.host, s.now())
}

func (s *monitorService) App(context.Context) (*dto.AppHealth, error) {
	return collectApp(s.now()), nil
}

func (s *monitorService) Database(context.Context) (*dto.DatabaseStatus, error) {
	pool, err := sqlPoolFromGorm(s.db)
	if err != nil {
		return nil, err
	}
	return collectDatabase(pool, s.now()), nil
}

func (s *monitorService) Redis(ctx context.Context) (*dto.RedisStatus, error) {
	return collectRedis(ctx, s.rdb, s.now())
}

func (s *monitorService) APIStats(context.Context) (*dto.APIStats, error) {
	return collectAPIStats(s.gatherer, s.now()), nil
}
