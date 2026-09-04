package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/user/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, tx *gorm.DB, user *model.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, tx *gorm.DB, user *model.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepo) List(ctx context.Context, page, pageSize int, keyword string, status *int, deptID uuid.UUID) ([]model.User, int64, error) {
	args := m.Called(ctx, page, pageSize, keyword, status, deptID)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	args := m.Called(ctx, id, ip)
	return args.Error(0)
}

func newTestUserService(repo *mockUserRepo) UserService {
	return NewUserService(nil, repo, &config.JWTConfig{Secret: "test"})
}

func requireAppError(t *testing.T, err error, code int) {
	t.Helper()
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, code, appErr.Code)
}

func TestParseRequiredUUID_Invalid(t *testing.T) {
	id, err := parseRequiredUUID("not-a-uuid", "用户ID")
	assert.Equal(t, uuid.Nil, id)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestParseRequiredUUID_Valid(t *testing.T) {
	want := uuid.New()
	got, err := parseRequiredUUID(want.String(), "用户ID")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestParseOptionalUUID_Empty(t *testing.T) {
	got, err := parseOptionalUUID("", "部门ID")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestParseOptionalUUID_Invalid(t *testing.T) {
	got, err := parseOptionalUUID("bad-id", "部门ID")
	assert.Nil(t, got)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestGetCurrentUser_InvalidUUID(t *testing.T) {
	svc := newTestUserService(&mockUserRepo{})
	result, err := svc.GetCurrentUser(context.Background(), "not-a-uuid")
	assert.Nil(t, result)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestChangePassword_InvalidUUID(t *testing.T) {
	svc := newTestUserService(&mockUserRepo{})
	err := svc.ChangePassword(context.Background(), "%%%", &dto.ChangePasswordRequest{
		OldPassword: "oldpass123",
		NewPassword: "newpass456",
	})
	requireAppError(t, err, response.CodeBadRequest)
}

func TestUpdateProfile_InvalidUUID(t *testing.T) {
	svc := newTestUserService(&mockUserRepo{})
	err := svc.UpdateProfile(context.Background(), "invalid", &dto.UpdateProfileRequest{RealName: "张三"})
	requireAppError(t, err, response.CodeBadRequest)
}

func TestList_InvalidDepartmentID(t *testing.T) {
	svc := newTestUserService(&mockUserRepo{})
	list, total, err := svc.List(context.Background(), &dto.ListUserRequest{
		Page:         1,
		PageSize:     10,
		DepartmentID: "not-a-uuid",
	})
	assert.Nil(t, list)
	assert.Equal(t, int64(0), total)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestCreate_InvalidDepartmentID(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newTestUserService(repo)
	repo.On("GetByUsername", mock.Anything, "alice").Return((*model.User)(nil), nil)

	result, err := svc.Create(context.Background(), &dto.CreateUserRequest{
		Username:     "alice",
		Password:     "pass1234",
		DepartmentID: "bad-dept",
	})
	assert.Nil(t, result)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestCreate_InvalidPositionID(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newTestUserService(repo)
	repo.On("GetByUsername", mock.Anything, "alice").Return((*model.User)(nil), nil)

	result, err := svc.Create(context.Background(), &dto.CreateUserRequest{
		Username:   "alice",
		Password:   "pass1234",
		PositionID: "bad-pos",
	})
	assert.Nil(t, result)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestUpdate_InvalidDepartmentID(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newTestUserService(repo)
	userID := uuid.New()
	repo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID, Username: "alice"}, nil)

	result, err := svc.Update(context.Background(), userID, &dto.UpdateUserRequest{
		DepartmentID: "not-uuid",
	})
	assert.Nil(t, result)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestUpdate_InvalidPositionID(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newTestUserService(repo)
	userID := uuid.New()
	repo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID, Username: "alice"}, nil)

	result, err := svc.Update(context.Background(), userID, &dto.UpdateUserRequest{
		PositionID: "not-uuid",
	})
	assert.Nil(t, result)
	requireAppError(t, err, response.CodeBadRequest)
}

func TestGetCurrentUser_Success(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newTestUserService(repo)
	userID := uuid.New()
	repo.On("GetByID", mock.Anything, userID).Return(&model.User{
		ID:       userID,
		Username: "alice",
		RealName: "Alice",
	}, nil)

	result, err := svc.GetCurrentUser(context.Background(), userID.String())
	assert.NoError(t, err)
	assert.Equal(t, "alice", result.Username)
}
