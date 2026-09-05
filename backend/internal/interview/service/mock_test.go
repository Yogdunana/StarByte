package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Create(ctx context.Context, s *model.Session) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockSessionRepo) Update(ctx context.Context, s *model.Session) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}
func (m *mockSessionRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.SessionWithNames, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SessionWithNames), args.Error(1)
}
func (m *mockSessionRepo) List(ctx context.Context, req *dto.ListSessionRequest, scope *rbacModel.DataScopeCondition) ([]model.SessionWithNames, int64, error) {
	args := m.Called(ctx, req, scope)
	return args.Get(0).([]model.SessionWithNames), args.Get(1).(int64), args.Error(2)
}
func (m *mockSessionRepo) CountCandidates(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(int64), args.Error(1)
}

type mockInterviewRepo struct{ mock.Mock }

func (m *mockInterviewRepo) Create(ctx context.Context, iv *model.Interview) error {
	return m.Called(ctx, iv).Error(0)
}
func (m *mockInterviewRepo) Update(ctx context.Context, iv *model.Interview) error {
	return m.Called(ctx, iv).Error(0)
}
func (m *mockInterviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Interview, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Interview), args.Error(1)
}
func (m *mockInterviewRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.InterviewWithNames, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InterviewWithNames), args.Error(1)
}
func (m *mockInterviewRepo) List(ctx context.Context, req *dto.ListInterviewRequest, scope *rbacModel.DataScopeCondition) ([]model.InterviewWithNames, int64, error) {
	args := m.Called(ctx, req, scope)
	return args.Get(0).([]model.InterviewWithNames), args.Get(1).(int64), args.Error(2)
}
func (m *mockInterviewRepo) ListMine(ctx context.Context, userID uuid.UUID, status *int16) ([]model.InterviewWithNames, error) {
	args := m.Called(ctx, userID, status)
	return args.Get(0).([]model.InterviewWithNames), args.Error(1)
}
func (m *mockInterviewRepo) FindBySessionApplicant(ctx context.Context, sessionID, applicantID uuid.UUID) (*model.Interview, error) {
	args := m.Called(ctx, sessionID, applicantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Interview), args.Error(1)
}
func (m *mockInterviewRepo) ReplaceInterviewers(ctx context.Context, interviewID uuid.UUID, rows []model.Interviewer) error {
	return m.Called(ctx, interviewID, rows).Error(0)
}
func (m *mockInterviewRepo) ListInterviewers(ctx context.Context, interviewIDs []uuid.UUID) ([]model.InterviewerNamed, error) {
	args := m.Called(ctx, interviewIDs)
	return args.Get(0).([]model.InterviewerNamed), args.Error(1)
}
func (m *mockInterviewRepo) HasInterviewerConflict(ctx context.Context, interviewerID uuid.UUID, start, end time.Time, exclude uuid.UUID) (bool, error) {
	args := m.Called(ctx, interviewerID, start, end, exclude)
	return args.Bool(0), args.Error(1)
}
func (m *mockInterviewRepo) MarkAbsentBySession(ctx context.Context, sessionID uuid.UUID) error {
	return m.Called(ctx, sessionID).Error(0)
}
func (m *mockInterviewRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.NamedUser), args.Error(1)
}
func (m *mockInterviewRepo) GetApplication(ctx context.Context, id uuid.UUID) (*model.ApplicationBrief, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ApplicationBrief), args.Error(1)
}

type mockEvalRepo struct{ mock.Mock }

func (m *mockEvalRepo) CreateBatch(ctx context.Context, rows []model.Evaluation) error {
	return m.Called(ctx, rows).Error(0)
}
func (m *mockEvalRepo) Update(ctx context.Context, ev *model.Evaluation) error {
	return m.Called(ctx, ev).Error(0)
}
func (m *mockEvalRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Evaluation), args.Error(1)
}
func (m *mockEvalRepo) ListByInterview(ctx context.Context, interviewID uuid.UUID) ([]model.EvaluationNamed, error) {
	args := m.Called(ctx, interviewID)
	return args.Get(0).([]model.EvaluationNamed), args.Error(1)
}
func (m *mockEvalRepo) HasEvaluatorScores(ctx context.Context, interviewID, evaluatorID uuid.UUID) (bool, error) {
	args := m.Called(ctx, interviewID, evaluatorID)
	return args.Bool(0), args.Error(1)
}
func (m *mockEvalRepo) ListDimensions(ctx context.Context) ([]model.Dimension, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Dimension), args.Error(1)
}
func (m *mockEvalRepo) GetDimensionByName(ctx context.Context, name string) (*model.Dimension, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Dimension), args.Error(1)
}
func (m *mockEvalRepo) GetDimensionByID(ctx context.Context, id uuid.UUID) (*model.Dimension, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Dimension), args.Error(1)
}
func (m *mockEvalRepo) CreateDimension(ctx context.Context, d *model.Dimension) error {
	return m.Called(ctx, d).Error(0)
}
func (m *mockEvalRepo) UpdateDimension(ctx context.Context, d *model.Dimension) error {
	return m.Called(ctx, d).Error(0)
}
func (m *mockEvalRepo) DeleteDimension(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockEvalRepo) Stats(ctx context.Context, q *dto.StatsQuery) (model.StatsRow, []model.ScoreBucket, []model.DeptStat, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(model.StatsRow), args.Get(1).([]model.ScoreBucket), args.Get(2).([]model.DeptStat), args.Error(3)
}

type mockNotify struct{ mock.Mock }

func (m *mockNotify) Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error {
	return m.Called(ctx, userIDs, template, vars).Error(0)
}

type mockSyncer struct{ mock.Mock }

func (m *mockSyncer) SyncFromInterview(ctx context.Context, operator, applicationID uuid.UUID, result int16, comment string) error {
	return m.Called(ctx, operator, applicationID, result, comment).Error(0)
}

func requireAppError(t mock.TestingT, err error, code int) {
	if err == nil {
		t.Errorf("expected error")
		return
	}
	app, ok := err.(*response.AppError)
	if !ok {
		t.Errorf("expected AppError, got %T %v", err, err)
		return
	}
	if app.Code != code {
		t.Errorf("code %d != %d", app.Code, code)
	}
}

func namedSession(s *model.Session) *model.SessionWithNames {
	return &model.SessionWithNames{Session: *s}
}

func namedInterview(iv *model.Interview, name string) *model.InterviewWithNames {
	return &model.InterviewWithNames{Interview: *iv, ApplicantName: name}
}
