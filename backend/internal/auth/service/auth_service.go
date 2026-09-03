package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/auth/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	userRepo "github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"github.com/google/uuid"
)

// AuthService defines the authentication service interface.
type AuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshResponse, error)
	Logout(ctx context.Context, userID, tokenID, refreshToken string) error
	GetCurrentUser(ctx context.Context, userID string) (*dto.UserInfo, error)
	ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error
}

type authService struct {
	authRepo     repo.AuthRepo
	userRepo     userRepo.UserRepo
	jwtConfig    *config.JWTConfig
	permCacheSvc rbacService.PermissionCacheService
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	authRepo repo.AuthRepo,
	userRepo userRepo.UserRepo,
	jwtConfig *config.JWTConfig,
	permCacheSvc rbacService.PermissionCacheService,
) AuthService {
	return &authService{
		authRepo:     authRepo,
		userRepo:     userRepo,
		jwtConfig:    jwtConfig,
		permCacheSvc: permCacheSvc,
	}
}

// Login authenticates a user and returns an access token + refresh token pair.
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error) {
	// 1. Check lockout
	locked, err := s.authRepo.IsLockedOut(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("check lockout: %w", err)
	}
	if locked {
		ttl, _ := s.authRepo.GetLockoutTTL(ctx, req.Username)
		return nil, response.NewError(response.CodeAccountLocked,
			fmt.Sprintf("登录失败次数过多，账号已被锁定，请 %d 分钟后重试", int(ttl.Minutes())+1))
	}

	// 2. Query user
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 3. Validate credentials
	if user == nil || !utils.CheckPassword(req.Password, user.PasswordHash) {
		s.recordFailedAttempt(ctx, req.Username)
		return nil, response.NewError(response.CodeInvalidCredentials, "用户名或密码错误")
	}

	// 4. Check user status
	if user.Status == 1 {
		return nil, response.NewError(response.CodeUserDisabled, "账号已被禁用")
	}
	if user.Status == 2 {
		return nil, response.NewError(response.CodeUserLocked, "账号已被锁定，请联系管理员")
	}

	// 5. Reset failed attempts on success
	_ = s.authRepo.ResetLoginAttempts(ctx, req.Username)

	// 6. Get roles and permissions
	userUUID, _ := uuid.Parse(user.ID.String())
	roles, permissions := s.getUserRolesAndPermissions(ctx, userUUID)

	// 7. Generate access token
	accessToken, _, err := authmiddleware.GenerateAccessToken(
		user.ID.String(), user.Username, roles, permissions, s.jwtConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// 8. Generate and store refresh token (random string in Redis)
	refreshToken := s.authRepo.GenerateRefreshToken()
	refreshTTL := time.Duration(s.jwtConfig.RefreshTokenExp) * time.Second
	if err := s.authRepo.StoreRefreshToken(ctx, refreshToken, user.ID.String(), refreshTTL); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	// 9. Store session metadata
	claims, _ := authmiddleware.ParseToken(accessToken, s.jwtConfig)
	if claims != nil && claims.ID != "" {
		_ = s.authRepo.StoreSession(ctx, user.ID.String(), claims.ID, ip, userAgent, accessTTL(s.jwtConfig))
	}

	// 10. Update last login info
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, ip)

	// 11. Build response
	userInfo := buildUserInfo(user, roles, permissions)

	return &dto.LoginResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(s.jwtConfig.AccessTokenExp),
		RefreshExpiresIn: int64(s.jwtConfig.RefreshTokenExp),
		User:             userInfo,
	}, nil
}

// RefreshToken validates a refresh token, rotates it, and returns a new token pair.
func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshResponse, error) {
	// 1. Look up refresh token in Redis
	userID, err := s.authRepo.GetRefreshTokenUserID(ctx, req.RefreshToken)
	if err != nil {
		return nil, response.NewError(response.CodeRefreshTokenInvalid, "Refresh Token 无效")
	}

	// 2. Delete old refresh token (rotation: one-time use)
	if err := s.authRepo.DeleteRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("delete refresh token: %w", err)
	}

	// 3. Query user
	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(userID))
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeUserNotFound, "用户不存在")
	}
	if user.Status != 0 {
		return nil, response.NewError(response.CodeUserDisabled, "用户已被禁用")
	}

	// 4. Get roles and permissions
	userUUID, _ := uuid.Parse(user.ID.String())
	roles, permissions := s.getUserRolesAndPermissions(ctx, userUUID)

	// 5. Generate new access token
	accessToken, _, err := authmiddleware.GenerateAccessToken(
		user.ID.String(), user.Username, roles, permissions, s.jwtConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// 6. Generate and store new refresh token (rotation)
	newRefreshToken := s.authRepo.GenerateRefreshToken()
	refreshTTL := time.Duration(s.jwtConfig.RefreshTokenExp) * time.Second
	if err := s.authRepo.StoreRefreshToken(ctx, newRefreshToken, user.ID.String(), refreshTTL); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessTokenExp),
	}, nil
}

// Logout invalidates the current access token (blacklist) and refresh token.
func (s *authService) Logout(ctx context.Context, userID, tokenID, refreshToken string) error {
	// Blacklist the access token
	if tokenID != "" {
		accessTTL := time.Duration(s.jwtConfig.AccessTokenExp) * time.Second
		_ = s.authRepo.BlacklistToken(ctx, tokenID, accessTTL)
	}

	// Delete the refresh token
	if refreshToken != "" {
		_ = s.authRepo.DeleteRefreshToken(ctx, refreshToken)
	}

	// Delete session metadata
	if tokenID != "" {
		_ = s.authRepo.DeleteSession(ctx, tokenID)
	}

	return nil
}

// GetCurrentUser returns the current user's information including roles and permissions.
func (s *authService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(userID))
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeUserNotFound, "用户不存在")
	}

	userUUID, _ := uuid.Parse(user.ID.String())
	roles, permissions := s.getUserRolesAndPermissions(ctx, userUUID)

	return buildUserInfo(user, roles, permissions), nil
}

// ChangePassword validates the old password, checks new password strength, and updates.
func (s *authService) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	// 1. Validate new password strength
	if !utils.ValidatePasswordStrength(req.NewPassword) {
		return response.NewError(response.CodePasswordTooWeak,
			"密码强度不足：至少 8 位，需包含字母和数字")
	}

	// 2. Query user
	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(userID))
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return response.NewError(response.CodeUserNotFound, "用户不存在")
	}

	// 3. Verify old password
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return response.NewError(response.CodeOldPasswordWrong, "原密码错误")
	}

	// 4. Hash new password
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.PasswordHash = newHash
	return s.userRepo.Update(ctx, nil, user)
}

// ========== Helper methods ==========

// recordFailedAttempt increments the failed login counter and locks out if threshold reached.
func (s *authService) recordFailedAttempt(ctx context.Context, username string) {
	attempts, err := s.authRepo.IncrLoginAttempts(ctx, username)
	if err != nil {
		return
	}
	if attempts >= int64(repo.MaxLoginAttempts()) {
		_ = s.authRepo.SetLockout(ctx, username, repo.LockoutDuration())
	}
}

// getUserRolesAndPermissions fetches roles and permissions via the permission cache service.
func (s *authService) getUserRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, []string) {
	permissions, isSuperAdmin, err := s.permCacheSvc.GetUserPermissionsAndSuperAdmin(ctx, userID)
	if err != nil || isSuperAdmin {
		return []string{"super_admin"}, []string{"*"}
	}
	// For non-super-admin users, we return role codes as roles.
	// In a production system, we'd fetch role codes here; for now,
	// permissions suffice and the frontend handles role-based UI.
	return nil, permissions
}

// buildUserInfo constructs the UserInfo response from a User model.
func buildUserInfo(user *model.User, roles, permissions []string) *dto.UserInfo {
	return &dto.UserInfo{
		ID:          user.ID.String(),
		Username:    user.Username,
		RealName:    user.RealName,
		AvatarURL:   user.AvatarURL,
		Email:       user.Email,
		Phone:       user.Phone,
		Gender:      user.Gender,
		Status:      user.Status,
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
	}
}

// accessTTL returns the access token TTL duration.
func accessTTL(cfg *config.JWTConfig) time.Duration {
	return time.Duration(cfg.AccessTokenExp) * time.Second
}
