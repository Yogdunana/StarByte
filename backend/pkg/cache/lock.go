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

// Lock is a Redis SET NX lock with optional watchdog and reentrant count.
// Each handle tracks only its own hold count; Redis stores the owner-wide count.
type Lock struct {
	rdb     *redis.Client
	key     string
	owner   string
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

func lockTTLMillis(ttl time.Duration) int64 {
	ms := ttl.Milliseconds()
	if ms < 1 {
		return 1000
	}
	return ms
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
	n, err := acquireScript.Run(ctx, rdb, []string{key}, owner, strconv.FormatInt(lockTTLMillis(ttl), 10)).Int64()
	if err != nil {
		return nil, err
	}
	if n < 1 {
		return nil, ErrLockBusy
	}
	return &Lock{rdb: rdb, key: key, owner: owner, token: encodeToken(owner, int(n)), ttl: ttl, count: 1}, nil
}

func (l *Lock) Token() string { return l.token }

// Reenter increments this handle and the Redis owner count atomically.
func (l *Lock) Reenter(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count < 1 {
		return ErrLockNotHeld
	}
	n, err := reenterScript.Run(ctx, l.rdb, []string{l.key}, l.owner, strconv.FormatInt(lockTTLMillis(l.ttl), 10)).Int64()
	if err != nil {
		return err
	}
	if n < 1 {
		return ErrLockNotHeld
	}
	l.count++
	l.token = encodeToken(l.owner, int(n))
	return nil
}

func (l *Lock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count < 1 {
		return ErrLockNotHeld
	}
	n, err := unlockOnceScript.Run(ctx, l.rdb, []string{l.key}, l.owner, strconv.FormatInt(lockTTLMillis(l.ttl), 10)).Int64()
	if err != nil {
		return err
	}
	if n == 0 {
		l.count = 0
		l.stopWatchdogLocked()
		return ErrLockNotHeld
	}
	l.count--
	if n < 0 || l.count < 1 {
		l.stopWatchdogLocked()
		l.count = 0
		return nil
	}
	l.token = encodeToken(l.owner, int(n))
	return nil
}

func (l *Lock) Refresh(ctx context.Context) error {
	l.mu.Lock()
	owner := l.owner
	ttl := l.ttl
	l.mu.Unlock()
	n, err := extendScript.Run(ctx, l.rdb, []string{l.key}, owner, strconv.FormatInt(lockTTLMillis(ttl), 10)).Int64()
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
