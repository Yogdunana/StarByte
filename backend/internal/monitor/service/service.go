package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/Yogdunana/StarByte/backend/internal/monitor/repo"
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
	SlowQueries(ctx context.Context) (*dto.SlowQueries, error)
	Snapshot(ctx context.Context) *dto.LiveSnapshot
}

type monitorService struct {
	db       *gorm.DB
	rdb      *redis.Client
	host     HostSampler
	gatherer prometheus.Gatherer
	slow     repo.SlowQueryRepo
	now      func() time.Time
}

// New builds the default collector (gopsutil + Prometheus default gatherer).
func New(db *gorm.DB, rdb *redis.Client) Service {
	return &monitorService{
		db:       db,
		rdb:      rdb,
		host:     newGopsutilHost(),
		gatherer: prometheus.DefaultGatherer,
		slow:     repo.NewSlowQueryRepo(db),
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

func (s *monitorService) SlowQueries(ctx context.Context) (*dto.SlowQueries, error) {
	limit := 20
	var out *dto.SlowQueries
	if s.slow != nil {
		out = s.slow.List(ctx, limit)
	} else {
		out = &dto.SlowQueries{
			Available:   true,
			Source:      "in_memory",
			Note:        "慢查询仓库未配置",
			Queries:     []dto.SlowQuery{},
			CollectedAt: s.now().UTC().Format(time.RFC3339),
		}
	}
	out.RedisCommands = collectRedisSlow(ctx, s.rdb, s.now(), 16)
	if out.Queries == nil {
		out.Queries = []dto.SlowQuery{}
	}
	return out, nil
}

func (s *monitorService) Snapshot(ctx context.Context) *dto.LiveSnapshot {
	out := &dto.LiveSnapshot{}
	if v, err := s.Server(ctx); err == nil {
		out.Server = v
	} else {
		out.Errors = append(out.Errors, "server")
	}
	if v, err := s.App(ctx); err == nil {
		out.App = v
	} else {
		out.Errors = append(out.Errors, "app")
	}
	if v, err := s.Database(ctx); err == nil {
		out.Database = v
	} else {
		out.Errors = append(out.Errors, "database")
	}
	if v, err := s.Redis(ctx); err == nil {
		out.Redis = v
	} else {
		out.Errors = append(out.Errors, "redis")
	}
	if v, err := s.APIStats(ctx); err == nil {
		out.API = v
	} else {
		out.Errors = append(out.Errors, "api")
	}
	return out
}
