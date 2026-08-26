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
	permCacheKeyPrefix       = "rbac:perms:"
	superAdminCacheKeyPrefix = "rbac:superadmin:"
	permCacheTTL             = 10 * time.Minute
	permCacheJitter          = 1 * time.Minute
	superAdminRoleCode       = "super_admin"
)

// rng 是包级别的随机数生成器，使用 time.Now().UnixNano() 作为种子，
// 确保在 Go 1.20 之前的版本中也能获得非确定性的随机序列。
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// jitteredTTL 返回在 base TTL 基础上叠加 ±jitter 随机偏移的 TTL，
// 用于避免大量缓存同时失效引发缓存雪崩
func jitteredTTL(base, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return base
	}
	offset := time.Duration(rng.Int63n(int64(jitter*2))) - jitter
	result := base + offset
	if result < 0 {
		return 0
	}
	return result
}

// PermissionCacheService 权限缓存服务接口
// 基于 Redis 实现用户权限和超级管理员身份的缓存，减少数据库查询压力。
type PermissionCacheService interface {
	// GetUserPermissions 获取用户权限码列表，优先读缓存，未命中回源数据库
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	// InvalidateUserPermissions 失效指定用户的权限缓存（同时失效超级管理员身份缓存）
	InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error
	// InvalidateRolePermissions 失效指定角色下所有用户的权限缓存
	InvalidateRolePermissions(ctx context.Context, roleID uuid.UUID) error
	// IsSuperAdmin 判断用户是否拥有 super_admin 角色（结果缓存）
	IsSuperAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	// GetUserPermissionsAndSuperAdmin 同时获取用户权限列表和超级管理员身份（使用 Pipeline 减少 Redis 往返）
	GetUserPermissionsAndSuperAdmin(ctx context.Context, userID uuid.UUID) ([]string, bool, error)
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
			return codes, nil
		}
		// 反序列化失败则继续回源查询
	}

	// 2. 缓存未命中，查询数据库（一次 JOIN 查询直接返回权限编码）

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

// InvalidateUserPermissions 失效指定用户的权限缓存（同时失效超级管理员身份缓存）
func (s *permissionCacheService) InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error {
	// 失效权限缓存
	permKey := s.permCacheKey(userID)
	if err := s.redisClient.Del(ctx, permKey).Err(); err != nil {
		return fmt.Errorf("delete permission cache: %w", err)
	}
	// 失效超级管理员身份缓存
	superKey := s.superAdminCacheKey(userID)
	if err := s.redisClient.Del(ctx, superKey).Err(); err != nil {
		logger.Warn("delete super admin cache failed",
			zap.String("user_id", userID.String()), zap.Error(err))
	}
	return nil
}

// InvalidateRolePermissions 失效指定角色下所有用户的权限缓存
// 同时失效超级管理员身份缓存，使用 Redis Pipeline 批量删除以提升性能
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

	if len(userIDs) == 0 {
		return nil
	}

	// 使用 Redis Pipeline 批量删除权限缓存和超级管理员身份缓存
	pipe := s.redisClient.Pipeline()
	for _, uid := range userIDs {
		permKey := s.permCacheKey(uid)
		pipe.Del(ctx, permKey)
		superKey := s.superAdminCacheKey(uid)
		pipe.Del(ctx, superKey)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Warn("invalidate role permissions cache pipeline failed",
			zap.String("role_id", roleID.String()),
			zap.Error(err))
		// Pipeline 部分失败时不返回错误，仅记录日志（单条删除失败不影响主流程）
	}

	return nil
}

// superAdminCacheKey 构造超级管理员身份缓存键
func (s *permissionCacheService) superAdminCacheKey(userID uuid.UUID) string {
	return superAdminCacheKeyPrefix + userID.String()
}

// IsSuperAdmin 判断用户是否拥有 super_admin 角色
// 注意：super_admin 角色本身必须是启用状态（status=0），否则不视为超级管理员
// 结果会缓存到 Redis，TTL 与权限缓存相同，避免每次请求查库
func (s *permissionCacheService) IsSuperAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	key := s.superAdminCacheKey(userID)

	// 1. 优先从 Redis 缓存读取
	cached, err := s.redisClient.Get(ctx, key).Result()
	if err == nil && cached != "" {
		return cached == "1", nil
	}

	// 2. 缓存未命中，查询数据库
	isSuper, err := s.checkIsSuperAdmin(ctx, userID)
	if err != nil {
		return false, err
	}

	// 3. 回填缓存（即使为 false 也缓存，防止缓存穿透）
	cacheVal := "0"
	if isSuper {
		cacheVal = "1"
	}
	if setErr := s.redisClient.Set(ctx, key, cacheVal, jitteredTTL(permCacheTTL, permCacheJitter)).Err(); setErr != nil {
		logger.Warn("set super admin cache failed",
			zap.String("user_id", userID.String()), zap.Error(setErr))
	}

	return isSuper, nil
}

// GetUserPermissionsAndSuperAdmin 同时获取用户权限列表和超级管理员身份
// 使用 Redis Pipeline 合并两次 GET 为一次往返，减少网络开销
// 缓存未命中的部分会回源查询数据库并回填缓存
func (s *permissionCacheService) GetUserPermissionsAndSuperAdmin(ctx context.Context, userID uuid.UUID) ([]string, bool, error) {
	permKey := s.permCacheKey(userID)
	superKey := s.superAdminCacheKey(userID)

	// 1. 使用 Pipeline 同时获取两个缓存 key，减少一次 Redis 往返
	pipe := s.redisClient.Pipeline()
	permCmd := pipe.Get(ctx, permKey)
	superCmd := pipe.Get(ctx, superKey)
	_, err := pipe.Exec(ctx)

	// 解析权限缓存
	var perms []string
	permCached := false
	if err == nil || (err != nil && permCmd.Err() == nil) {
		cachedPerms, pErr := permCmd.Result()
		if pErr == nil && cachedPerms != "" {
			if jsonErr := json.Unmarshal([]byte(cachedPerms), &perms); jsonErr == nil {
				permCached = true
			}
		}
	}

	// 解析超级管理员缓存
	var isSuper bool
	superCached := false
	if err == nil || (err != nil && superCmd.Err() == nil) {
		cachedSuper, sErr := superCmd.Result()
		if sErr == nil && cachedSuper != "" {
			isSuper = cachedSuper == "1"
			superCached = true
		}
	}

	// 2. 如果两者都命中缓存，直接返回
	if permCached && superCached {
		return perms, isSuper, nil
	}

	// 3. 缓存未命中的部分回源查询数据库
	if !permCached {
		perms, err = s.permissionRepo.GetPermissionCodesByUserID(ctx, userID)
		if err != nil {
			return nil, false, fmt.Errorf("get user permission codes: %w", err)
		}
		// 回填权限缓存
		data, marshalErr := json.Marshal(perms)
		if marshalErr != nil {
			logger.Error("marshal permission codes failed",
				zap.String("user_id", userID.String()), zap.Error(marshalErr))
		} else if setErr := s.redisClient.Set(ctx, permKey, string(data), jitteredTTL(permCacheTTL, permCacheJitter)).Err(); setErr != nil {
			logger.Warn("set permission cache failed",
				zap.String("user_id", userID.String()), zap.Error(setErr))
		}
	}

	if !superCached {
		isSuper, err = s.checkIsSuperAdmin(ctx, userID)
		if err != nil {
			return nil, false, err
		}
		// 回填超级管理员缓存
		cacheVal := "0"
		if isSuper {
			cacheVal = "1"
		}
		if setErr := s.redisClient.Set(ctx, superKey, cacheVal, jitteredTTL(permCacheTTL, permCacheJitter)).Err(); setErr != nil {
			logger.Warn("set super admin cache failed",
				zap.String("user_id", userID.String()), zap.Error(setErr))
		}
	}

	return perms, isSuper, nil
}

// checkIsSuperAdmin 实际查询数据库判断用户是否为超级管理员
// 使用一次 JOIN 查询直接判断用户是否拥有启用状态的 super_admin 角色
func (s *permissionCacheService) checkIsSuperAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var result bool
	err := s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) > 0
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ?
		  AND r.code = ?
		  AND r.status = 0
		  AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
	`, userID, superAdminRoleCode).Scan(&result).Error
	if err != nil {
		return false, fmt.Errorf("check super admin: %w", err)
	}
	return result, nil
}
