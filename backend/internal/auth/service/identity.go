package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"github.com/google/uuid"
)

// MemberIdentity 登录用户关联的会员档案身份。
type MemberIdentity struct {
	StudentNo      string
	RealName       string
	Grade          string
	Major          string
	DepartmentID   string
	DepartmentName string
	PositionID     string
	PositionName   string
}

// MemberIdentityLookup 从人员档案解析学号与姓名等身份信息。
// 实现放在 member 模块，auth 不直接依赖 member。
type MemberIdentityLookup interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*MemberIdentity, error)
	GetUserIDByStudentNo(ctx context.Context, studentNo string) (uuid.UUID, error)
	EnsureStudentNo(ctx context.Context, userID uuid.UUID, studentNo, realName string) error
}

func (s *authService) authenticateLogin(ctx context.Context, identifier, password string) (*model.User, error) {
	identifier = strings.TrimSpace(identifier)
	if err := s.rejectIfLockedOut(ctx, identifier); err != nil {
		return nil, err
	}

	user, err := s.resolveLoginUser(ctx, identifier)
	if err != nil {
		return nil, err
	}

	// 用户名与学号对应同一账号，锁定计数共用 canonical username。
	lockoutKey := identifier
	if user != nil {
		lockoutKey = user.Username
	}
	if lockoutKey != identifier {
		if err := s.rejectIfLockedOut(ctx, lockoutKey); err != nil {
			return nil, err
		}
	}

	if user == nil || !utils.CheckPassword(password, user.PasswordHash) {
		s.recordFailedAttempt(ctx, lockoutKey)
		return nil, response.NewError(response.CodeInvalidCredentials, "用户名/学号或密码错误")
	}
	if user.Status == 1 {
		return nil, response.NewError(response.CodeUserDisabled, "账号已被禁用")
	}
	if user.Status == 2 {
		return nil, response.NewError(response.CodeUserLocked, "账号已被锁定，请联系管理员")
	}

	s.resetLoginAttempts(ctx, lockoutKey, identifier)
	return user, nil
}

func (s *authService) rejectIfLockedOut(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	locked, err := s.authRepo.IsLockedOut(ctx, key)
	if err != nil {
		return fmt.Errorf("check lockout: %w", err)
	}
	if !locked {
		return nil
	}
	ttl, _ := s.authRepo.GetLockoutTTL(ctx, key)
	return response.NewError(response.CodeAccountLocked,
		fmt.Sprintf("登录失败次数过多，账号已被锁定，请 %d 分钟后重试", int(ttl.Minutes())+1))
}

func (s *authService) resetLoginAttempts(ctx context.Context, keys ...string) {
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		_ = s.authRepo.ResetLoginAttempts(ctx, key)
	}
}

func (s *authService) resolveLoginUser(ctx context.Context, identifier string) (*model.User, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, nil
	}
	user, err := s.userRepo.GetByUsername(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user != nil {
		return user, nil
	}
	if s.identity == nil {
		return nil, nil
	}
	userID, err := s.identity.GetUserIDByStudentNo(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("lookup student no: %w", err)
	}
	if userID == uuid.Nil {
		return nil, nil
	}
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (s *authService) buildUserInfo(ctx context.Context, user *model.User, roles, permissions []string) *dto.UserInfo {
	registeredOnly := true
	for _, role := range roles {
		if role != "user" {
			registeredOnly = false
		}
	}
	if registeredOnly {
		permissions = []string{"announcement:read"}
	}
	if roles == nil {
		roles = []string{}
	}
	if permissions == nil {
		permissions = []string{}
	}
	info := &dto.UserInfo{
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
	if user.DepartmentID != nil {
		info.DepartmentID = user.DepartmentID.String()
	}
	if user.PositionID != nil {
		info.PositionID = user.PositionID.String()
	}
	s.attachIdentity(ctx, user.ID, info)
	return info
}

func (s *authService) attachIdentity(ctx context.Context, userID uuid.UUID, info *dto.UserInfo) {
	if s.identity == nil || info == nil {
		return
	}
	ident, err := s.identity.GetByUserID(ctx, userID)
	if err != nil || ident == nil {
		return
	}
	info.StudentNo = ident.StudentNo
	info.Grade = ident.Grade
	info.Major = ident.Major
	info.DepartmentName = ident.DepartmentName
	info.PositionName = ident.PositionName
	if ident.DepartmentID != "" {
		info.DepartmentID = ident.DepartmentID
	}
	if ident.PositionID != "" {
		info.PositionID = ident.PositionID
	}
	if ident.RealName != "" {
		info.RealName = ident.RealName
	}
}
