package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

// MemberService 入会申请 + 人员档案。
type MemberService interface {
	Submit(ctx context.Context, userID uuid.UUID, req *dto.SubmitApplicationRequest) (*dto.ApplicationResponse, error)
	Resubmit(ctx context.Context, userID, id uuid.UUID, req *dto.ResubmitApplicationRequest) (*dto.ApplicationResponse, error)
	ListApplications(ctx context.Context, viewer uuid.UUID, req *dto.ListApplicationRequest, scope *rbacModel.DataScopeCondition) ([]*dto.ApplicationResponse, int64, error)
	GetApplication(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ApplicationResponse, error)
	MyApplications(ctx context.Context, userID uuid.UUID) ([]*dto.ApplicationResponse, error)
	ApplicationHistory(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.ApplicationHistoryResponse, error)
	Approve(ctx context.Context, reviewer, id uuid.UUID, comment string) (*dto.ApplicationResponse, error)
	Reject(ctx context.Context, reviewer, id uuid.UUID, comment string) (*dto.ApplicationResponse, error)
	Supplement(ctx context.Context, reviewer, id uuid.UUID, req *dto.SupplementRequest) (*dto.ApplicationResponse, error)
	ListDepartments(ctx context.Context) ([]dto.DepartmentOption, error)

	ListProfiles(ctx context.Context, viewer uuid.UUID, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]*dto.ProfileResponse, int64, error)
	GetProfile(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateProfileRequest, scope *rbacModel.DataScopeCondition) (*dto.ProfileResponse, error)
	UpdateProfileStatus(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateProfileStatusRequest) (*dto.ProfileResponse, error)
	ProfileHistory(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.ProfileHistoryResponse, error)
	ExportProfiles(ctx context.Context, viewer uuid.UUID, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]byte, error)

	ApplicationStats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error)
	MemberStats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error)

	SyncFromInterview(ctx context.Context, operator, applicationID uuid.UUID, result int16, comment string) error
}

type memberService struct {
	apps    repo.ApplicationRepo
	profs   repo.ProfileRepo
	starter InterviewStarter
}

// NewMemberService 创建会员服务。
func NewMemberService(apps repo.ApplicationRepo, profs repo.ProfileRepo, starter InterviewStarter) MemberService {
	return &memberService{apps: apps, profs: profs, starter: starter}
}
