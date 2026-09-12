package repo

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"gorm.io/gorm"
)

const defaultSlowLimit = 20

// SlowQueryRepo loads slow SQL from Postgres and/or the in-process GORM ring.
type SlowQueryRepo interface {
	List(ctx context.Context, limit int) *dto.SlowQueries
}

type gormRunner interface {
	WithContext(ctx context.Context) *gorm.DB
}

type pgQuerier interface {
	hasExtension(ctx context.Context, name string) (bool, error)
	listStatements(ctx context.Context, limit int) ([]dto.SlowQuery, error)
	listActivity(ctx context.Context, limit int) ([]dto.SlowQuery, error)
}

type slowQueryRepo struct {
	db      gormRunner
	querier pgQuerier
	now     func() time.Time
}

// NewSlowQueryRepo uses GORM when db is non-nil; otherwise only the in-memory ring.
func NewSlowQueryRepo(db *gorm.DB) SlowQueryRepo {
	r := &slowQueryRepo{now: time.Now}
	if db != nil {
		r.db = db
		r.querier = r
	}
	return r
}

func (r *slowQueryRepo) List(ctx context.Context, limit int) *dto.SlowQueries {
	if limit <= 0 {
		limit = defaultSlowLimit
	}
	stamp := r.now().UTC().Format(time.RFC3339)
	out := &dto.SlowQueries{
		Queries:     []dto.SlowQuery{},
		CollectedAt: stamp,
	}

	memory := fromGormRing(limit, stamp)
	if r.db == nil && r.querier == nil {
		out.Source = "in_memory"
		out.Note = "数据库未配置；仅返回进程内 GORM 慢查询环（阈值 500ms）。未启用 pg_stat_statements。"
		out.Queries = memory
		out.Available = true
		return out
	}

	q := r.querier
	if q == nil {
		q = r
	}
	if has, err := q.hasExtension(ctx, "pg_stat_statements"); err == nil && has {
		if rows, qerr := q.listStatements(ctx, limit); qerr == nil {
			out.Source = "pg_stat_statements"
			out.Note = "来自 pg_stat_statements；SQL 已截断。"
			out.Queries = mergeSlow(rows, memory, limit)
			out.Available = true
			return out
		}
	}

	activity, aerr := q.listActivity(ctx, limit)
	sources := []string{"in_memory"}
	combined := memory
	if aerr == nil && len(activity) > 0 {
		sources = append([]string{"pg_stat_activity"}, sources...)
		combined = mergeSlow(activity, memory, limit)
	}
	out.Source = strings.Join(sources, "+")
	out.Note = "未安装或无法读取 pg_stat_statements；已回退到 pg_stat_activity（当前长查询）与进程内 GORM 慢查询环。"
	out.Queries = combined
	out.Available = true
	return out
}

func (r *slowQueryRepo) hasExtension(ctx context.Context, name string) (bool, error) {
	var ok bool
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = ?)", name).
		Scan(&ok).Error
	return ok, err
}

type statementRow struct {
	Query       string  `gorm:"column:query"`
	Calls       int64   `gorm:"column:calls"`
	MeanTimeMs  float64 `gorm:"column:mean_time_ms"`
	TotalTimeMs float64 `gorm:"column:total_time_ms"`
	MaxTimeMs   float64 `gorm:"column:max_time_ms"`
	Rows        int64   `gorm:"column:rows"`
}

func (r *slowQueryRepo) listStatements(ctx context.Context, limit int) ([]dto.SlowQuery, error) {
	var rows []statementRow
	err := r.db.WithContext(ctx).Raw(`
SELECT
  query,
  calls,
  mean_exec_time AS mean_time_ms,
  total_exec_time AS total_time_ms,
  max_exec_time AS max_time_ms,
  rows
FROM pg_stat_statements
WHERE query NOT ILIKE '%pg_stat_statements%'
ORDER BY mean_exec_time DESC
LIMIT ?`, limit).Scan(&rows).Error
	if err != nil {
		err = r.db.WithContext(ctx).Raw(`
SELECT
  query,
  calls,
  mean_time AS mean_time_ms,
  total_time AS total_time_ms,
  max_time AS max_time_ms,
  rows
FROM pg_stat_statements
WHERE query NOT ILIKE '%pg_stat_statements%'
ORDER BY mean_time DESC
LIMIT ?`, limit).Scan(&rows).Error
	}
	if err != nil {
		return nil, err
	}
	return mapStatementRows(rows, "pg_stat_statements", r.now().UTC().Format(time.RFC3339)), nil
}

func (r *slowQueryRepo) listActivity(ctx context.Context, limit int) ([]dto.SlowQuery, error) {
	var rows []statementRow
	err := r.db.WithContext(ctx).Raw(`
SELECT
  query,
  1::bigint AS calls,
  EXTRACT(EPOCH FROM (now() - query_start)) * 1000 AS mean_time_ms,
  EXTRACT(EPOCH FROM (now() - query_start)) * 1000 AS total_time_ms,
  EXTRACT(EPOCH FROM (now() - query_start)) * 1000 AS max_time_ms,
  0::bigint AS rows
FROM pg_stat_activity
WHERE datname = current_database()
  AND state = 'active'
  AND pid <> pg_backend_pid()
  AND query_start IS NOT NULL
  AND now() - query_start >= interval '500 milliseconds'
  AND query NOT ILIKE '%pg_stat_%'
ORDER BY query_start
LIMIT ?`, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapStatementRows(rows, "pg_stat_activity", r.now().UTC().Format(time.RFC3339)), nil
}

func mapStatementRows(rows []statementRow, source, stamp string) []dto.SlowQuery {
	out := make([]dto.SlowQuery, 0, len(rows))
	for _, row := range rows {
		q := database.SanitizeSQL(row.Query)
		if q == "" {
			continue
		}
		out = append(out, dto.SlowQuery{
			Query:       q,
			Calls:       row.Calls,
			MeanTimeMs:  round2(row.MeanTimeMs),
			TotalTimeMs: round2(row.TotalTimeMs),
			MaxTimeMs:   round2(row.MaxTimeMs),
			Rows:        row.Rows,
			Source:      source,
			CollectedAt: stamp,
		})
	}
	return out
}

func fromGormRing(limit int, stamp string) []dto.SlowQuery {
	raw := database.RecentSlowQueries(limit)
	out := make([]dto.SlowQuery, 0, len(raw))
	for _, item := range raw {
		collected := stamp
		if !item.At.IsZero() {
			collected = item.At.UTC().Format(time.RFC3339)
		}
		out = append(out, dto.SlowQuery{
			Query:       item.SQL,
			Calls:       1,
			MeanTimeMs:  float64(item.DurationMs),
			TotalTimeMs: float64(item.DurationMs),
			MaxTimeMs:   float64(item.DurationMs),
			Rows:        item.Rows,
			Source:      "in_memory",
			CollectedAt: collected,
		})
	}
	return out
}

func mergeSlow(primary, extra []dto.SlowQuery, limit int) []dto.SlowQuery {
	seen := map[string]struct{}{}
	out := make([]dto.SlowQuery, 0, limit)
	add := func(items []dto.SlowQuery) {
		for _, item := range items {
			if len(out) >= limit {
				return
			}
			key := item.Source + "\n" + item.Query
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	add(primary)
	add(extra)
	return out
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
