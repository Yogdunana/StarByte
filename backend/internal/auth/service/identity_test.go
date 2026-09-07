package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockIdentityLookup struct {
	mock.Mock
}

func (m *mockIdentityLookup) GetByUserID(ctx context.Context, userID uuid.UUID) (*MemberIdentity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MemberIdentity), args.Error(1)
}

func (m *mockIdentityLookup) GetUserIDByStudentNo(ctx context.Context, studentNo string) (uuid.UUID, error) {
	args := m.Called(ctx, studentNo)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func TestLogin_ByStudentNo(t *testing.T) {
	svc, userRepo, authRepo, permCache := setupTestService()
	ident := &mockIdentityLookup{}
	svc.identity = ident
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		RealName:     "测试用户",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       0,
	}

	authRepo.On("IsLockedOut", ctx, "20210002").Return(false, nil)
	userRepo.On("GetByUsername", ctx, "20210002").Return((*model.User)(nil), nil)
	ident.On("GetUserIDByStudentNo", ctx, "20210002").Return(userID, nil)
	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	authRepo.On("ResetLoginAttempts", ctx, "20210002").Return(nil)
	permCache.On("GetUserPermissionsAndSuperAdmin", ctx, userID).Return([]string{"user:read"}, false, nil)
	permCache.On("GetUserRoleCodes", ctx, userID).Return([]string{"member"}, nil)
	authRepo.On("StoreRefreshToken", ctx, mock.Anything, userID.String(), mock.Anything, mock.Anything).Return(nil)
	authRepo.On("StoreSession", ctx, userID.String(), mock.Anything, "127.0.0.1", mock.Anything, mock.Anything).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, userID, "127.0.0.1").Return(nil)
	ident.On("GetByUserID", ctx, userID).Return(&MemberIdentity{
		StudentNo: "20210002", RealName: "测试会员", Grade: "2021", Major: "软件工程",
		DepartmentName: "品牌传播部", PositionName: "干事",
	}, nil)

	result, err := svc.Login(ctx, &dto.LoginRequest{
		Username: "20210002",
		Password: "password123",
	}, "127.0.0.1", "test-agent")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.User)
	assert.Equal(t, "testuser", result.User.Username)
	assert.Equal(t, "测试会员", result.User.RealName)
	assert.Equal(t, "20210002", result.User.StudentNo)
	assert.Equal(t, "2021", result.User.Grade)
	assert.Equal(t, "软件工程", result.User.Major)
	assert.Equal(t, "品牌传播部", result.User.DepartmentName)
}

func TestGetCurrentUser_WithIdentity(t *testing.T) {
	svc, userRepo, _, permCache := setupTestService()
	ident := &mockIdentityLookup{}
	svc.identity = ident
	ctx := context.Background()
	userID := uuid.New()
	deptID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "admin",
		RealName:     "账号名",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       0,
		DepartmentID: &deptID,
	}

	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	permCache.On("GetUserPermissionsAndSuperAdmin", ctx, userID).Return([]string{"*"}, true, nil)
	ident.On("GetByUserID", ctx, userID).Return(&MemberIdentity{
		StudentNo: "20210001", RealName: "管理员", Grade: "2021", Major: "计算机科学与技术",
		DepartmentID: deptID.String(), DepartmentName: "项目开发部", PositionName: "社长",
	}, nil)

	result, err := svc.GetCurrentUser(ctx, userID.String())

	assert.NoError(t, err)
	assert.Equal(t, "20210001", result.StudentNo)
	assert.Equal(t, "管理员", result.RealName)
	assert.Equal(t, "2021", result.Grade)
	assert.Equal(t, "计算机科学与技术", result.Major)
	assert.Equal(t, "项目开发部", result.DepartmentName)
	assert.Equal(t, "社长", result.PositionName)
	assert.Equal(t, deptID.String(), result.DepartmentID)
}

func TestGetCurrentUser_IdentityMissing(t *testing.T) {
	svc, userRepo, _, permCache := setupTestService()
	ident := &mockIdentityLookup{}
	svc.identity = ident
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "fresh",
		RealName:     "新用户",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       0,
	}

	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	permCache.On("GetUserPermissionsAndSuperAdmin", ctx, userID).Return([]string{"user:read"}, false, nil)
	permCache.On("GetUserRoleCodes", ctx, userID).Return([]string{"member"}, nil)
	ident.On("GetByUserID", ctx, userID).Return((*MemberIdentity)(nil), nil)

	result, err := svc.GetCurrentUser(ctx, userID.String())

	assert.NoError(t, err)
	assert.Equal(t, "新用户", result.RealName)
	assert.Empty(t, result.StudentNo)
	assert.Empty(t, result.Grade)
	assert.Empty(t, result.Major)
}
