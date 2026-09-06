package service

import (
	"sync"
	"time"
)

const emailPerMinute = 50

type minuteLimiter struct {
	mu     sync.Mutex
	stamps []time.Time
	limit  int
	window time.Duration
	now    func() time.Time
}

func NewMinuteLimiter(limit int, now func() time.Time) *minuteLimiter {
	return newMinuteLimiter(limit, now)
}

func newMinuteLimiter(limit int, now func() time.Time) *minuteLimiter {
	if now == nil {
		now = time.Now
	}
	if limit <= 0 {
		limit = emailPerMinute
	}
	return &minuteLimiter{limit: limit, window: time.Minute, now: now}
}

func (l *minuteLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	cut := now.Add(-l.window)
	kept := l.stamps[:0]
	for _, t := range l.stamps {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	l.stamps = kept
	if len(l.stamps) >= l.limit {
		return false
	}
	l.stamps = append(l.stamps, now)
	return true
}
