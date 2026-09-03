package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// AuthRepo handles all Redis-based authentication data operations.
type AuthRepo interface {
	// StoreRefreshToken stores a refresh token in Redis with the given TTL.
	// The token is stored at key "auth:refresh:{token}" with userID as value.
	StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error

	// GetRefreshTokenUserID retrieves the userID associated with a refresh token.
	// Returns "" and redis.Nil if the token does not exist.
	GetRefreshTokenUserID(ctx context.Context, token string) (string, error)

	// DeleteRefreshToken removes a refresh token from Redis (used during rotation).
	DeleteRefreshToken(ctx context.Context, token string) error

	// BlacklistToken adds a token (access or refresh) to the blacklist with the given TTL.
	BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error

	// IsBlacklisted checks whether a token ID is in the blacklist.
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)

	// IncrLoginAttempts increments the failed login attempt counter for a username.
	IncrLoginAttempts(ctx context.Context, username string) (int64, error)

	// GetLoginAttempts returns the current failed attempt count for a username.
	GetLoginAttempts(ctx context.Context, username string) (int64, error)

	// ResetLoginAttempts clears the failed attempt counter for a username.
	ResetLoginAttempts(ctx context.Context, username string) error

	// SetLockout locks a username for the given duration.
	SetLockout(ctx context.Context, username string, duration time.Duration) error

	// IsLockedOut checks whether a username is currently locked out.
	IsLockedOut(ctx context.Context, username string) (bool, error)

	// GetLockoutTTL returns the remaining lockout time for a username.
	GetLockoutTTL(ctx context.Context, username string) (time.Duration, error)

	// StoreSession stores session metadata for a user.
	StoreSession(ctx context.Context, userID, tokenID, ip, userAgent string, ttl time.Duration) error

	// DeleteSession removes a session by token ID.
	DeleteSession(ctx context.Context, tokenID string) error

	// GenerateRefreshToken generates a cryptographically random refresh token string.
	GenerateRefreshToken() string
}

type authRepo struct {
	rdb *redis.Client
}

// NewAuthRepo creates a new AuthRepo backed by Redis.
func NewAuthRepo(rdb *redis.Client) AuthRepo {
	return &authRepo{rdb: rdb}
}

const (
	keyRefreshToken  = "auth:refresh:%s"
	keyBlacklist     = "auth:blacklist:%s"
	keyLoginAttempts = "auth:login_attempts:%s"
	keyLockout       = "auth:lockout:%s"
	keySession       = "auth:session:%s"

	maxLoginAttempts = 5
	lockoutDuration  = 15 * time.Minute
)

func (r *authRepo) StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	key := fmt.Sprintf(keyRefreshToken, token)
	return r.rdb.Set(ctx, key, userID, ttl).Err()
}

func (r *authRepo) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf(keyRefreshToken, token)
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", redis.Nil
	}
	return val, err
}

func (r *authRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf(keyRefreshToken, token)
	return r.rdb.Del(ctx, key).Err()
}

func (r *authRepo) BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil // no need to blacklist an already-expired token
	}
	key := fmt.Sprintf(keyBlacklist, tokenID)
	return r.rdb.Set(ctx, key, "1", ttl).Err()
}

func (r *authRepo) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return false, nil
	}
	key := fmt.Sprintf(keyBlacklist, tokenID)
	n, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *authRepo) IncrLoginAttempts(ctx context.Context, username string) (int64, error) {
	key := fmt.Sprintf(keyLoginAttempts, username)
 pipe := r.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, lockoutDuration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Result()
}

func (r *authRepo) GetLoginAttempts(ctx context.Context, username string) (int64, error) {
	key := fmt.Sprintf(keyLoginAttempts, username)
	n, err := r.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return n, err
}

func (r *authRepo) ResetLoginAttempts(ctx context.Context, username string) error {
	key := fmt.Sprintf(keyLoginAttempts, username)
	return r.rdb.Del(ctx, key).Err()
}

func (r *authRepo) SetLockout(ctx context.Context, username string, duration time.Duration) error {
	key := fmt.Sprintf(keyLockout, username)
	return r.rdb.Set(ctx, key, "1", duration).Err()
}

func (r *authRepo) IsLockedOut(ctx context.Context, username string) (bool, error) {
	key := fmt.Sprintf(keyLockout, username)
	n, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *authRepo) GetLockoutTTL(ctx context.Context, username string) (time.Duration, error) {
	key := fmt.Sprintf(keyLockout, username)
	ttl, err := r.rdb.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return ttl, nil
}

func (r *authRepo) StoreSession(ctx context.Context, userID, tokenID, ip, userAgent string, ttl time.Duration) error {
	key := fmt.Sprintf(keySession, tokenID)
	val := fmt.Sprintf("%s|%s|%s", userID, ip, userAgent)
	return r.rdb.Set(ctx, key, val, ttl).Err()
}

func (r *authRepo) DeleteSession(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf(keySession, tokenID)
	return r.rdb.Del(ctx, key).Err()
}

// GenerateRefreshToken generates a cryptographically random refresh token string.
func (r *authRepo) GenerateRefreshToken() string {
	return uuid.NewString() + "-" + uuid.NewString()
}

// MaxLoginAttempts returns the maximum allowed failed login attempts before lockout.
func MaxLoginAttempts() int {
	return maxLoginAttempts
}

// LockoutDuration returns the lockout duration.
func LockoutDuration() time.Duration {
	return lockoutDuration
}
