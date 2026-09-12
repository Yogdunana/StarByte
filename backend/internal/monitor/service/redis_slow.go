package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"github.com/redis/go-redis/v9"
)

func collectRedisSlow(ctx context.Context, rdb *redis.Client, now time.Time, limit int) []dto.SlowQuery {
	if rdb == nil {
		return []dto.SlowQuery{}
	}
	if limit <= 0 {
		limit = 16
	}
	qctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	entries, err := rdb.SlowLogGet(qctx, int64(limit)).Result()
	if err != nil {
		return []dto.SlowQuery{}
	}
	return mapRedisSlow(entries, now)
}

func mapRedisSlow(entries []redis.SlowLog, now time.Time) []dto.SlowQuery {
	stamp := now.UTC().Format(time.RFC3339)
	out := make([]dto.SlowQuery, 0, len(entries))
	for _, e := range entries {
		cmd := database.SanitizeSQL(strings.Join(e.Args, " "))
		if cmd == "" {
			continue
		}
		ms := float64(e.Duration.Microseconds()) / 1000
		out = append(out, dto.SlowQuery{
			Query:       cmd,
			Calls:       1,
			MeanTimeMs:  round2(ms),
			TotalTimeMs: round2(ms),
			MaxTimeMs:   round2(ms),
			Source:      "redis_slowlog",
			CollectedAt: stamp,
		})
	}
	return out
}
