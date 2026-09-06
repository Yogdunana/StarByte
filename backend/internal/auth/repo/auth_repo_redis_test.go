package repo

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) (*authRepo, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return &authRepo{rdb: rdb}, mr
}

func TestSessionIndex_StoreListDelete(t *testing.T) {
	r, _ := newTestRepo(t)
	ctx := context.Background()

	require.NoError(t, r.StoreSession(ctx, "u1", "j1", "1.1.1.1", "Chrome UA", time.Hour))
	require.NoError(t, r.StoreSession(ctx, "u1", "j2", "2.2.2.2", "Safari UA", time.Hour))
	require.NoError(t, r.StoreSession(ctx, "u2", "j3", "3.3.3.3", "Firefox UA", time.Hour))

	got, err := r.GetSession(ctx, "j1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u1", got.UserID)
	assert.Equal(t, "1.1.1.1", got.IP)

	all, err := r.ListSessions(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	mine, err := r.ListSessionsByUser(ctx, "u1")
	require.NoError(t, err)
	assert.Len(t, mine, 2)

	require.NoError(t, r.DeleteSession(ctx, "j1"))
	gone, err := r.GetSession(ctx, "j1")
	require.NoError(t, err)
	assert.Nil(t, gone)

	mine, err = r.ListSessionsByUser(ctx, "u1")
	require.NoError(t, err)
	assert.Len(t, mine, 1)
}

func TestSessionIndex_ScanFallback(t *testing.T) {
	r, mr := newTestRepo(t)
	ctx := context.Background()
	require.NoError(t, r.StoreSession(ctx, "u9", "legacy", "8.8.8.8", "ua", time.Hour))
	mr.Del("auth:user_sessions:u9")

	got, err := r.ListSessionsByUser(ctx, "u9")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "legacy", got[0].TokenID)
}

func TestRefreshToken_JSONAndLegacy(t *testing.T) {
	r, mr := newTestRepo(t)
	ctx := context.Background()

	require.NoError(t, r.StoreRefreshToken(ctx, "tok-new", "u1", "jti-1", time.Hour))
	uid, jti, err := r.GetRefreshTokenMeta(ctx, "tok-new")
	require.NoError(t, err)
	assert.Equal(t, "u1", uid)
	assert.Equal(t, "jti-1", jti)

	uid, err = r.GetRefreshTokenUserID(ctx, "tok-new")
	require.NoError(t, err)
	assert.Equal(t, "u1", uid)

	mr.Set("auth:refresh:tok-old", "plain-user")
	uid, jti, err = r.GetRefreshTokenMeta(ctx, "tok-old")
	require.NoError(t, err)
	assert.Equal(t, "plain-user", uid)
	assert.Equal(t, "", jti)

	require.NoError(t, r.DeleteRefreshToken(ctx, "tok-new"))
	_, _, err = r.GetRefreshTokenMeta(ctx, "tok-new")
	assert.Error(t, err)
}

func TestRefreshToken_DeleteByUserAndJTI(t *testing.T) {
	r, _ := newTestRepo(t)
	ctx := context.Background()
	require.NoError(t, r.StoreRefreshToken(ctx, "a", "u1", "j1", time.Hour))
	require.NoError(t, r.StoreRefreshToken(ctx, "b", "u1", "j2", time.Hour))
	require.NoError(t, r.StoreRefreshToken(ctx, "c", "u2", "j3", time.Hour))

	require.NoError(t, r.DeleteRefreshTokensByJTI(ctx, "u1", "j1"))
	_, _, err := r.GetRefreshTokenMeta(ctx, "a")
	assert.Error(t, err)
	uid, err := r.GetRefreshTokenUserID(ctx, "b")
	require.NoError(t, err)
	assert.Equal(t, "u1", uid)

	require.NoError(t, r.DeleteRefreshTokensByUser(ctx, "u1"))
	_, _, err = r.GetRefreshTokenMeta(ctx, "b")
	assert.Error(t, err)
	uid, err = r.GetRefreshTokenUserID(ctx, "c")
	require.NoError(t, err)
	assert.Equal(t, "u2", uid)
}

func TestBlacklistAndLockout(t *testing.T) {
	r, _ := newTestRepo(t)
	ctx := context.Background()

	ok, err := r.IsBlacklisted(ctx, "j1")
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, r.BlacklistToken(ctx, "j1", time.Minute))
	ok, err = r.IsBlacklisted(ctx, "j1")
	require.NoError(t, err)
	assert.True(t, ok)

	n, err := r.IncrLoginAttempts(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
	n, err = r.GetLoginAttempts(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
	require.NoError(t, r.ResetLoginAttempts(ctx, "alice"))
	n, err = r.GetLoginAttempts(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)

	locked, err := r.IsLockedOut(ctx, "alice")
	require.NoError(t, err)
	assert.False(t, locked)
	require.NoError(t, r.SetLockout(ctx, "alice", time.Minute))
	locked, err = r.IsLockedOut(ctx, "alice")
	require.NoError(t, err)
	assert.True(t, locked)
	ttl, err := r.GetLockoutTTL(ctx, "alice")
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
}

func TestGetSession_Empty(t *testing.T) {
	r, _ := newTestRepo(t)
	got, err := r.GetSession(context.Background(), "")
	require.NoError(t, err)
	assert.Nil(t, got)
	got, err = r.GetSession(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, got)
}
