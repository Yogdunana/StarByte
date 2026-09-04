package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/user/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// 认证相关
	Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error)
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

func (s *userService) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	uid, err := parseRequiredUUID(userID, "用户ID")
	if err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, uid)
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
	uid, err := parseRequiredUUID(userID, "用户ID")
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, uid)
}

func (s *userService) List(ctx context.Context, req *dto.ListUserRequest) ([]dto.UserListResponse, int64, error) {
	var deptID uuid.UUID
	if req.DepartmentID != "" {
		parsed, err := parseRequiredUUID(req.DepartmentID, "部门ID")
		if err != nil {
			return nil, 0, err
		}
		deptID = parsed
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
	deptID, err := parseOptionalUUID(req.DepartmentID, "部门ID")
	if err != nil {
		return nil, err
	}
	user.DepartmentID = deptID
	posID, err := parseOptionalUUID(req.PositionID, "职位ID")
	if err != nil {
		return nil, err
	}
	user.PositionID = posID

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
	deptID, parseErr := parseOptionalUUID(req.DepartmentID, "部门ID")
	if parseErr != nil {
		return nil, parseErr
	}
	if deptID != nil {
		user.DepartmentID = deptID
	}
	posID, parseErr := parseOptionalUUID(req.PositionID, "职位ID")
	if parseErr != nil {
		return nil, parseErr
	}
	if posID != nil {
		user.PositionID = posID
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
	uid, err := parseRequiredUUID(userID, "用户ID")
	if err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, uid)
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
