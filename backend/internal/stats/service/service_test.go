package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/internal/stats/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct{}

func (stubRepo) MemberSummary(context.Context, repo.Query) (int64, int64, int64, error) {
	return 12, 10, 2, nil
}
func (stubRepo) MemberByDepartment(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "d1", Label: "技术研发部", Value: 7}, {Key: "d2", Label: "人力资源部", Value: 5}}, nil
}
func (stubRepo) MemberByGrade(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "2023", Label: "2023", Value: 6}}, nil
}
func (stubRepo) MemberTrend(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "2026-01", Label: "2026-01", Value: 3}}, nil
}
func (stubRepo) InterviewSummary(context.Context, repo.Query) (int64, float64, float64, error) {
	return 8, 0.625, 82.5, nil
}
func (stubRepo) InterviewByDepartment(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "d1", Label: "技术研发部", Value: 4}}, nil
}
func (stubRepo) InterviewTrend(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "2026-01", Label: "2026-01", Value: 2}}, nil
}
func (stubRepo) InterviewScoreHist(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "80", Label: "80-89", Value: 3}}, nil
}
func (stubRepo) MeetingAttendanceTrend(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "2026-01", Label: "2026-01", Value: 90}}, nil
}
func (stubRepo) MeetingByDepartment(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "d1", Label: "技术研发部", Value: 2}}, nil
}
func (stubRepo) MeetingCalendar(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "2026-01-15", Label: "2026-01-15", Value: 1}}, nil
}
func (stubRepo) TaskByStatus(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "0", Label: "待处理", Value: 1}, {Key: "2", Label: "已完成", Value: 4}}, nil
}
func (stubRepo) TaskOnTimeRate(context.Context, repo.Query) (float64, error) { return 0.8, nil }
func (stubRepo) TaskTrend(context.Context, repo.Query) ([]string, map[string][]repo.Bucket, error) {
	return []string{"0", "2"}, map[string][]repo.Bucket{
		"0": {{Key: "2026-01", Label: "2026-01", Value: 1}},
		"2": {{Key: "2026-01", Label: "2026-01", Value: 2}},
	}, nil
}
func (stubRepo) InternshipRanking(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "u1", Label: "张三", Value: 90}}, nil
}
func (stubRepo) InternshipDeptAvg(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "d1", Label: "技术研发部", Value: 60}}, nil
}
func (stubRepo) InternshipTrend(context.Context, repo.Query) ([]repo.Bucket, error) {
	return []repo.Bucket{{Key: "2026-01", Label: "2026-01", Value: 45}}, nil
}
func (stubRepo) Overview(context.Context, uuid.UUID) (*dto.OverviewResponse, error) {
	return &dto.OverviewResponse{TotalMembers: 12, TotalMeetingsThisMonth: 2, TotalTasksInProgress: 3}, nil
}

func TestRegistryListAndMissing(t *testing.T) {
	svc := NewStatsService(stubRepo{})
	list := svc.ListProviders()
	require.Len(t, list, 5)
	_, err := svc.GetStats(context.Background(), "nope", nil)
	require.Error(t, err)
	app, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeStatsProviderNotFound, app.Code)
}

func TestProvidersAll(t *testing.T) {
	svc := NewStatsService(stubRepo{})
	codes := []string{"member-distribution", "interview-data", "meeting-attendance", "task-progress", "internship-duration"}
	for _, code := range codes {
		res, err := svc.GetStats(context.Background(), code, &dto.StatsQuery{Granularity: "month", GroupBy: "department"})
		require.NoError(t, err, code)
		require.NotNil(t, res)
		assert.Equal(t, code, res.Provider)
		assert.NotEmpty(t, res.Series)
		assert.NotNil(t, res.ChartConfig)
	}
}

func TestMemberSummaryFields(t *testing.T) {
	svc := NewStatsService(stubRepo{})
	res, err := svc.GetStats(context.Background(), "member-distribution", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(12), res.Summary["total_members"])
	assert.Equal(t, "pie", res.Series[0].Type)
	assert.Equal(t, "bar", res.Series[1].Type)
	assert.Equal(t, "line", res.Series[2].Type)
}

func TestInterviewGauge(t *testing.T) {
	svc := NewStatsService(stubRepo{})
	res, err := svc.GetStats(context.Background(), "interview-data", nil)
	require.NoError(t, err)
	assert.Equal(t, "gauge", res.Series[0].Type)
	assert.InDelta(t, 0.625, res.Summary["pass_rate"], 0.0001)
}

func TestInvalidQuery(t *testing.T) {
	svc := NewStatsService(stubRepo{})
	_, err := svc.GetStats(context.Background(), "task-progress", &dto.StatsQuery{Granularity: "year"})
	require.Error(t, err)
	_, err = svc.GetStats(context.Background(), "task-progress", &dto.StatsQuery{GroupBy: "unknown"})
	require.Error(t, err)
}

func TestOverviewAndExport(t *testing.T) {
	svc := NewStatsService(stubRepo{})
	ov, err := svc.Overview(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, int64(12), ov.TotalMembers)

	_, _, err = svc.Export(context.Background(), "member-distribution", "pdf", nil)
	require.Error(t, err)

	data, name, err := svc.Export(context.Background(), "member-distribution", "csv", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, name, "csv")

	data, name, err = svc.Export(context.Background(), "interview-data", "excel", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, name, "xlsx")
}

func TestExportUnknownProvider(t *testing.T) {
	svc := NewStatsService(stubRepo{})
	_, _, err := svc.Export(context.Background(), "ghost", "csv", nil)
	require.Error(t, err)
}

func TestExportTooLarge(t *testing.T) {
	svc := NewStatsService(hugeRepo{})
	_, _, err := svc.Export(context.Background(), "member-distribution", "csv", nil)
	require.Error(t, err)
	app, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeStatsTooLarge, app.Code)
}

func TestValidateQueryRange(t *testing.T) {
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	err := validateQuery(&dto.StatsQuery{StartDate: &start, EndDate: &end})
	require.Error(t, err)
}

type hugeRepo struct{ stubRepo }

func (hugeRepo) MemberByDepartment(context.Context, repo.Query) ([]repo.Bucket, error) {
	out := make([]repo.Bucket, 5001)
	for i := range out {
		out[i] = repo.Bucket{Key: "k", Label: "l", Value: 1}
	}
	return out, nil
}

func TestRound4(t *testing.T) {
	assert.Equal(t, 1.2346, round4(1.23455))
}

func TestToSeriesEmptyLabel(t *testing.T) {
	s := toSeries("x", "bar", []repo.Bucket{{Key: "k", Value: 1}})
	assert.Equal(t, "k", s.Data[0].Label)
	assert.Equal(t, []string{"k"}, s.XAxis)
}

func TestRegistryRegisterNil(t *testing.T) {
	r := NewRegistry()
	r.Register(nil)
	assert.Empty(t, r.ListProviders())
}
