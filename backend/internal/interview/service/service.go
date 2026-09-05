package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

// Notifier 发送模板通知。
type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

// ApplicationSyncer 把面试结果写回入会申请。
type ApplicationSyncer interface {
	SyncFromInterview(ctx context.Context, operator, applicationID uuid.UUID, result int16, comment string) error
}

// InterviewService 面试管理。
type InterviewService interface {
	CreateSession(ctx context.Context, operator uuid.UUID, req *dto.CreateSessionRequest) (*dto.SessionResponse, error)
	ListSessions(ctx context.Context, viewer uuid.UUID, req *dto.ListSessionRequest, scope *rbacModel.DataScopeCondition) ([]*dto.SessionResponse, int64, error)
	GetSession(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.SessionResponse, error)
	UpdateSession(ctx context.Context, id uuid.UUID, req *dto.UpdateSessionRequest) (*dto.SessionResponse, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	StartSession(ctx context.Context, id uuid.UUID) (*dto.SessionResponse, error)
	EndSession(ctx context.Context, id uuid.UUID) (*dto.SessionResponse, error)
	SessionQRCode(ctx context.Context, id uuid.UUID) (*dto.QRCodeResponse, []byte, error)

	CreateInterview(ctx context.Context, operator uuid.UUID, req *dto.CreateInterviewRequest) (*dto.InterviewResponse, error)
	ListInterviews(ctx context.Context, viewer uuid.UUID, req *dto.ListInterviewRequest, scope *rbacModel.DataScopeCondition) ([]*dto.InterviewResponse, int64, error)
	GetInterview(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.InterviewResponse, error)
	AssignEvaluators(ctx context.Context, id uuid.UUID, req *dto.AssignEvaluatorsRequest) (*dto.InterviewResponse, error)
	Checkin(ctx context.Context, userID, id uuid.UUID, token string) (*dto.InterviewResponse, error)
	StartInterview(ctx context.Context, operator, id uuid.UUID) (*dto.InterviewResponse, error)
	EndInterview(ctx context.Context, operator, id uuid.UUID) (*dto.InterviewResponse, error)
	MyInterviews(ctx context.Context, userID uuid.UUID, status *int16) ([]*dto.InterviewResponse, error)

	SubmitEvaluations(ctx context.Context, evaluator, id uuid.UUID, req *dto.SubmitEvaluationsRequest) (*dto.EvaluationSummary, error)
	GetEvaluations(ctx context.Context, id uuid.UUID) (*dto.EvaluationSummary, error)
	UpdateEvaluation(ctx context.Context, evaluator, interviewID, eid uuid.UUID, req *dto.UpdateEvaluationRequest) (*dto.EvaluationResponse, error)
	SubmitResult(ctx context.Context, operator, id uuid.UUID, req *dto.SubmitResultRequest) (*dto.InterviewResponse, error)

	ListDimensions(ctx context.Context) ([]dto.DimensionResponse, error)
	CreateDimension(ctx context.Context, req *dto.CreateDimensionRequest) (*dto.DimensionResponse, error)
	UpdateDimension(ctx context.Context, id uuid.UUID, req *dto.UpdateDimensionRequest) (*dto.DimensionResponse, error)
	DeleteDimension(ctx context.Context, id uuid.UUID) error
	Stats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error)
}

type interviewService struct {
	sessions repo.SessionRepo
	records  repo.InterviewRepo
	evals    repo.EvaluationRepo
	notify   Notifier
	syncer   ApplicationSyncer
}

func NewInterviewService(
	sessions repo.SessionRepo,
	records repo.InterviewRepo,
	evals repo.EvaluationRepo,
	notify Notifier,
	syncer ApplicationSyncer,
) InterviewService {
	return &interviewService{
		sessions: sessions,
		records:  records,
		evals:    evals,
		notify:   notify,
		syncer:   syncer,
	}
}
