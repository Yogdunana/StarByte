package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	// 始终 SCAN：索引里已有新登录时，部署前未入集的旧 session 仍需一并返回
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
	if err := r.deleteRefreshTokensMatching(ctx, userID, ""); err != nil {
		return err
	}
	return r.rdb.Del(ctx, fmt.Sprintf(keyUserRefresh, userID)).Err()
}

func (r *authRepo) DeleteRefreshTokensByJTI(ctx context.Context, userID, jti string) error {
	if userID == "" || jti == "" {
		return nil
	}
	return r.deleteRefreshTokensMatching(ctx, userID, jti)
}

// deleteRefreshTokensMatching 扫描 auth:refresh:*，覆盖从未写入 auth:user_refresh 的旧 key。
// jti 为空时删除该用户全部 refresh；非空时删除匹配 jti 的记录，以及无法绑定会话的旧格式（无 jti）记录。
func (r *authRepo) deleteRefreshTokensMatching(ctx context.Context, userID, jti string) error {
	var cursor uint64
	prefix := fmt.Sprintf(keyRefreshToken, "")
	for {
		keys, next, err := r.rdb.Scan(ctx, cursor, fmt.Sprintf(keyRefreshToken, "*"), 64).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			val, err := r.rdb.Get(ctx, key).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return err
			}
			uid, recJTI := parseRefreshRecord(val)
			if uid != userID {
				continue
			}
			if jti != "" && recJTI != "" && recJTI != jti {
				continue
			}
			_ = r.DeleteRefreshToken(ctx, strings.TrimPrefix(key, prefix))
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
