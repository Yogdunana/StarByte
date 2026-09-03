package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mockUserRepo mocks the user repo for testing.
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

// mockAuthRepo mocks the auth repo for testing.
type mockAuthRepo struct {
	mock.Mock
}

func (m *mockAuthRepo) StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	args := m.Called(ctx, token, userID, ttl)
	return args.Error(0)
}

func (m *mockAuthRepo) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *mockAuthRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAuthRepo) BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	args := m.Called(ctx, tokenID, ttl)
	return args.Error(0)
}

func (m *mockAuthRepo) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	args := m.Called(ctx, tokenID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAuthRepo) IncrLoginAttempts(ctx context.Context, username string) (int64, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockAuthRepo) GetLoginAttempts(ctx context.Context, username string) (int64, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockAuthRepo) ResetLoginAttempts(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

func (m *mockAuthRepo) SetLockout(ctx context.Context, username string, duration time.Duration) error {
	args := m.Called(ctx, username, duration)
	return args.Error(0)
}

func (m *mockAuthRepo) IsLockedOut(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *mockAuthRepo) GetLockoutTTL(ctx context.Context, username string) (time.Duration, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *mockAuthRepo) StoreSession(ctx context.Context, userID, tokenID, ip, userAgent string, ttl time.Duration) error {
	args := m.Called(ctx, userID, tokenID, ip, userAgent, ttl)
	return args.Error(0)
}

func (m *mockAuthRepo) DeleteSession(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *mockAuthRepo) GenerateRefreshToken() string {
	return "test-refresh-token-uuid"
}

// mockPermCache mocks the permission cache service.
type mockPermCache struct {
	mock.Mock
}

func (m *mockPermCache) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockPermCache) InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockPermCache) InvalidateRolePermissions(ctx context.Context, roleID uuid.UUID) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

func (m *mockPermCache) IsSuperAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockPermCache) GetUserPermissionsAndSuperAdmin(ctx context.Context, userID uuid.UUID) ([]string, bool, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Bool(1), args.Error(2)
}

// setup creates a test service with mocks
func setupTestService() (*authService, *mockUserRepo, *mockAuthRepo, *mockPermCache) {
	jwtConfig := &config.JWTConfig{
		Secret:          "test-secret-key-for-testing-only",
		AccessTokenExp:  7200,
		RefreshTokenExp: 604800,
		Issuer:          "starbyte-test",
	}
	userRepo := &mockUserRepo{}
	authRepo := &mockAuthRepo{}
	permCache := &mockPermCache{}
	svc := &authService{
		authRepo:     authRepo,
		userRepo:     userRepo,
		jwtConfig:    jwtConfig,
		permCacheSvc: permCache,
	}
	_ = zap.NewNop() // suppress unused import
	return svc, userRepo, authRepo, permCache
}

// hashPasswordForTest creates a password hash for test users
func hashPasswordForTest(password string) string {
	hash, _ := utils.HashPassword(password)
	return hash
}

func TestLogin_Success(t *testing.T) {
	svc, userRepo, authRepo, permCache := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       0,
	}

	authRepo.On("IsLockedOut", ctx, "testuser").Return(false, nil)
	userRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
	authRepo.On("ResetLoginAttempts", ctx, "testuser").Return(nil)
	permCache.On("GetUserPermissionsAndSuperAdmin", ctx, userID).Return([]string{"user:read"}, false, nil)
	authRepo.On("StoreRefreshToken", ctx, mock.Anything, userID.String(), mock.Anything).Return(nil)
	authRepo.On("StoreSession", ctx, userID.String(), mock.Anything, "127.0.0.1", mock.Anything, mock.Anything).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, userID, "127.0.0.1").Return(nil)

	result, err := svc.Login(ctx, &dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}, "127.0.0.1", "test-agent")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, int64(7200), result.ExpiresIn)
	assert.NotNil(t, result.User)
	assert.Equal(t, "testuser", result.User.Username)
	assert.Equal(t, []string{"user:read"}, result.User.Permissions)
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()

	authRepo.On("IsLockedOut", ctx, "nonexistent").Return(false, nil)
	userRepo.On("GetByUsername", ctx, "nonexistent").Return((*model.User)(nil), nil)
	authRepo.On("IncrLoginAttempts", ctx, "nonexistent").Return(int64(1), nil)

	result, err := svc.Login(ctx, &dto.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}, "127.0.0.1", "test-agent")

	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeInvalidCredentials, appErr.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: hashPasswordForTest("correctpass123"),
		Status:       0,
	}

	authRepo.On("IsLockedOut", ctx, "testuser").Return(false, nil)
	userRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
	authRepo.On("IncrLoginAttempts", ctx, "testuser").Return(int64(1), nil)

	result, err := svc.Login(ctx, &dto.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}, "127.0.0.1", "test-agent")

	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeInvalidCredentials, appErr.Code)
}

func TestLogin_DisabledUser(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "disabled",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       1, // disabled
	}

	authRepo.On("IsLockedOut", ctx, "disabled").Return(false, nil)
	userRepo.On("GetByUsername", ctx, "disabled").Return(user, nil)

	result, err := svc.Login(ctx, &dto.LoginRequest{
		Username: "disabled",
		Password: "password123",
	}, "127.0.0.1", "test-agent")

	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeUserDisabled, appErr.Code)
}

func TestLogin_LockedOutUser(t *testing.T) {
	svc, _, authRepo, _ := setupTestService()
	ctx := context.Background()

	authRepo.On("IsLockedOut", ctx, "lockeduser").Return(true, nil)
	authRepo.On("GetLockoutTTL", ctx, "lockeduser").Return(10 * time.Minute, nil)

	result, err := svc.Login(ctx, &dto.LoginRequest{
		Username: "lockeduser",
		Password: "password123",
	}, "127.0.0.1", "test-agent")

	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeAccountLocked, appErr.Code)
}

func TestLogin_LockoutTriggeredAfter5Attempts(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()

	// Simulate 5th failed attempt
	authRepo.On("IsLockedOut", ctx, "testuser").Return(false, nil)
	userRepo.On("GetByUsername", ctx, "testuser").Return((*model.User)(nil), nil)
	authRepo.On("IncrLoginAttempts", ctx, "testuser").Return(int64(5), nil)
	authRepo.On("SetLockout", ctx, "testuser", mock.Anything).Return(nil)

	_, err := svc.Login(ctx, &dto.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}, "127.0.0.1", "test-agent")

	assert.Error(t, err)
	authRepo.AssertCalled(t, "SetLockout", ctx, "testuser", mock.Anything)
}

func TestRefreshToken_Success(t *testing.T) {
	svc, userRepo, authRepo, permCache := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       0,
	}

	authRepo.On("GetRefreshTokenUserID", ctx, "valid-refresh-token").Return(userID.String(), nil)
	authRepo.On("DeleteRefreshToken", ctx, "valid-refresh-token").Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	permCache.On("GetUserPermissionsAndSuperAdmin", ctx, userID).Return([]string{"user:read"}, false, nil)
	authRepo.On("StoreRefreshToken", ctx, mock.Anything, userID.String(), mock.Anything).Return(nil)

	result, err := svc.RefreshToken(ctx, &dto.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, int64(7200), result.ExpiresIn)
	// Verify old token was deleted (rotation)
	authRepo.AssertCalled(t, "DeleteRefreshToken", ctx, "valid-refresh-token")
	// Verify new token is different
	assert.NotEqual(t, "valid-refresh-token", result.RefreshToken)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc, _, authRepo, _ := setupTestService()
	ctx := context.Background()

	// Redis returns nil for invalid token
	authRepo.On("GetRefreshTokenUserID", ctx, "invalid-token").Return("", errors.New("redis: nil"))

	result, err := svc.RefreshToken(ctx, &dto.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeRefreshTokenInvalid, appErr.Code)
}

func TestRefreshToken_UserDisabled(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       1, // disabled
	}

	authRepo.On("GetRefreshTokenUserID", ctx, "valid-refresh-token").Return(userID.String(), nil)
	authRepo.On("DeleteRefreshToken", ctx, "valid-refresh-token").Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(user, nil)

	result, err := svc.RefreshToken(ctx, &dto.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeUserDisabled, appErr.Code)
}

func TestLogout_Success(t *testing.T) {
	svc, _, authRepo, _ := setupTestService()
	ctx := context.Background()

	authRepo.On("BlacklistToken", ctx, "token-jti-123", mock.Anything).Return(nil)
	authRepo.On("DeleteRefreshToken", ctx, "refresh-token-123").Return(nil)
	authRepo.On("DeleteSession", ctx, "token-jti-123").Return(nil)

	err := svc.Logout(ctx, "user-123", "token-jti-123", "refresh-token-123")
	assert.NoError(t, err)

	authRepo.AssertCalled(t, "BlacklistToken", ctx, "token-jti-123", mock.Anything)
	authRepo.AssertCalled(t, "DeleteRefreshToken", ctx, "refresh-token-123")
	authRepo.AssertCalled(t, "DeleteSession", ctx, "token-jti-123")
}

func TestChangePassword_Success(t *testing.T) {
	svc, userRepo, _, _ := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: hashPasswordForTest("oldpass123"),
		Status:       0,
	}

	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.Anything, mock.Anything).Return(nil)

	err := svc.ChangePassword(ctx, userID.String(), &dto.ChangePasswordRequest{
		OldPassword: "oldpass123",
		NewPassword: "newpass456",
	})

	assert.NoError(t, err)
}

func TestChangePassword_WeakPassword(t *testing.T) {
	svc, _, _, _ := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	err := svc.ChangePassword(ctx, userID.String(), &dto.ChangePasswordRequest{
		OldPassword: "oldpass123",
		NewPassword: "short",
	})

	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodePasswordTooWeak, appErr.Code)
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	svc, userRepo, _, _ := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: hashPasswordForTest("correctpass123"),
		Status:       0,
	}

	userRepo.On("GetByID", ctx, userID).Return(user, nil)

	err := svc.ChangePassword(ctx, userID.String(), &dto.ChangePasswordRequest{
		OldPassword: "wrongoldpass",
		NewPassword: "newpass456",
	})

	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeOldPasswordWrong, appErr.Code)
}

func TestGetCurrentUser_Success(t *testing.T) {
	svc, userRepo, _, permCache := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "testuser",
		RealName:     "Test User",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       0,
	}

	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	permCache.On("GetUserPermissionsAndSuperAdmin", ctx, userID).Return([]string{"user:read", "user:write"}, false, nil)

	result, err := svc.GetCurrentUser(ctx, userID.String())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, "Test User", result.RealName)
	assert.Equal(t, []string{"user:read", "user:write"}, result.Permissions)
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	svc, userRepo, _, _ := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	userRepo.On("GetByID", ctx, userID).Return((*model.User)(nil), nil)

	result, err := svc.GetCurrentUser(ctx, userID.String())

	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeUserNotFound, appErr.Code)
}

func TestGetCurrentUser_SuperAdmin(t *testing.T) {
	svc, userRepo, _, permCache := setupTestService()
	ctx := context.Background()
	userID := uuid.New()

	user := &model.User{
		ID:           userID,
		Username:     "admin",
		RealName:     "Admin",
		PasswordHash: hashPasswordForTest("password123"),
		Status:       0,
	}

	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	permCache.On("GetUserPermissionsAndSuperAdmin", ctx, userID).Return([]string{}, true, nil)

	result, err := svc.GetCurrentUser(ctx, userID.String())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{"*"}, result.Permissions)
	assert.Equal(t, []string{"super_admin"}, result.Roles)
}
