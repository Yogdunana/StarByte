package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCompleteReportAndComment(t *testing.T) {
	svc, mem := newTestSvc()
	ctx := context.Background()
	owner := uuid.New()
	mentor := uuid.New()
	mem.users[owner] = &model.NamedUser{ID: owner, RealName: "学员"}
	mem.users[mentor] = &model.NamedUser{ID: mentor, RealName: "导师"}

	created, err := svc.Create(ctx, owner, &dto.CreateInternshipRequest{
		Title: "后端实习", Organization: "技术部", StartDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		MentorID: mentor.String(),
	}, nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)

	reported, err := svc.SubmitReport(ctx, owner, id, "阶段报告", nil)
	require.NoError(t, err)
	require.Equal(t, "阶段报告", reported.Report)

	done, err := svc.Complete(ctx, owner, id, &dto.CompleteRequest{Report: "结项报告", Achievements: "完成模块"}, nil)
	require.NoError(t, err)
	require.Equal(t, int16(1), done.Status)
	require.Equal(t, "结项报告", done.Report)
	require.NotNil(t, done.EndDate)

	_, err = svc.Complete(ctx, owner, id, &dto.CompleteRequest{}, nil)
	require.Error(t, err)
	ae, ok := err.(*response.AppError)
	require.True(t, ok)
	require.Equal(t, response.CodeInternshipDupComplete, ae.Code)

	commented, err := svc.MentorComment(ctx, mentor, id, "表现优秀", nil)
	require.NoError(t, err)
	require.Equal(t, "表现优秀", commented.MentorComment)
}

func TestMinisterCannotComplete(t *testing.T) {
	svc, mem := newTestSvc()
	ctx := context.Background()
	owner := uuid.New()
	minister := uuid.New()
	dept := uuid.New()
	mem.users[owner] = &model.NamedUser{ID: owner, RealName: "学员", DepartmentID: &dept}
	created, err := svc.Create(ctx, owner, &dto.CreateInternshipRequest{
		Title: "A", Organization: "B", StartDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}, nil)
	require.NoError(t, err)
	scope := &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}
	_, err = svc.Complete(ctx, minister, uuid.MustParse(created.ID), &dto.CompleteRequest{}, scope)
	require.Error(t, err)
	ae, ok := err.(*response.AppError)
	require.True(t, ok)
	require.Equal(t, response.CodeInternshipNoAccess, ae.Code)
}

func TestRankingHiddenAndStats(t *testing.T) {
	svc, mem := newTestSvc()
	ctx := context.Background()
	u1 := uuid.New()
	u2 := uuid.New()
	dept := uuid.New()
	mem.users[u1] = &model.NamedUser{ID: u1, RealName: "甲", DepartmentID: &dept}
	mem.users[u2] = &model.NamedUser{ID: u2, RealName: "乙", DepartmentID: &dept}
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(ctx, u1, &dto.CreateInternshipRequest{Title: "甲实习", Organization: "技术部", StartDate: start, EndDate: &end}, nil)
	require.NoError(t, err)
	_, err = svc.Create(ctx, u2, &dto.CreateInternshipRequest{Title: "乙实习", Organization: "技术部", StartDate: start}, nil)
	require.NoError(t, err)

	rank, err := svc.Ranking(ctx, u1, &dto.RankingRequest{SortBy: "duration"}, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, rank.Total, 2)
	require.Equal(t, 1, rank.Rankings[0].Rank)

	hidden := false
	_, err = svc.UpdateConfig(ctx, u1, &dto.InternshipConfigRequest{RankingVisible: &hidden})
	require.NoError(t, err)
	_, err = svc.Ranking(ctx, u1, &dto.RankingRequest{}, &rbacModel.DataScopeCondition{Query: "1 = 0"})
	require.Error(t, err)
	ae, ok := err.(*response.AppError)
	require.True(t, ok)
	require.Equal(t, response.CodeInternshipRankHidden, ae.Code)

	dur, err := svc.DurationStats(ctx, u1, &dto.DurationStatsRequest{GroupBy: "user"}, nil)
	require.NoError(t, err)
	require.Equal(t, "user", dur.GroupBy)
	require.NotEmpty(t, dur.Items)

	month, err := svc.DurationStats(ctx, u1, &dto.DurationStatsRequest{GroupBy: "month"}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, month.Items)

	deptStats, err := svc.DepartmentStats(ctx, u1, &dto.DepartmentStatsRequest{}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, deptStats.Items)
}

func TestEncodeDecodeSkills(t *testing.T) {
	require.Equal(t, []string{"Go", "React"}, decodeSkills(encodeSkills([]string{" Go ", "", "React"})))
	require.Empty(t, decodeSkills("not-json"))
	require.Empty(t, decodeSkills(""))
}

func TestDefaultAndParseConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.True(t, cfg.AllowStudentEdit)
	require.Equal(t, cfg, parseConfig(""))
	require.Equal(t, cfg, parseConfig("{"))
	got := parseConfig(`{"allow_student_edit":false,"allow_minister_edit":true,"ranking_visible":false}`)
	require.False(t, got.AllowStudentEdit)
	require.True(t, got.AllowMinisterEdit)
	require.False(t, got.RankingVisible)
}
