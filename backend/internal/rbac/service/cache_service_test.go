package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newCacheSvc(t *testing.T, perm repo.PermissionRepo) (*permissionCacheService, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	svc := NewPermissionCacheService(nil, rdb, perm, nil).(*permissionCacheService)
	return svc, mr
}

func TestGetUserPermissions_CacheMissAndHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPerm := repo.NewMockPermissionRepo(ctrl)
	svc, _ := newCacheSvc(t, mockPerm)
	uid := uuid.New()
	ctx := context.Background()

	mockPerm.EXPECT().GetPermissionCodesByUserID(gomock.Any(), uid).Return([]string{"user:read"}, nil)
	codes, err := svc.GetUserPermissions(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, []string{"user:read"}, codes)

	codes, err = svc.GetUserPermissions(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, []string{"user:read"}, codes)
}

func TestInvalidateUserPermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPerm := repo.NewMockPermissionRepo(ctrl)
	svc, mr := newCacheSvc(t, mockPerm)
	uid := uuid.New()
	ctx := context.Background()
	require.NoError(t, mr.Set(svc.permCacheKey(uid), `["x"]`))
	require.NoError(t, mr.Set(svc.superAdminCacheKey(uid), "1"))

	require.NoError(t, svc.InvalidateUserPermissions(ctx, uid))
	assert.False(t, mr.Exists(svc.permCacheKey(uid)))
	assert.False(t, mr.Exists(svc.superAdminCacheKey(uid)))
}

func TestIsSuperAdmin_CacheHit(t *testing.T) {
	svc, mr := newCacheSvc(t, nil)
	uid := uuid.New()
	require.NoError(t, mr.Set(svc.superAdminCacheKey(uid), "1"))
	ok, err := svc.IsSuperAdmin(context.Background(), uid)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestGetUserPermissionsAndSuperAdmin_BothCached(t *testing.T) {
	svc, mr := newCacheSvc(t, nil)
	uid := uuid.New()
	require.NoError(t, mr.Set(svc.permCacheKey(uid), `["a","b"]`))
	require.NoError(t, mr.Set(svc.superAdminCacheKey(uid), "0"))
	codes, isSuper, err := svc.GetUserPermissionsAndSuperAdmin(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, codes)
	assert.False(t, isSuper)
}

func TestJitteredTTL_NonNegative(t *testing.T) {
	for i := 0; i < 20; i++ {
		got := jitteredTTL(0, time.Millisecond)
		assert.GreaterOrEqual(t, got, time.Duration(0))
	}
}
