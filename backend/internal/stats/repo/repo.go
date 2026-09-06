package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StatsRepo 统计只读查询。
type StatsRepo interface {
	MemberSummary(ctx context.Context, q Query) (total, active, newMonth int64, err error)
	MemberByDepartment(ctx context.Context, q Query) ([]Bucket, error)
	MemberByGrade(ctx context.Context, q Query) ([]Bucket, error)
	MemberTrend(ctx context.Context, q Query) ([]Bucket, error)
	InterviewSummary(ctx context.Context, q Query) (total int64, passRate, avgScore float64, err error)
	InterviewByDepartment(ctx context.Context, q Query) ([]Bucket, error)
	InterviewTrend(ctx context.Context, q Query) ([]Bucket, error)
	InterviewScoreHist(ctx context.Context, q Query) ([]Bucket, error)
	MeetingAttendanceTrend(ctx context.Context, q Query) ([]Bucket, error)
	MeetingByDepartment(ctx context.Context, q Query) ([]Bucket, error)
	MeetingCalendar(ctx context.Context, q Query) ([]Bucket, error)
	TaskByStatus(ctx context.Context, q Query) ([]Bucket, error)
	TaskOnTimeRate(ctx context.Context, q Query) (float64, error)
	TaskTrend(ctx context.Context, q Query) (statuses []string, points map[string][]Bucket, err error)
	InternshipRanking(ctx context.Context, q Query) ([]Bucket, error)
	InternshipDeptAvg(ctx context.Context, q Query) ([]Bucket, error)
	InternshipTrend(ctx context.Context, q Query) ([]Bucket, error)
	Overview(ctx context.Context, userID uuid.UUID) (*dto.OverviewResponse, error)
}

type statsRepo struct {
	db *gorm.DB
}

// NewStatsRepo 创建仓储。
func NewStatsRepo(db *gorm.DB) StatsRepo {
	return &statsRepo{db: db}
}
