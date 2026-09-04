package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockAppRepo struct{ mock.Mock }

func (m *mockAppRepo) Create(ctx context.Context, app *model.MemberApplication) error {
	return m.Called(ctx, app).Error(0)
}
func (m *mockAppRepo) Update(ctx context.Context, app *model.MemberApplication) error {
	return m.Called(ctx, app).Error(0)
}
func (m *mockAppRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.MemberApplication, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberApplication), args.Error(1)
}
func (m *mockAppRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ApplicationWithNames, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ApplicationWithNames), args.Error(1)
}
func (m *mockAppRepo) List(ctx context.Context, req *dto.ListApplicationRequest, scope *rbacModel.DataScopeCondition) ([]model.ApplicationWithNames, int64, error) {
	args := m.Called(ctx, req, scope)
	return args.Get(0).([]model.ApplicationWithNames), args.Get(1).(int64), args.Error(2)
}
func (m *mockAppRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.ApplicationWithNames, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.ApplicationWithNames), args.Error(1)
}
func (m *mockAppRepo) HasOpenApplication(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}
func (m *mockAppRepo) CreateHistory(ctx context.Context, h *model.ApplicationHistory) error {
	return m.Called(ctx, h).Error(0)
}
func (m *mockAppRepo) ListHistory(ctx context.Context, applicationID uuid.UUID) ([]model.ApplicationHistory, error) {
	args := m.Called(ctx, applicationID)
	return args.Get(0).([]model.ApplicationHistory), args.Error(1)
}
func (m *mockAppRepo) ListDepartments(ctx context.Context) ([]model.NamedItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.NamedItem), args.Error(1)
}
func (m *mockAppRepo) Stats(ctx context.Context, start, end, groupBy string) ([]model.StatBucket, error) {
	args := m.Called(ctx, start, end, groupBy)
	return args.Get(0).([]model.StatBucket), args.Error(1)
}

type mockProfRepo struct{ mock.Mock }

func (m *mockProfRepo) Create(ctx context.Context, p *model.MemberProfile) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProfRepo) Update(ctx context.Context, p *model.MemberProfile) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProfRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.MemberProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberProfile), args.Error(1)
}
func (m *mockProfRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.MemberProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberProfile), args.Error(1)
}
func (m *mockProfRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ProfileWithNames, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProfileWithNames), args.Error(1)
}
func (m *mockProfRepo) GetByStudentNo(ctx context.Context, studentNo string, excludeID *uuid.UUID) (*model.MemberProfile, error) {
	args := m.Called(ctx, studentNo, excludeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberProfile), args.Error(1)
}
func (m *mockProfRepo) List(ctx context.Context, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]model.ProfileWithNames, int64, error) {
	args := m.Called(ctx, req, scope)
	return args.Get(0).([]model.ProfileWithNames), args.Get(1).(int64), args.Error(2)
}
func (m *mockProfRepo) CreateHistories(ctx context.Context, rows []model.ProfileHistory) error {
	return m.Called(ctx, rows).Error(0)
}
func (m *mockProfRepo) ListHistory(ctx context.Context, profileID uuid.UUID) ([]model.ProfileHistory, error) {
	args := m.Called(ctx, profileID)
	return args.Get(0).([]model.ProfileHistory), args.Error(1)
}
func (m *mockProfRepo) Stats(ctx context.Context, groupBy string) ([]model.StatBucket, error) {
	args := m.Called(ctx, groupBy)
	return args.Get(0).([]model.StatBucket), args.Error(1)
}

type mockStarter struct{ mock.Mock }

func (m *mockStarter) StartOfficerInterview(ctx context.Context, applicationID, initiatorID uuid.UUID, vars map[string]interface{}) (uuid.UUID, error) {
	args := m.Called(ctx, applicationID, initiatorID, vars)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func requireAppError(t mock.TestingT, err error, code int, msg string) {
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
	if msg != "" && app.Message != msg {
		t.Errorf("message %q != %q", app.Message, msg)
	}
}
