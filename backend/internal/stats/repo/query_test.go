package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRepo(t *testing.T) StatsRepo {
	t.Helper()
	candidates := []string{
		os.Getenv("STATS_TEST_DSN"),
		"host=localhost user=starbyte password=starbyte dbname=starbyte_dev port=5432 sslmode=disable",
		"host=localhost user=starbyte password=starbyte dbname=starbyte_test port=5432 sslmode=disable",
		"host=localhost user=postgres password=postgres dbname=starbyte_dev port=5432 sslmode=disable",
	}
	var last error
	for _, dsn := range candidates {
		if dsn == "" {
			continue
		}
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			last = err
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			last = err
			continue
		}
		if err := sqlDB.Ping(); err != nil {
			last = err
			continue
		}
		return NewStatsRepo(db)
	}
	t.Skipf("skipping stats repo DB test: %v", last)
	return nil
}

func TestRepoQueries(t *testing.T) {
	r := setupRepo(t)
	ctx := context.Background()
	end := time.Now()
	start := end.AddDate(0, -6, 0)
	dept := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	q := Query{Start: &start, End: &end, Granularity: "month", DepartmentID: &dept}

	_, _, _, err := r.MemberSummary(ctx, q)
	require.NoError(t, err)
	_, err = r.MemberByDepartment(ctx, q)
	require.NoError(t, err)
	_, err = r.MemberByGrade(ctx, q)
	require.NoError(t, err)
	_, err = r.MemberTrend(ctx, Query{Granularity: "day"})
	require.NoError(t, err)
	_, _, _, err = r.InterviewSummary(ctx, Query{Granularity: "week"})
	require.NoError(t, err)
	_, err = r.InterviewByDepartment(ctx, q)
	require.NoError(t, err)
	_, err = r.InterviewTrend(ctx, q)
	require.NoError(t, err)
	_, err = r.InterviewScoreHist(ctx, q)
	require.NoError(t, err)
	_, err = r.MeetingAttendanceTrend(ctx, q)
	require.NoError(t, err)
	_, err = r.MeetingByDepartment(ctx, q)
	require.NoError(t, err)
	_, err = r.MeetingCalendar(ctx, q)
	require.NoError(t, err)
	_, err = r.TaskByStatus(ctx, q)
	require.NoError(t, err)
	_, err = r.TaskOnTimeRate(ctx, q)
	require.NoError(t, err)
	_, _, err = r.TaskTrend(ctx, q)
	require.NoError(t, err)
	_, err = r.InternshipRanking(ctx, q)
	require.NoError(t, err)
	_, err = r.InternshipDeptAvg(ctx, q)
	require.NoError(t, err)
	_, err = r.InternshipTrend(ctx, q)
	require.NoError(t, err)
	_, err = r.Overview(ctx, uuid.Nil)
	require.NoError(t, err)
	_, err = r.Overview(ctx, uuid.New())
	require.NoError(t, err)
}
