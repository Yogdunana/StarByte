package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/cache/dto"
	pkgcache "github.com/Yogdunana/StarByte/backend/pkg/cache"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/redis/go-redis/v9"
)

const maxScan = int64(200)
const maxDelete = int64(500)

const authKeyPrefix = "auth:"

// CacheService is the admin API for Redis + L1 cache.
type CacheService interface {
	Stats(ctx context.Context, pattern string) (*dto.Stats, error)
	DeleteKey(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) (*dto.DeleteResult, error)
	Warmup(ctx context.Context, req *dto.WarmupRequest) (*dto.WarmupResult, error)
	Close()
}

type cacheService struct {
	layered *pkgcache.Layered
	cancel  context.CancelFunc
}

func NewCacheService(rdb *redis.Client) CacheService {
	layered := pkgcache.NewLayered(rdb)
	ctx, cancel := context.WithCancel(context.Background())
	_ = layered.L2().SubscribeExpired(ctx, func(key string) {
		layered.L1().Delete(key)
	})
	return &cacheService{layered: layered, cancel: cancel}
}

func (s *cacheService) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *cacheService) Stats(ctx context.Context, pattern string) (*dto.Stats, error) {
	pattern = strings.TrimSpace(pattern)
	out := &dto.Stats{Pattern: pattern, Keys: []dto.KeyInfo{}}
	hits, misses, size := s.layered.Stats()
	out.L1Hits, out.L1Misses, out.L1Size = hits, misses, size

	pool, err := s.layered.L2().Health(ctx)
	if err != nil {
		if pool != nil {
			out.Pool = *pool
		}
		return out, nil
	}
	out.Pool = *pool
	out.Healthy = pool.PingOK

	if pattern == "" || pattern == "*" || pattern == "?" {
		return out, nil
	}
	if err := denySensitivePattern(pattern); err != nil {
		return nil, err
	}
	keys, err := s.layered.L2().ScanKeys(ctx, pattern, maxScan)
	if err != nil {
		return nil, response.NewError(response.CodeCacheRedisDown, "扫描缓存键失败")
	}
	for _, k := range dropSensitiveKeys(keys) {
		ttl, terr := s.layered.L2().TTL(ctx, k)
		info := dto.KeyInfo{Key: k, TTLSeconds: -1}
		if terr == nil {
			if ttl < 0 {
				info.TTLSeconds = -1
			} else {
				info.TTLSeconds = int64(ttl.Seconds())
			}
		}
		out.Keys = append(out.Keys, info)
	}
	out.KeyCount = len(out.Keys)
	return out, nil
}

func (s *cacheService) DeleteKey(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return response.NewError(response.CodeCacheInvalidKey, "缓存键不能为空")
	}
	if isSensitiveKey(key) {
		return response.NewForbiddenError("不能操作认证相关缓存键")
	}
	ok, err := s.layered.L2().Exists(ctx, key)
	if err != nil {
		return response.NewError(response.CodeCacheRedisDown, "检查缓存键失败")
	}
	if err := s.layered.Delete(ctx, key); err != nil {
		return response.NewError(response.CodeCacheRedisDown, "删除缓存失败")
	}
	if !ok {
		return response.NewError(response.CodeCacheKeyNotFound, "缓存键不存在")
	}
	return nil
}

func (s *cacheService) DeletePattern(ctx context.Context, pattern string) (*dto.DeleteResult, error) {
	pattern = strings.TrimSpace(pattern)
	if err := validatePattern(pattern); err != nil {
		return nil, err
	}
	if err := denySensitivePattern(pattern); err != nil {
		return nil, err
	}
	keys, err := s.layered.L2().ScanKeys(ctx, pattern, maxDelete)
	if err != nil {
		return nil, response.NewError(response.CodeCacheRedisDown, "扫描缓存键失败")
	}
	keys = dropSensitiveKeys(keys)
	if len(keys) == 0 {
		return &dto.DeleteResult{Keys: []string{}}, nil
	}
	for _, k := range keys {
		s.layered.L1().Delete(k)
	}
	n, err := s.layered.L2().Del(ctx, keys...)
	if err != nil {
		return nil, response.NewError(response.CodeCacheRedisDown, "按模式删除失败")
	}
	return &dto.DeleteResult{Deleted: n, Keys: keys}, nil
}

func (s *cacheService) Warmup(ctx context.Context, req *dto.WarmupRequest) (*dto.WarmupResult, error) {
	if req == nil {
		return nil, response.NewError(response.CodeCacheWarmupFail, "预热参数不能为空")
	}
	out := &dto.WarmupResult{Keys: []string{}}
	for _, e := range req.Entries {
		key := strings.TrimSpace(e.Key)
		if key == "" {
			return nil, response.NewError(response.CodeCacheInvalidKey, "预热键不能为空")
		}
		if isSensitiveKey(key) {
			return nil, response.NewForbiddenError("不能操作认证相关缓存键")
		}
		ttl := time.Duration(e.TTLSeconds) * time.Second
		if err := s.layered.Set(ctx, key, e.Value, ttl); err != nil {
			return nil, response.NewError(response.CodeCacheWarmupFail, "写入预热键失败")
		}
		out.Keys = append(out.Keys, key)
	}
	prefix := strings.TrimSpace(req.ScanPrefix)
	if prefix != "" {
		if err := validatePattern(prefix); err != nil {
			return nil, err
		}
		if err := denySensitivePattern(prefix); err != nil {
			return nil, err
		}
		keys, err := s.layered.L2().ScanKeys(ctx, prefix, maxScan)
		if err != nil {
			return nil, response.NewError(response.CodeCacheWarmupFail, "扫描预热前缀失败")
		}
		for _, k := range dropSensitiveKeys(keys) {
			v, gerr := s.layered.L2().Get(ctx, k)
			if gerr != nil {
				continue
			}
			ttl, _ := s.layered.L2().TTL(ctx, k)
			s.layered.L1().Set(k, v, ttl)
			out.Keys = append(out.Keys, k)
		}
	}
	if len(out.Keys) == 0 {
		return nil, response.NewError(response.CodeCacheWarmupFail, "没有可预热的键")
	}
	out.Loaded = len(out.Keys)
	return out, nil
}

func validatePattern(pattern string) error {
	if pattern == "" || pattern == "*" || pattern == "?" {
		return response.NewError(response.CodeCacheInvalidPattern, "清除模式过宽，请指定前缀")
	}
	stripped := strings.ReplaceAll(strings.ReplaceAll(pattern, "*", ""), "?", "")
	if len(stripped) < 2 {
		return response.NewError(response.CodeCacheInvalidPattern, "清除模式过宽，请指定前缀")
	}
	return nil
}

func isSensitiveKey(key string) bool {
	return strings.HasPrefix(key, authKeyPrefix)
}

func denySensitivePattern(pattern string) error {
	p := strings.ToLower(strings.TrimSpace(pattern))
	if strings.HasPrefix(p, authKeyPrefix) || strings.Contains(p, "auth:") {
		return response.NewForbiddenError("不能操作认证相关缓存键")
	}
	return nil
}

func dropSensitiveKeys(keys []string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if !isSensitiveKey(k) {
			out = append(out, k)
		}
	}
	return out
}
