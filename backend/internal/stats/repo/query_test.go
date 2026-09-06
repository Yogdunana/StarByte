package repo

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func setupRepo(t *testing.T) StatsRepo {
	t.Helper()
	return NewStatsRepo(testutil.OpenPostgres(t))
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
	_, err = r.RankingHidden(ctx)
	require.NoError(t, err)
	_, err = r.InternshipRanking(ctx, Query{HideRanking: true})
	require.NoError(t, err)
}
