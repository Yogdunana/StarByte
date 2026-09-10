package identity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

type mockProfiles struct{ mock.Mock }

func (m *mockProfiles) Create(ctx context.Context, p *model.MemberProfile) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProfiles) Update(ctx context.Context, p *model.MemberProfile) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProfiles) GetByID(ctx context.Context, id uuid.UUID) (*model.MemberProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberProfile), args.Error(1)
}
func (m *mockProfiles) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.MemberProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberProfile), args.Error(1)
}
func (m *mockProfiles) GetByUserIDWithNames(ctx context.Context, userID uuid.UUID) (*model.ProfileWithNames, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProfileWithNames), args.Error(1)
}
func (m *mockProfiles) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ProfileWithNames, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProfileWithNames), args.Error(1)
}
func (m *mockProfiles) GetByStudentNo(ctx context.Context, studentNo string, excludeID *uuid.UUID) (*model.MemberProfile, error) {
	args := m.Called(ctx, studentNo, excludeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberProfile), args.Error(1)
}
func (m *mockProfiles) List(ctx context.Context, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]model.ProfileWithNames, int64, error) {
	args := m.Called(ctx, req, scope)
	return args.Get(0).([]model.ProfileWithNames), args.Get(1).(int64), args.Error(2)
}
func (m *mockProfiles) CreateHistories(ctx context.Context, rows []model.ProfileHistory) error {
	return m.Called(ctx, rows).Error(0)
}
func (m *mockProfiles) ListHistory(ctx context.Context, profileID uuid.UUID) ([]model.ProfileHistory, error) {
	args := m.Called(ctx, profileID)
	return args.Get(0).([]model.ProfileHistory), args.Error(1)
}
func (m *mockProfiles) Stats(ctx context.Context, groupBy string, scope *rbacModel.DataScopeCondition) ([]model.StatBucket, error) {
	args := m.Called(ctx, groupBy, scope)
	return args.Get(0).([]model.StatBucket), args.Error(1)
}

func TestLookup_GetUserIDByStudentNo(t *testing.T) {
	profs := &mockProfiles{}
	lookup := NewLookup(profs)
	ctx := context.Background()
	userID := uuid.New()
	profs.On("GetByStudentNo", ctx, "20210001", (*uuid.UUID)(nil)).
		Return(&model.MemberProfile{UserID: userID, StudentNo: "20210001"}, nil)

	got, err := lookup.GetUserIDByStudentNo(ctx, "20210001")
	assert.NoError(t, err)
	assert.Equal(t, userID, got)

	profs.On("GetByStudentNo", ctx, "missing", (*uuid.UUID)(nil)).Return((*model.MemberProfile)(nil), nil)
	got, err = lookup.GetUserIDByStudentNo(ctx, "missing")
	assert.NoError(t, err)
	assert.Equal(t, uuid.Nil, got)
}

func TestLookup_GetByUserID(t *testing.T) {
	profs := &mockProfiles{}
	lookup := NewLookup(profs)
	ctx := context.Background()
	userID := uuid.New()
	deptID := uuid.New()
	profs.On("GetByUserIDWithNames", ctx, userID).Return(&model.ProfileWithNames{
		MemberProfile: model.MemberProfile{
			UserID: userID, RealName: "管理员", StudentNo: "20210001",
			Grade: "2021", Major: "计算机科学与技术", DepartmentID: &deptID,
		},
		DepartmentName: "项目开发部",
		PositionName:   "社长",
	}, nil)

	ident, err := lookup.GetByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, "20210001", ident.StudentNo)
	assert.Equal(t, "管理员", ident.RealName)
	assert.Equal(t, deptID.String(), ident.DepartmentID)
	assert.Equal(t, "项目开发部", ident.DepartmentName)
}
