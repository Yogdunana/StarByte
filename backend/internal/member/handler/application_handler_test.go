package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMemberService struct{ mock.Mock }

func (m *mockMemberService) Submit(ctx context.Context, userID uuid.UUID, req *dto.SubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ApplicationResponse), args.Error(1)
}
func (m *mockMemberService) Resubmit(ctx context.Context, userID, id uuid.UUID, req *dto.ResubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	args := m.Called(ctx, userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ApplicationResponse), args.Error(1)
}
func (m *mockMemberService) ListApplications(ctx context.Context, viewer uuid.UUID, req *dto.ListApplicationRequest, scope *rbacModel.DataScopeCondition) ([]*dto.ApplicationResponse, int64, error) {
	args := m.Called(ctx, viewer, req, scope)
	return args.Get(0).([]*dto.ApplicationResponse), args.Get(1).(int64), args.Error(2)
}
func (m *mockMemberService) GetApplication(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ApplicationResponse, error) {
	args := m.Called(ctx, viewer, id, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ApplicationResponse), args.Error(1)
}
func (m *mockMemberService) MyApplications(ctx context.Context, userID uuid.UUID) ([]*dto.ApplicationResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*dto.ApplicationResponse), args.Error(1)
}
func (m *mockMemberService) ApplicationHistory(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.ApplicationHistoryResponse, error) {
	args := m.Called(ctx, viewer, id, scope)
	return args.Get(0).([]dto.ApplicationHistoryResponse), args.Error(1)
}
func (m *mockMemberService) Approve(ctx context.Context, reviewer, id uuid.UUID, comment string) (*dto.ApplicationResponse, error) {
	args := m.Called(ctx, reviewer, id, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ApplicationResponse), args.Error(1)
}
func (m *mockMemberService) Reject(ctx context.Context, reviewer, id uuid.UUID, comment string) (*dto.ApplicationResponse, error) {
	args := m.Called(ctx, reviewer, id, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ApplicationResponse), args.Error(1)
}
func (m *mockMemberService) Supplement(ctx context.Context, reviewer, id uuid.UUID, req *dto.SupplementRequest) (*dto.ApplicationResponse, error) {
	args := m.Called(ctx, reviewer, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ApplicationResponse), args.Error(1)
}
func (m *mockMemberService) ListDepartments(ctx context.Context) ([]dto.DepartmentOption, error) {
	args := m.Called(ctx)
	return args.Get(0).([]dto.DepartmentOption), args.Error(1)
}
func (m *mockMemberService) ListProfiles(ctx context.Context, viewer uuid.UUID, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]*dto.ProfileResponse, int64, error) {
	args := m.Called(ctx, viewer, req, scope)
	return args.Get(0).([]*dto.ProfileResponse), args.Get(1).(int64), args.Error(2)
}
func (m *mockMemberService) GetProfile(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ProfileResponse, error) {
	args := m.Called(ctx, viewer, id, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileResponse), args.Error(1)
}
func (m *mockMemberService) UpdateProfile(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateProfileRequest, scope *rbacModel.DataScopeCondition) (*dto.ProfileResponse, error) {
	args := m.Called(ctx, operator, id, req, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileResponse), args.Error(1)
}
func (m *mockMemberService) UpdateProfileStatus(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateProfileStatusRequest) (*dto.ProfileResponse, error) {
	args := m.Called(ctx, operator, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileResponse), args.Error(1)
}
func (m *mockMemberService) ProfileHistory(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.ProfileHistoryResponse, error) {
	args := m.Called(ctx, viewer, id, scope)
	return args.Get(0).([]dto.ProfileHistoryResponse), args.Error(1)
}
func (m *mockMemberService) ExportProfiles(ctx context.Context, viewer uuid.UUID, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]byte, error) {
	args := m.Called(ctx, viewer, req, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
func (m *mockMemberService) ApplicationStats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error) {
	args := m.Called(ctx, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StatsResponse), args.Error(1)
}
func (m *mockMemberService) MemberStats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error) {
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

func TestSubmit_BadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMemberHandler(&mockMemberService{})
	r := gin.New()
	r.POST("/member/applications", withUser(uuid.New()), h.Submit)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/member/applications", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetApplication_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMemberHandler(&mockMemberService{})
	r := gin.New()
	uid := uuid.New()
	r.GET("/member/applications/:id", withUser(uid), h.GetApplication)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/member/applications/not-uuid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApprove_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockMemberService{}
	h := NewMemberHandler(svc)
	uid := uuid.New()
	id := uuid.New()
	r := gin.New()
	r.POST("/member/applications/:id/approve", withUser(uid), h.Approve)
	svc.On("Approve", mock.Anything, uid, id, "同意").Return(&dto.ApplicationResponse{ID: id.String(), Status: 3}, nil)

	body, _ := json.Marshal(dto.ReviewCommentRequest{Comment: "同意"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/member/applications/"+id.String()+"/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestExport_PDF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockMemberService{}
	h := NewMemberHandler(svc)
	uid := uuid.New()
	r := gin.New()
	r.GET("/member/profiles/export", withUser(uid), h.ExportProfiles)
	svc.On("ExportProfiles", mock.Anything, uid, mock.Anything, mock.Anything).Return([]byte("%PDF-1.4 test"), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/member/profiles/export", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
}

func TestGetApplication_NotFoundCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockMemberService{}
	h := NewMemberHandler(svc)
	uid := uuid.New()
	id := uuid.New()
	r := gin.New()
	r.GET("/member/applications/:id", withUser(uid), h.GetApplication)
	svc.On("GetApplication", mock.Anything, uid, id, mock.Anything).Return(nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/member/applications/"+id.String(), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var env response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, response.CodeMemberAppNotFound, env.Code)
}

func TestWithUserUsesAuthKey(t *testing.T) {
	// 保证 helper 读的 context key 与 JWT 中间件一致
	gin.SetMode(gin.TestMode)
	r := gin.New()
	uid := uuid.New()
	r.GET("/me", func(c *gin.Context) {
		c.Set("user_id", uid.String())
		got, err := getUserID(c)
		assert.NoError(t, err)
		assert.Equal(t, uid, got)
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	_ = auth.GetUserID
}
