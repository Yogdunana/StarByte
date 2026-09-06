package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/model"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// AuthRepo handles all Redis-based authentication data operations.
type AuthRepo interface {
	StoreRefreshToken(ctx context.Context, token, userID, accessJTI string, ttl time.Duration) error
	GetRefreshTokenUserID(ctx context.Context, token string) (string, error)
	GetRefreshTokenMeta(ctx context.Context, token string) (userID, jti string, err error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokensByUser(ctx context.Context, userID string) error
	DeleteRefreshTokensByJTI(ctx context.Context, userID, jti string) error

	BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)

	IncrLoginAttempts(ctx context.Context, username string) (int64, error)
	GetLoginAttempts(ctx context.Context, username string) (int64, error)
	ResetLoginAttempts(ctx context.Context, username string) error
	SetLockout(ctx context.Context, username string, duration time.Duration) error
	IsLockedOut(ctx context.Context, username string) (bool, error)
	GetLockoutTTL(ctx context.Context, username string) (time.Duration, error)

	StoreSession(ctx context.Context, userID, tokenID, ip, userAgent string, ttl time.Duration) error
	GetSession(ctx context.Context, tokenID string) (*model.Session, error)
	DeleteSession(ctx context.Context, tokenID string) error
	ListSessions(ctx context.Context) ([]model.Session, error)
	ListSessionsByUser(ctx context.Context, userID string) ([]model.Session, error)

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
	keyUserSessions  = "auth:user_sessions:%s"
	keyUserRefresh   = "auth:user_refresh:%s"

	maxLoginAttempts = 5
	lockoutDuration  = 15 * time.Minute
	sessionIndexTTL  = 8 * 24 * time.Hour
)

type refreshRecord struct {
	UserID string `json:"user_id"`
	JTI    string `json:"jti,omitempty"`
}

func parseRefreshRecord(val string) (userID, jti string) {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "{") {
		var rec refreshRecord
		if json.Unmarshal([]byte(val), &rec) == nil && rec.UserID != "" {
			return rec.UserID, rec.JTI
		}
	}
	return val, ""
}

func (r *authRepo) StoreRefreshToken(ctx context.Context, token, userID, accessJTI string, ttl time.Duration) error {
	key := fmt.Sprintf(keyRefreshToken, token)
	data, err := json.Marshal(refreshRecord{UserID: userID, JTI: accessJTI})
	if err != nil {
		return fmt.Errorf("marshal refresh: %w", err)
	}
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, key, string(data), ttl)
	ukey := fmt.Sprintf(keyUserRefresh, userID)
	pipe.SAdd(ctx, ukey, token)
	pipe.Expire(ctx, ukey, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *authRepo) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	uid, _, err := r.GetRefreshTokenMeta(ctx, token)
	return uid, err
}

func (r *authRepo) GetRefreshTokenMeta(ctx context.Context, token string) (string, string, error) {
	key := fmt.Sprintf(keyRefreshToken, token)
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", "", redis.Nil
	}
	if err != nil {
		return "", "", err
	}
	uid, jti := parseRefreshRecord(val)
	return uid, jti, nil
}

func (r *authRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf(keyRefreshToken, token)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}
	if val != "" {
		if userID, _ := parseRefreshRecord(val); userID != "" {
			_ = r.rdb.SRem(ctx, fmt.Sprintf(keyUserRefresh, userID), token).Err()
		}
	}
	return nil
}

func (r *authRepo) BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
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
	session := model.Session{
		UserID:    userID,
		TokenID:   tokenID,
		IP:        ip,
		UserAgent: userAgent,
		LoginAt:   time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, key, string(data), ttl)
	ukey := fmt.Sprintf(keyUserSessions, userID)
	pipe.SAdd(ctx, ukey, tokenID)
	pipe.Expire(ctx, ukey, sessionIndexTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *authRepo) DeleteSession(ctx context.Context, tokenID string) error {
	sess, err := r.GetSession(ctx, tokenID)
	if err != nil {
		return err
	}
	if err := r.rdb.Del(ctx, fmt.Sprintf(keySession, tokenID)).Err(); err != nil {
		return err
	}
	if sess != nil && sess.UserID != "" {
		_ = r.rdb.SRem(ctx, fmt.Sprintf(keyUserSessions, sess.UserID), tokenID).Err()
	}
	return nil
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
