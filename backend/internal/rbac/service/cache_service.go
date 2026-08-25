package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 权限缓存相关常量
const (
	permCacheKeyPrefix = "rbac:perms:"
	permCacheTTL       = 10 * time.Minute
	permCacheJitter    = 1 * time.Minute
	superAdminRoleCode = "super_admin"
)

// jitteredTTL 返回在 base TTL 基础上叠加 ±jitter 随机偏移的 TTL，
// 用于避免大量缓存同时失效引发缓存雪崩
func jitteredTTL(base, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return base
	}
	offset := time.Duration(rand.Int63n(int64(jitter*2))) - jitter
	result := base + offset
	if result < 0 {
		return 0
	}
	return result
}

// PermissionCacheService 权限缓存服务接口
type PermissionCacheService interface {
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error
	InvalidateRolePermissions(ctx context.Context, roleID uuid.UUID) error
	IsSuperAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

type permissionCacheService struct {
	db             *gorm.DB
	redisClient    *redis.Client
	permissionRepo repo.PermissionRepo
	roleRepo       repo.RoleRepo
}

// NewPermissionCacheService 创建权限缓存服务
func NewPermissionCacheService(db *gorm.DB, redisClient *redis.Client, permissionRepo repo.PermissionRepo, roleRepo repo.RoleRepo) PermissionCacheService {
	return &permissionCacheService{
		db:             db,
		redisClient:    redisClient,
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
	}
}

// permCacheKey 构造用户权限缓存键
func (s *permissionCacheService) permCacheKey(userID uuid.UUID) string {
	return permCacheKeyPrefix + userID.String()
}

// GetUserPermissions 获取用户权限码列表，优先读取 Redis 缓存，
// 缓存未命中时查询数据库并回填缓存（10 分钟 TTL）
func (s *permissionCacheService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	key := s.permCacheKey(userID)

	// 1. 优先从 Redis 缓存读取
	cached, err := s.redisClient.Get(ctx, key).Result()
	if err == nil && cached != "" {
		var codes []string
		if jsonErr := json.Unmarshal([]byte(cached), &codes); jsonErr == nil {
			logger.Info("rbac permission cache hit",
				zap.String("user_id", userID.String()))
			return codes, nil
		}
		// 反序列化失败则继续回源查询
	}

	// 2. 缓存未命中，查询数据库（一次 JOIN 查询直接返回权限编码）
	logger.Info("rbac permission cache miss, querying DB",
		zap.String("user_id", userID.String()))

	codes, err := s.permissionRepo.GetPermissionCodesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user permission codes: %w", err)
	}

	// 3. 回填缓存（即使为空也缓存，防止缓存穿透）
	data, err := json.Marshal(codes)
	if err != nil {
		logger.Error("marshal permission codes failed",
			zap.String("user_id", userID.String()), zap.Error(err))
	} else if setErr := s.redisClient.Set(ctx, key, string(data), jitteredTTL(permCacheTTL, permCacheJitter)).Err(); setErr != nil {
		logger.Warn("set permission cache failed",
			zap.String("user_id", userID.String()), zap.Error(setErr))
	}

	return codes, nil
}

// InvalidateUserPermissions 失效指定用户的权限缓存
func (s *permissionCacheService) InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error {
	key := s.permCacheKey(userID)
	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete permission cache: %w", err)
	}
	return nil
}

// InvalidateRolePermissions 失效指定角色下所有用户的权限缓存
func (s *permissionCacheService) InvalidateRolePermissions(ctx context.Context, roleID uuid.UUID) error {
	// 查询该角色下所有有效用户 ID
	var userIDs []uuid.UUID
	err := s.db.WithContext(ctx).
		Model(&model.UserRole{}).
		Where("role_id = ? AND (expired_at IS NULL OR expired_at > NOW())", roleID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return fmt.Errorf("get user ids by role: %w", err)
	}

	// 逐个失效用户缓存，单条失败仅记录日志不中断流程
	for _, uid := range userIDs {
		key := s.permCacheKey(uid)
		if delErr := s.redisClient.Del(ctx, key).Err(); delErr != nil {
			logger.Warn("invalidate user permission cache failed",
				zap.String("user_id", uid.String()),
				zap.String("role_id", roleID.String()),
				zap.Error(delErr))
		}
	}

	return nil
}

// IsSuperAdmin 判断用户是否拥有 super_admin 角色
// 注意：super_admin 角色本身必须是启用状态（status=0），否则不视为超级管理员
func (s *permissionCacheService) IsSuperAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	// 查询用户拥有的角色 ID 列表（已过滤已禁用角色）
	roleIDs, err := s.roleRepo.GetRoleIDsByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("get user role ids: %w", err)
	}
	if len(roleIDs) == 0 {
		return false, nil
	}

	// 查询 super_admin 角色是否存在且启用
	role, err := s.roleRepo.GetByCode(ctx, superAdminRoleCode)
	if err != nil {
		return false, fmt.Errorf("get super_admin role: %w", err)
	}
	if role == nil {
		return false, nil
	}

	// super_admin 角色若被禁用，则不授予超级管理员权限
	if role.Status != 0 {
		return false, nil
	}

	// 判断用户角色列表中是否包含 super_admin
	for _, rid := range roleIDs {
		if rid == role.ID {
			return true, nil
		}
	}

	return false, nil
}
