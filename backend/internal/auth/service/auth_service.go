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
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"github.com/google/uuid"
)

// AuthService defines the authentication service interface.
type AuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, ip, userAgent string) (*dto.RefreshResponse, error)
	Logout(ctx context.Context, userID, tokenID, refreshToken string) error
	GetCurrentUser(ctx context.Context, userID string) (*dto.UserInfo, error)
	ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error
	ListSessions(ctx context.Context, keyword, userID string) (*dto.SessionListResponse, error)
	GetUserSessions(ctx context.Context, userID string) (*dto.UserSessionsResponse, error)
	KickSession(ctx context.Context, tokenID string) error
	KickUserSessions(ctx context.Context, userID string) error
	CASStatus() dto.CASStatusResponse
	BuildCASLoginURL(ctx context.Context, redirect, publicOrigin string) (*dto.CASLoginStart, error)
	CompleteCASCallback(ctx context.Context, ticket, state, ip, userAgent, publicOrigin string) (string, error)
	ExchangeCASCode(ctx context.Context, code string) (*dto.CASExchangeResponse, error)
}

type authService struct {
	authRepo     repo.AuthRepo
	userRepo     userRepo.UserRepo
	jwtConfig    *config.JWTConfig
	permCacheSvc rbacService.PermissionCacheService
	eventBus     *events.EventBus
	identity     MemberIdentityLookup
	cas          *config.CASConfig
	casStore     repo.CASTicketStore
	casValidator TicketValidator
	casRole      RoleAssigner
}

// NewAuthService creates a new authentication service.
// identity 可为 nil（无档案时登录仍可用，学号登录与档案字段为空）。
// cas 可为 nil（学校统一认证关闭）。
func NewAuthService(
	authRepo repo.AuthRepo,
	userRepo userRepo.UserRepo,
	jwtConfig *config.JWTConfig,
	permCacheSvc rbacService.PermissionCacheService,
	eventBus *events.EventBus,
	identity MemberIdentityLookup,
	cas *CASDeps,
) AuthService {
	svc := &authService{
		authRepo:     authRepo,
		userRepo:     userRepo,
		jwtConfig:    jwtConfig,
		permCacheSvc: permCacheSvc,
		eventBus:     eventBus,
		identity:     identity,
	}
	if cas != nil {
		svc.cas = cas.Config
		svc.casStore = cas.Store
		svc.casValidator = cas.Validator
		svc.casRole = cas.AssignRole
	}
	return svc
}

// Login authenticates a user and returns an access token + refresh token pair.
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error) {
	user, err := s.authenticateLogin(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return s.issueSession(ctx, user, ip, userAgent)
}

func (s *authService) issueSession(ctx context.Context, user *model.User, ip, userAgent string) (*dto.LoginResponse, error) {
	userUUID, _ := uuid.Parse(user.ID.String())
	roles, permissions, err := s.getUserRolesAndPermissions(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get roles and permissions: %w", err)
	}

	accessToken, _, err := authmiddleware.GenerateAccessToken(
		user.ID.String(), user.Username, roles, permissions, s.jwtConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	claims, _ := authmiddleware.ParseToken(accessToken, s.jwtConfig)
	accessJTI := ""
	if claims != nil {
		accessJTI = claims.ID
	}

	refreshToken := s.authRepo.GenerateRefreshToken()
	refreshTTL := time.Duration(s.jwtConfig.RefreshTokenExp) * time.Second
	if err := s.authRepo.StoreRefreshToken(ctx, refreshToken, user.ID.String(), accessJTI, refreshTTL); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	if accessJTI != "" {
		_ = s.authRepo.StoreSession(ctx, user.ID.String(), accessJTI, ip, userAgent, accessTTL(s.jwtConfig))
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, ip)
	s.publishLogin(ctx, user, ip, userAgent)
	userInfo := s.buildUserInfo(ctx, user, roles, permissions)

	return &dto.LoginResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(s.jwtConfig.AccessTokenExp),
		RefreshExpiresIn: int64(s.jwtConfig.RefreshTokenExp),
		User:             userInfo,
	}, nil
}

// RefreshToken validates a refresh token, rotates it, and returns a new token pair.
func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, ip, userAgent string) (*dto.RefreshResponse, error) {
	userID, oldJTI, err := s.authRepo.GetRefreshTokenMeta(ctx, req.RefreshToken)
	if err != nil {
		return nil, response.NewError(response.CodeRefreshTokenInvalid, "Refresh Token 无效")
	}

	if err := s.authRepo.DeleteRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("delete refresh token: %w", err)
	}
	if oldJTI != "" {
		_ = s.authRepo.BlacklistToken(ctx, oldJTI, accessTTL(s.jwtConfig))
		_ = s.authRepo.DeleteSession(ctx, oldJTI)
	}

	// 3. Query user
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, response.NewError(response.CodeTokenInvalid, "无效的用户标识")
	}
	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeUserNotFound, "用户不存在")
	}
	if user.Status != 0 {
		return nil, response.NewError(response.CodeUserDisabled, "用户已被禁用")
	}

	// 4. Get roles and permissions（缓存/DB 失败必须 fail-closed，不得提权）
	roles, permissions, err := s.getUserRolesAndPermissions(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get roles and permissions: %w", err)
	}

	// 5. Generate new access token
	accessToken, _, err := authmiddleware.GenerateAccessToken(
		user.ID.String(), user.Username, roles, permissions, s.jwtConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newClaims, _ := authmiddleware.ParseToken(accessToken, s.jwtConfig)
	newJTI := ""
	if newClaims != nil {
		newJTI = newClaims.ID
	}

	newRefreshToken := s.authRepo.GenerateRefreshToken()
	refreshTTL := time.Duration(s.jwtConfig.RefreshTokenExp) * time.Second
	if err := s.authRepo.StoreRefreshToken(ctx, newRefreshToken, user.ID.String(), newJTI, refreshTTL); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}
	if newJTI != "" {
		_ = s.authRepo.StoreSession(ctx, user.ID.String(), newJTI, ip, userAgent, accessTTL(s.jwtConfig))
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

	s.publishLogout(ctx, userID)
	return nil
}

// GetCurrentUser returns the current user's information including roles and permissions.
func (s *authService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserInfo, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, response.NewError(response.CodeTokenInvalid, "无效的用户标识")
	}
	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeUserNotFound, "用户不存在")
	}

	roles, permissions, err := s.getUserRolesAndPermissions(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get roles and permissions: %w", err)
	}

	return s.buildUserInfo(ctx, user, roles, permissions), nil
}

// ChangePassword validates the old password, checks new password strength, and updates.
func (s *authService) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	// 1. Validate new password strength
	if !utils.ValidatePasswordStrength(req.NewPassword) {
		return response.NewError(response.CodePasswordTooWeak,
			"密码强度不足：至少 8 位，需包含字母和数字")
	}

	// 2. Query user
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewError(response.CodeTokenInvalid, "无效的用户标识")
	}
	user, err := s.userRepo.GetByID(ctx, userUUID)
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
// Fail-closed: cache/DB errors must not elevate the caller to super_admin / "*".
// Only isSuperAdmin==true && err==nil grants wildcard permissions.
func (s *authService) getUserRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, []string, error) {
	permissions, isSuperAdmin, err := s.permCacheSvc.GetUserPermissionsAndSuperAdmin(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("get user permissions: %w", err)
	}
	if isSuperAdmin {
		return []string{"super_admin"}, []string{"*"}, nil
	}
	// For non-super-admin users, query actual role codes.
	roles, err := s.permCacheSvc.GetUserRoleCodes(ctx, userID)
	if err != nil {
		// Role lookup failure must not invent privileges; keep fetched permissions.
		return nil, permissions, nil
	}
	return roles, permissions, nil
}

// accessTTL returns the access token TTL duration.
func accessTTL(cfg *config.JWTConfig) time.Duration {
	return time.Duration(cfg.AccessTokenExp) * time.Second
}

func (s *authService) publishLogin(ctx context.Context, user *model.User, ip, userAgent string) {
	if s.eventBus == nil || user == nil {
		return
	}
	s.eventBus.Publish(ctx, events.UserLoginEvent{
		UserID:    user.ID,
		Username:  user.Username,
		RealName:  user.RealName,
		IP:        ip,
		UserAgent: userAgent,
	})
}

func (s *authService) publishLogout(ctx context.Context, userID string) {
	if s.eventBus == nil {
		return
	}
	ev := events.UserLogoutEvent{}
	if uid, err := uuid.Parse(userID); err == nil {
		ev.UserID = uid
		if user, err := s.userRepo.GetByID(ctx, uid); err == nil && user != nil {
			ev.Username = user.Username
			ev.RealName = user.RealName
		}
	}
	s.eventBus.Publish(ctx, ev)
}
