package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSvc struct{ mock.Mock }

func (m *mockSvc) CreateSession(ctx context.Context, operator uuid.UUID, req *dto.CreateSessionRequest) (*dto.SessionResponse, error) {
	args := m.Called(ctx, operator, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}
func (m *mockSvc) ListSessions(ctx context.Context, viewer uuid.UUID, req *dto.ListSessionRequest, scope *rbacModel.DataScopeCondition) ([]*dto.SessionResponse, int64, error) {
	args := m.Called(ctx, viewer, req, scope)
	return args.Get(0).([]*dto.SessionResponse), args.Get(1).(int64), args.Error(2)
}
func (m *mockSvc) GetSession(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.SessionResponse, error) {
	args := m.Called(ctx, id, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}
func (m *mockSvc) UpdateSession(ctx context.Context, id uuid.UUID, req *dto.UpdateSessionRequest) (*dto.SessionResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}
func (m *mockSvc) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSvc) StartSession(ctx context.Context, id uuid.UUID) (*dto.SessionResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}
func (m *mockSvc) EndSession(ctx context.Context, id uuid.UUID) (*dto.SessionResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}
func (m *mockSvc) SessionQRCode(ctx context.Context, id uuid.UUID) (*dto.QRCodeResponse, []byte, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*dto.QRCodeResponse), args.Get(1).([]byte), args.Error(2)
}
func (m *mockSvc) CreateInterview(ctx context.Context, operator uuid.UUID, req *dto.CreateInterviewRequest) (*dto.InterviewResponse, error) {
	args := m.Called(ctx, operator, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) ListInterviews(ctx context.Context, viewer uuid.UUID, req *dto.ListInterviewRequest, scope *rbacModel.DataScopeCondition) ([]*dto.InterviewResponse, int64, error) {
	args := m.Called(ctx, viewer, req, scope)
	return args.Get(0).([]*dto.InterviewResponse), args.Get(1).(int64), args.Error(2)
}
func (m *mockSvc) GetInterview(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.InterviewResponse, error) {
	args := m.Called(ctx, id, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) AssignEvaluators(ctx context.Context, id uuid.UUID, req *dto.AssignEvaluatorsRequest) (*dto.InterviewResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) Checkin(ctx context.Context, userID, id uuid.UUID, token string) (*dto.InterviewResponse, error) {
	args := m.Called(ctx, userID, id, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) StartInterview(ctx context.Context, operator, id uuid.UUID) (*dto.InterviewResponse, error) {
	args := m.Called(ctx, operator, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) EndInterview(ctx context.Context, operator, id uuid.UUID) (*dto.InterviewResponse, error) {
	args := m.Called(ctx, operator, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) MyInterviews(ctx context.Context, userID uuid.UUID, status *int16) ([]*dto.InterviewResponse, error) {
	args := m.Called(ctx, userID, status)
	return args.Get(0).([]*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) SubmitEvaluations(ctx context.Context, evaluator, id uuid.UUID, req *dto.SubmitEvaluationsRequest) (*dto.EvaluationSummary, error) {
	args := m.Called(ctx, evaluator, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EvaluationSummary), args.Error(1)
}
func (m *mockSvc) GetEvaluations(ctx context.Context, id uuid.UUID) (*dto.EvaluationSummary, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EvaluationSummary), args.Error(1)
}
func (m *mockSvc) UpdateEvaluation(ctx context.Context, evaluator, interviewID, eid uuid.UUID, req *dto.UpdateEvaluationRequest) (*dto.EvaluationResponse, error) {
	args := m.Called(ctx, evaluator, interviewID, eid, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EvaluationResponse), args.Error(1)
}
func (m *mockSvc) SubmitResult(ctx context.Context, operator, id uuid.UUID, req *dto.SubmitResultRequest) (*dto.InterviewResponse, error) {
	args := m.Called(ctx, operator, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InterviewResponse), args.Error(1)
}
func (m *mockSvc) ListDimensions(ctx context.Context) ([]dto.DimensionResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]dto.DimensionResponse), args.Error(1)
}
func (m *mockSvc) CreateDimension(ctx context.Context, req *dto.CreateDimensionRequest) (*dto.DimensionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DimensionResponse), args.Error(1)
}
func (m *mockSvc) UpdateDimension(ctx context.Context, id uuid.UUID, req *dto.UpdateDimensionRequest) (*dto.DimensionResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DimensionResponse), args.Error(1)
}
func (m *mockSvc) DeleteDimension(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSvc) Stats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error) {
	args := m.Called(ctx, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StatsResponse), args.Error(1)
}

func withUser(uid uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", uid.String())
		c.Next()
	}
}

func TestCreateSession_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockSvc{}
	h := NewInterviewHandler(svc)
	uid := uuid.New()
	r := gin.New()
	r.POST("/interviews/sessions", withUser(uid), h.CreateSession)
	svc.On("CreateSession", mock.Anything, uid, mock.AnythingOfType("*dto.CreateSessionRequest")).
		Return(&dto.SessionResponse{Title: "一面", Status: 0}, nil)
	now := time.Now()
	body, _ := json.Marshal(dto.CreateSessionRequest{
		Title: "一面", Round: 1, MaxCandidates: 10, StartTime: now, EndTime: now.Add(time.Hour),
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/interviews/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetInterview_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockSvc{}
	h := NewInterviewHandler(svc)
	id := uuid.New()
	r := gin.New()
	r.GET("/interviews/:id", withUser(uuid.New()), h.GetInterview)
	svc.On("GetInterview", mock.Anything, id, mock.Anything).
		Return(nil, response.NewError(response.CodeInterviewRecordGone, "面试记录不存在"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/interviews/"+id.String(), nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var env response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, response.CodeInterviewRecordGone, env.Code)
}

func TestMyInterviews_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockSvc{}
	h := NewInterviewHandler(svc)
	uid := uuid.New()
	r := gin.New()
	r.GET("/interviews/my", withUser(uid), h.MyInterviews)
	svc.On("MyInterviews", mock.Anything, uid, mock.Anything).Return([]*dto.InterviewResponse{}, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/interviews/my", nil))
	require.Equal(t, http.StatusOK, w.Code)
}
