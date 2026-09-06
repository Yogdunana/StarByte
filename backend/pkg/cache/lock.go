package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const lockPrefix = "lock:"

var (
	ErrLockNotHeld = errors.New("cache: lock not held")
	ErrLockBusy    = errors.New("cache: lock busy")
)

var extendScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`)

var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

// Lock is a Redis SET NX lock with optional watchdog and reentrant count.
type Lock struct {
	rdb     *redis.Client
	key     string
	token   string
	ttl     time.Duration
	mu      sync.Mutex
	count   int
	stopWD  chan struct{}
	stopped bool
}

func lockKey(name string) string { return lockPrefix + name }

func encodeToken(owner string, n int) string {
	return owner + ":" + strconv.Itoa(n)
}

func parseToken(raw string) (owner string, n int) {
	i := strings.LastIndex(raw, ":")
	if i < 0 {
		return raw, 1
	}
	n, _ = strconv.Atoi(raw[i+1:])
	if n < 1 {
		n = 1
	}
	return raw[:i], n
}

// Acquire takes a non-fair lock. owner may be empty (a UUID is used).
func Acquire(ctx context.Context, rdb *redis.Client, name, owner string, ttl time.Duration) (*Lock, error) {
	if ttl <= 0 {
		ttl = 8 * time.Second
	}
	if owner == "" {
		owner = uuid.NewString()
	}
	key := lockKey(name)
	cur, err := rdb.Get(ctx, key).Result()
	if err == nil {
		o, n := parseToken(cur)
		if o == owner {
			n++
			if err := rdb.Set(ctx, key, encodeToken(owner, n), ttl).Err(); err != nil {
				return nil, err
			}
			return &Lock{rdb: rdb, key: key, token: encodeToken(owner, n), ttl: ttl, count: n}, nil
		}
		return nil, ErrLockBusy
	}
	if err != redis.Nil {
		return nil, err
	}
	tok := encodeToken(owner, 1)
	ok, err := rdb.SetNX(ctx, key, tok, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockBusy
	}
	return &Lock{rdb: rdb, key: key, token: tok, ttl: ttl, count: 1}, nil
}

func (l *Lock) Token() string { return l.token }

// Reenter increments the in-process + Redis reentrant count for this lock.
func (l *Lock) Reenter(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count < 1 {
		return ErrLockNotHeld
	}
	owner, _ := parseToken(l.token)
	l.count++
	l.token = encodeToken(owner, l.count)
	return l.rdb.Set(ctx, l.key, l.token, l.ttl).Err()
}

func (l *Lock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count > 1 {
		l.count--
		owner, _ := parseToken(l.token)
		l.token = encodeToken(owner, l.count)
		return l.rdb.Set(ctx, l.key, l.token, l.ttl).Err()
	}
	l.stopWatchdogLocked()
	n, err := unlockScript.Run(ctx, l.rdb, []string{l.key}, l.token).Int64()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrLockNotHeld
	}
	l.count = 0
	return nil
}

func (l *Lock) Refresh(ctx context.Context) error {
	l.mu.Lock()
	token := l.token
	ttl := l.ttl
	l.mu.Unlock()
	ms := ttl.Milliseconds()
	if ms < 1 {
		ms = 1000
	}
	n, err := extendScript.Run(ctx, l.rdb, []string{l.key}, token, strconv.FormatInt(ms, 10)).Int64()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// StartWatchdog extends the lock until StopWatchdog or Unlock.
func (l *Lock) StartWatchdog() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.stopWD != nil {
		return
	}
	l.stopWD = make(chan struct{})
	interval := l.ttl / 3
	if interval < 200*time.Millisecond {
		interval = 200 * time.Millisecond
	}
	go func(stop <-chan struct{}) {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				_ = l.Refresh(context.Background())
			}
		}
	}(l.stopWD)
}

func (l *Lock) StopWatchdog() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stopWatchdogLocked()
}

func (l *Lock) stopWatchdogLocked() {
	if l.stopWD != nil && !l.stopped {
		close(l.stopWD)
		l.stopped = true
	}
}

func (l *Lock) String() string {
	return fmt.Sprintf("lock(%s)", l.key)
}

// AcquireFair waits in a Redis list until it can take the lock (simple FIFO).
func AcquireFair(ctx context.Context, rdb *redis.Client, name, owner string, ttl, wait time.Duration) (*Lock, error) {
	queue := lockPrefix + name + ":wait"
	ticket := uuid.NewString()
	if err := rdb.RPush(ctx, queue, ticket).Err(); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(wait)
	if wait <= 0 {
		deadline = time.Now().Add(5 * time.Second)
	}
	for time.Now().Before(deadline) {
		head, err := rdb.LIndex(ctx, queue, 0).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		if head == ticket {
			lk, aerr := Acquire(ctx, rdb, name, owner, ttl)
			if aerr == nil {
				_, _ = rdb.LPop(ctx, queue).Result()
				return lk, nil
			}
			if aerr != ErrLockBusy {
				_, _ = rdb.LRem(ctx, queue, 1, ticket).Result()
				return nil, aerr
			}
		}
		select {
		case <-ctx.Done():
			_, _ = rdb.LRem(ctx, queue, 1, ticket).Result()
			return nil, ctx.Err()
		case <-time.After(40 * time.Millisecond):
		}
	}
	_, _ = rdb.LRem(ctx, queue, 1, ticket).Result()
	return nil, ErrLockBusy
}
