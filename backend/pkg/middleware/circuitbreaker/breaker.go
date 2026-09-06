package circuitbreaker

import (
	"math"
	"sync"
	"time"
)

const (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half_open"
)

// Settings control trip / recover thresholds.
type Settings struct {
	MinRequests    int
	ErrorRate      float64       // 0.5 = 50%
	P99            time.Duration // trip if P99 above this
	Window         int           // last N samples
	OpenFor        time.Duration
	HalfOpenProbes int
}

func DefaultSettings() Settings {
	return Settings{
		MinRequests:    20,
		ErrorRate:      0.5,
		P99:            5 * time.Second,
		Window:         100,
		OpenFor:        30 * time.Second,
		HalfOpenProbes: 3,
	}
}

// Breaker is a three-state circuit per name (usually a route).
type Breaker struct {
	mu       sync.Mutex
	settings Settings
	now      func() time.Time
	byName   map[string]*probe
}

type probe struct {
	state     string
	openedAt  time.Time
	halfLeft  int
	samples   []sample
	halfFails int
	halfOK    int
}

type sample struct {
	err  bool
	took time.Duration
}

func New(s Settings) *Breaker {
	if s.Window <= 0 {
		s.Window = 100
	}
	if s.MinRequests <= 0 {
		s.MinRequests = 20
	}
	if s.HalfOpenProbes <= 0 {
		s.HalfOpenProbes = 3
	}
	if s.OpenFor <= 0 {
		s.OpenFor = 30 * time.Second
	}
	return &Breaker{settings: s, now: time.Now, byName: map[string]*probe{}}
}

func (b *Breaker) get(name string) *probe {
	p, ok := b.byName[name]
	if !ok {
		p = &probe{state: StateClosed}
		b.byName[name] = p
	}
	return p
}

// Allow reports whether a call may proceed. false means fail-fast (open).
func (b *Breaker) Allow(name string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	p := b.get(name)
	switch p.state {
	case StateOpen:
		if b.now().Sub(p.openedAt) >= b.settings.OpenFor {
			p.state = StateHalfOpen
			p.halfLeft = b.settings.HalfOpenProbes
			p.halfFails = 0
			p.halfOK = 0
			p.halfLeft--
			return true
		}
		return false
	case StateHalfOpen:
		if p.halfLeft <= 0 {
			return false
		}
		p.halfLeft--
		return true
	default:
		return true
	}
}

// Record stores one finished call.
func (b *Breaker) Record(name string, failed bool, took time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	p := b.get(name)
	p.samples = append(p.samples, sample{err: failed, took: took})
	if len(p.samples) > b.settings.Window {
		p.samples = p.samples[len(p.samples)-b.settings.Window:]
	}
	switch p.state {
	case StateHalfOpen:
		if failed {
			p.halfFails++
			b.trip(p)
			return
		}
		p.halfOK++
		if p.halfOK >= b.settings.HalfOpenProbes {
			p.state = StateClosed
			p.samples = nil
		}
	case StateClosed:
		if shouldTrip(p.samples, b.settings) {
			b.trip(p)
		}
	}
}

func (b *Breaker) trip(p *probe) {
	p.state = StateOpen
	p.openedAt = b.now()
	p.halfLeft = 0
}

func (b *Breaker) State(name string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.get(name).state
}

func shouldTrip(ss []sample, s Settings) bool {
	if len(ss) < s.MinRequests {
		return false
	}
	var fails int
	lats := make([]time.Duration, len(ss))
	for i, x := range ss {
		if x.err {
			fails++
		}
		lats[i] = x.took
	}
	if float64(fails)/float64(len(ss)) >= s.ErrorRate {
		return true
	}
	if s.P99 > 0 && percentile(lats, 0.99) >= s.P99 {
		return true
	}
	return false
}

func percentile(lats []time.Duration, p float64) time.Duration {
	if len(lats) == 0 {
		return 0
	}
	cp := append([]time.Duration(nil), lats...)
	// insertion sort — window is small
	for i := 1; i < len(cp); i++ {
		j := i
		for j > 0 && cp[j] < cp[j-1] {
			cp[j], cp[j-1] = cp[j-1], cp[j]
			j--
		}
	}
	idx := int(math.Ceil(p*float64(len(cp)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}
