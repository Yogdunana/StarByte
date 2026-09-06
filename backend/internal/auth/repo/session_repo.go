package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/auth/model"
	"github.com/redis/go-redis/v9"
)

func (r *authRepo) GetSession(ctx context.Context, tokenID string) (*model.Session, error) {
	if tokenID == "" {
		return nil, nil
	}
	val, err := r.rdb.Get(ctx, fmt.Sprintf(keySession, tokenID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sess model.Session
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (r *authRepo) ListSessions(ctx context.Context) ([]model.Session, error) {
	return r.scanSessions(ctx, "")
}

func (r *authRepo) ListSessionsByUser(ctx context.Context, userID string) ([]model.Session, error) {
	if userID == "" {
		return nil, nil
	}
	ids, err := r.rdb.SMembers(ctx, fmt.Sprintf(keyUserSessions, userID)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	out := make([]model.Session, 0, len(ids))
	for _, id := range ids {
		sess, err := r.GetSession(ctx, id)
		if err != nil {
			return nil, err
		}
		if sess == nil {
			_ = r.rdb.SRem(ctx, fmt.Sprintf(keyUserSessions, userID), id).Err()
			continue
		}
		out = append(out, *sess)
	}
	if len(out) > 0 {
		return out, nil
	}
	// 兼容登录时尚未写入用户索引的旧 session
	return r.scanSessions(ctx, userID)
}

func (r *authRepo) scanSessions(ctx context.Context, userID string) ([]model.Session, error) {
	var (
		out    []model.Session
		cursor uint64
	)
	for {
		keys, next, err := r.rdb.Scan(ctx, cursor, "auth:session:*", 64).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			val, err := r.rdb.Get(ctx, key).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return nil, err
			}
			var sess model.Session
			if err := json.Unmarshal([]byte(val), &sess); err != nil {
				continue
			}
			if userID != "" && sess.UserID != userID {
				continue
			}
			out = append(out, sess)
			if userID != "" && sess.TokenID != "" {
				_ = r.rdb.SAdd(ctx, fmt.Sprintf(keyUserSessions, userID), sess.TokenID).Err()
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return out, nil
}

func (r *authRepo) DeleteRefreshTokensByUser(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	ukey := fmt.Sprintf(keyUserRefresh, userID)
	tokens, err := r.rdb.SMembers(ctx, ukey).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	for _, token := range tokens {
		_ = r.rdb.Del(ctx, fmt.Sprintf(keyRefreshToken, token)).Err()
	}
	return r.rdb.Del(ctx, ukey).Err()
}

func (r *authRepo) DeleteRefreshTokensByJTI(ctx context.Context, userID, jti string) error {
	if userID == "" || jti == "" {
		return nil
	}
	tokens, err := r.rdb.SMembers(ctx, fmt.Sprintf(keyUserRefresh, userID)).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	for _, token := range tokens {
		uid, recJTI, err := r.GetRefreshTokenMeta(ctx, token)
		if err != nil {
			continue
		}
		if uid == userID && recJTI == jti {
			_ = r.DeleteRefreshToken(ctx, token)
		}
	}
	return nil
}
