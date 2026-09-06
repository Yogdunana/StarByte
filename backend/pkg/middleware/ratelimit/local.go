package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// localLimiter is the in-process token bucket used when Redis is down (#75).
type localLimiter struct {
	mu sync.Mutex
	m  map[string]*rate.Limiter
}

func newLocalLimiter() *localLimiter {
	return &localLimiter{m: make(map[string]*rate.Limiter)}
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
