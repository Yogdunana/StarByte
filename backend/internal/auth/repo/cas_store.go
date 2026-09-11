package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyCASState = "auth:cas:state:%s"
	keyCASCode  = "auth:cas:code:%s"
)

// CASTicketStore persists short-lived CAS login state and one-time exchange codes.
type CASTicketStore interface {
	PutState(ctx context.Context, state, redirect string, ttl time.Duration) error
	TakeState(ctx context.Context, state string) (redirect string, ok bool, err error)
	PutCode(ctx context.Context, code string, payload []byte, ttl time.Duration) error
	TakeCode(ctx context.Context, code string) ([]byte, error)
}

type casStore struct {
	rdb *redis.Client
}

// NewCASStore creates a Redis-backed CAS ticket store. rdb may be nil (all ops no-op / miss).
func NewCASStore(rdb *redis.Client) CASTicketStore {
	return &casStore{rdb: rdb}
}

func (s *casStore) PutState(ctx context.Context, state, redirect string, ttl time.Duration) error {
	if s == nil || s.rdb == nil || state == "" {
		return fmt.Errorf("cas store unavailable")
	}
	return s.rdb.Set(ctx, fmt.Sprintf(keyCASState, state), redirect, ttl).Err()
}

func (s *casStore) TakeState(ctx context.Context, state string) (string, bool, error) {
	if s == nil || s.rdb == nil || state == "" {
		return "", false, nil
	}
	key := fmt.Sprintf(keyCASState, state)
	val, err := s.rdb.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (s *casStore) PutCode(ctx context.Context, code string, payload []byte, ttl time.Duration) error {
	if s == nil || s.rdb == nil || code == "" {
		return fmt.Errorf("cas store unavailable")
	}
	return s.rdb.Set(ctx, fmt.Sprintf(keyCASCode, code), payload, ttl).Err()
}

func (s *casStore) TakeCode(ctx context.Context, code string) ([]byte, error) {
	if s == nil || s.rdb == nil || code == "" {
		return nil, redis.Nil
	}
	key := fmt.Sprintf(keyCASCode, code)
	val, err := s.rdb.GetDel(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return val, nil
}
