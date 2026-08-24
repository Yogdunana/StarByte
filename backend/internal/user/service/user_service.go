package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/user/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/redis"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// 认证相关
	Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req *dto.LoginRequest, ip string) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	Logout(ctx context.Context, userID string, tokenID string) error
	ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error

	// 用户管理
	GetByID(ctx context.Context, id uuid.UUID) (*dto.UserInfoResponse, error)
	GetCurrentUser(ctx context.Context, userID string) (*dto.UserInfoResponse, error)
	List(ctx context.Context, req *dto.ListUserRequest) ([]dto.UserListResponse, int64, error)
	Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfoResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserInfoResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) error
}

type userService struct {
	db        *gorm.DB
	userRepo  repo.UserRepo
	jwtConfig *config.JWTConfig
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, userRepo repo.UserRepo, jwtConfig *config.JWTConfig) UserService {
	return &userService{
		db:        db,
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// ========== 认证相关 ==========

func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	existing, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, response.NewError(2001, "用户名已存在")
	}

	// 密码加密
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		RealName:     req.RealName,
		Email:        req.Email,
		Phone:        req.Phone,
		Status:       0,
	}

	// 创建用户（默认分配 member 角色，这里简化处理）
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, user); err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, req *dto.LoginRequest, ip string) (*dto.LoginResponse, error) {
	// 查询用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(2002, "用户名或密码错误")
	}

	// 检查用户状态
	if user.Status == 1 {
		return nil, response.NewError(2003, "账号已被禁用")
	}
	if user.Status == 2 {
		return nil, response.NewError(2004, "账号已被锁定")
	}

	// 校验密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, response.NewError(2002, "用户名或密码错误")
	}

	// 生成 Token
	tokenPair, err := authmiddleware.GenerateTokenPair(user.ID.String(), user.Username, s.jwtConfig)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 更新最后登录信息
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, ip)

	// 构建用户信息
	userInfo := &dto.UserInfoResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		RealName:  user.RealName,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Phone:     user.Phone,
		Gender:    user.Gender,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	return &dto.LoginResponse{
		AccessToken:         tokenPair.AccessToken,
		RefreshToken:        tokenPair.RefreshToken,
		AccessTokenExpires:  tokenPair.AccessTokenExpires.Unix(),
		RefreshTokenExpires: tokenPair.RefreshTokenExpires.Unix(),
		User:                userInfo,
	}, nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	// 解析 Refresh Token
	claims, err := authmiddleware.ParseToken(refreshToken, s.jwtConfig)
	if err != nil {
		return nil, response.NewError(2005, "Refresh Token 无效或已过期")
	}

	// 检查 Token 类型
	if claims.TokenType != authmiddleware.RefreshTokenType {
		return nil, response.NewError(2006, "Token 类型错误")
	}

	// 检查是否在黑名单中
	blacklistKey := "jwt:blacklist:" + claims.ID
	exists, _ := redis.Exists(ctx, blacklistKey)
	if exists {
		return nil, response.NewError(2007, "Token 已失效")
	}

	// 查询用户
	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(claims.UserID))
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil || user.Status != 0 {
		return nil, response.NewError(2008, "用户不存在或已被禁用")
	}

	// 生成新的 Token 对
	tokenPair, err := authmiddleware.GenerateTokenPair(user.ID.String(), user.Username, s.jwtConfig)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 将旧的 Refresh Token 加入黑名单
	ttl := int(time.Until(claims.ExpiresAt.Time).Seconds())
	if ttl > 0 {
		_ = redis.Get().Set(ctx, blacklistKey, "1", 0).Err()
	}

	userInfo := &dto.UserInfoResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		RealName:  user.RealName,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Phone:     user.Phone,
		Gender:    user.Gender,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	return &dto.LoginResponse{
		AccessToken:         tokenPair.AccessToken,
		RefreshToken:        tokenPair.RefreshToken,
		AccessTokenExpires:  tokenPair.AccessTokenExpires.Unix(),
		RefreshTokenExpires: tokenPair.RefreshTokenExpires.Unix(),
		User:                userInfo,
	}, nil
}

func (s *userService) Logout(ctx context.Context, userID string, tokenID string) error {
	// 将 Access Token 加入黑名单
	if tokenID != "" {
		key := "jwt:blacklist:" + tokenID
		_ = redis.Get().Set(ctx, key, "1", 0).Err()
	}
	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(userID))
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return response.NewError(response.CodeNotFound, "用户不存在")
	}

	// 校验旧密码
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return response.NewError(2009, "原密码错误")
	}

	// 加密新密码
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.PasswordHash = newHash
	return s.userRepo.Update(ctx, nil, user)
}

// ========== 用户管理 ==========

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserInfoResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeNotFound, "用户不存在")
	}

	return &dto.UserInfoResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		RealName:  user.RealName,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Phone:     user.Phone,
		Gender:    user.Gender,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *userService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserInfoResponse, error) {
	return s.GetByID(ctx, uuid.MustParse(userID))
}

func (s *userService) List(ctx context.Context, req *dto.ListUserRequest) ([]dto.UserListResponse, int64, error) {
	var deptID uuid.UUID
	if req.DepartmentID != "" {
		deptID = uuid.MustParse(req.DepartmentID)
	}

	users, total, err := s.userRepo.List(ctx, req.Page, req.PageSize, req.Keyword, req.Status, deptID)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	result := make([]dto.UserListResponse, 0, len(users))
	for _, user := range users {
		item := dto.UserListResponse{
			ID:          user.ID.String(),
			Username:    user.Username,
			RealName:    user.RealName,
			AvatarURL:   user.AvatarURL,
			Email:       user.Email,
			Phone:       user.Phone,
			Gender:      user.Gender,
			Status:      user.Status,
			LastLoginAt: formatTimePtr(user.LastLoginAt),
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		}
		if user.DepartmentID != nil {
			item.DepartmentID = user.DepartmentID.String()
		}
		if user.PositionID != nil {
			item.PositionID = user.PositionID.String()
		}
		result = append(result, item)
	}

	return result, total, nil
}

func (s *userService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfoResponse, error) {
	// 检查用户名
	existing, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, response.NewError(2001, "用户名已存在")
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		RealName:     req.RealName,
		Email:        req.Email,
		Phone:        req.Phone,
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.DepartmentID != "" {
		deptID := uuid.MustParse(req.DepartmentID)
		user.DepartmentID = &deptID
	}
	if req.PositionID != "" {
		posID := uuid.MustParse(req.PositionID)
		user.PositionID = &posID
	}

	err = s.userRepo.Create(ctx, nil, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.GetByID(ctx, user.ID)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserInfoResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeNotFound, "用户不存在")
	}

	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.DepartmentID != "" {
		deptID := uuid.MustParse(req.DepartmentID)
		user.DepartmentID = &deptID
	}
	if req.PositionID != "" {
		posID := uuid.MustParse(req.PositionID)
		user.PositionID = &posID
	}

	err = s.userRepo.Update(ctx, nil, user)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return s.GetByID(ctx, id)
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return response.NewError(response.CodeNotFound, "用户不存在")
	}

	return s.userRepo.Delete(ctx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(userID))
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return response.NewError(response.CodeNotFound, "用户不存在")
	}

	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}

	return s.userRepo.Update(ctx, nil, user)
}

// ========== 工具函数 ==========

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// 避免 errors 未使用告警（实际会用）
var _ = errors.New
