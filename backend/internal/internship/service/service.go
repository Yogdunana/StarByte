package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

type InternshipService interface {
	Create(ctx context.Context, operator uuid.UUID, req *dto.CreateInternshipRequest, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error)
	List(ctx context.Context, viewer uuid.UUID, req *dto.ListInternshipRequest, scope *rbacModel.DataScopeCondition) ([]*dto.InternshipResponse, int64, error)
	Get(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error)
	Update(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateInternshipRequest, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error)
	Delete(ctx context.Context, operator, id uuid.UUID, scope *rbacModel.DataScopeCondition) error
	ListMine(ctx context.Context, userID uuid.UUID, status *int16) ([]*dto.InternshipResponse, error)
	Complete(ctx context.Context, operator, id uuid.UUID, req *dto.CompleteRequest, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error)
	SubmitReport(ctx context.Context, operator, id uuid.UUID, report string, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error)
	MentorComment(ctx context.Context, operator, id uuid.UUID, comment string, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error)
	DurationStats(ctx context.Context, viewer uuid.UUID, req *dto.DurationStatsRequest, scope *rbacModel.DataScopeCondition) (*dto.DurationStatsResponse, error)
	Ranking(ctx context.Context, viewer uuid.UUID, req *dto.RankingRequest, scope *rbacModel.DataScopeCondition) (*dto.RankingResponse, error)
	DepartmentStats(ctx context.Context, viewer uuid.UUID, req *dto.DepartmentStatsRequest, scope *rbacModel.DataScopeCondition) (*dto.DepartmentStatsResponse, error)
	GetConfig(ctx context.Context) (*dto.InternshipConfigResponse, error)
	UpdateConfig(ctx context.Context, operator uuid.UUID, req *dto.InternshipConfigRequest) (*dto.InternshipConfigResponse, error)
}

type internshipService struct {
	rows repo.InternshipRepo
}

func NewInternshipService(rows repo.InternshipRepo) InternshipService {
	return &internshipService{rows: rows}
}
