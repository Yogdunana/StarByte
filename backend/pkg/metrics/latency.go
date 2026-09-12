package metrics

import (
	"math"
	"sort"
	"sync"
)

const recentLatencyCap = 2048

type latencyWindow struct {
	mu      sync.Mutex
	samples []float64
	next    int
	n       int
}

var recentHTTP = newLatencyWindow(recentLatencyCap)

func newLatencyWindow(cap int) *latencyWindow {
	if cap <= 0 {
		cap = recentLatencyCap
	}
	return &latencyWindow{samples: make([]float64, cap)}
}

// RecordHTTPLatency stores a request duration (seconds) for recent P50/P95/P99.
func RecordHTTPLatency(seconds float64) {
	recentHTTP.record(seconds)
}

// RecentHTTPPercentiles returns nearest-rank P50/P95/P99 of the recent window.
func RecentHTTPPercentiles() (p50, p95, p99 float64, ok bool) {
	return recentHTTP.percentiles()
}

// ResetRecentHTTPLatenciesForTest clears the process-wide window. Tests only.
func ResetRecentHTTPLatenciesForTest() {
	recentHTTP.reset()
}

func (w *latencyWindow) record(seconds float64) {
	if w == nil || seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.samples) == 0 {
		return
	}
	w.samples[w.next] = seconds
	w.next = (w.next + 1) % len(w.samples)
	if w.n < len(w.samples) {
		w.n++
	}
}

func (w *latencyWindow) percentiles() (p50, p95, p99 float64, ok bool) {
	w.mu.Lock()
	n := w.n
	if n == 0 {
		w.mu.Unlock()
		return 0, 0, 0, false
	}
	cp := make([]float64, n)
	if n < len(w.samples) {
		copy(cp, w.samples[:n])
	} else {
		// ring is full: samples are [next..end) + [0..next)
		copied := copy(cp, w.samples[w.next:])
		copy(cp[copied:], w.samples[:w.next])
	}
	w.mu.Unlock()

	p50, p95, p99 = nearestRankPercentiles(cp)
	return p50, p95, p99, true
}

func (w *latencyWindow) reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.next = 0
	w.n = 0
}

func nearestRankPercentiles(samples []float64) (p50, p95, p99 float64) {
	sort.Float64s(samples)
	return nearestRank(samples, 0.50), nearestRank(samples, 0.95), nearestRank(samples, 0.99)
}

func nearestRank(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	idx := int(math.Ceil(p*float64(n))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return sorted[idx]
}
