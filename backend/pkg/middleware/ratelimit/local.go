package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

const defaultMaxLocalBuckets = 4096

// localLimiter is the in-process token bucket used when Redis is down (#75).
type localLimiter struct {
	mu  sync.Mutex
	m   map[string]*rate.Limiter
	max int
}

func newLocalLimiter() *localLimiter {
	return &localLimiter{m: make(map[string]*rate.Limiter), max: defaultMaxLocalBuckets}
}

func (l *localLimiter) allow(key string, b Bucket) Result {
	if l == nil {
		return Result{Allowed: true, Remaining: int64(b.Burst)}
	}
	burst := int(b.Burst)
	if burst < 1 {
		burst = 1
	}
	l.mu.Lock()
	lim, ok := l.m[key]
	if !ok {
		if l.max > 0 && len(l.m) >= l.max {
			// Bound memory: evict one arbitrary entry instead of growing forever.
			for k := range l.m {
				delete(l.m, k)
				break
			}
		}
		lim = rate.NewLimiter(rate.Limit(b.Rate), burst)
		l.m[key] = lim
	}
	l.mu.Unlock()

	if lim.Allow() {
		return Result{Allowed: true, Remaining: int64(lim.Tokens())}
	}
	retry := 1
	if b.Rate > 0 {
		retry = int(1.0/b.Rate + 0.999)
		if retry < 1 {
			retry = 1
		}
	}
	return Result{Allowed: false, Remaining: 0, RetryAfter: retry}
}
