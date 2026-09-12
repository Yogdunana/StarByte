package metrics

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecentHTTPPercentiles(t *testing.T) {
	ResetRecentHTTPLatenciesForTest()
	t.Cleanup(ResetRecentHTTPLatenciesForTest)

	_, _, _, ok := RecentHTTPPercentiles()
	assert.False(t, ok)

	RecordHTTPLatency(math.NaN())
	RecordHTTPLatency(-1)
	RecordHTTPLatency(math.Inf(1))
	_, _, _, ok = RecentHTTPPercentiles()
	assert.False(t, ok)

	for i := 1; i <= 100; i++ {
		RecordHTTPLatency(float64(i) / 1000)
	}
	p50, p95, p99, ok := RecentHTTPPercentiles()
	require.True(t, ok)
	assert.InDelta(t, 0.050, p50, 0.002)
	assert.InDelta(t, 0.095, p95, 0.002)
	assert.InDelta(t, 0.099, p99, 0.002)
}

func TestLatencyWindowWraps(t *testing.T) {
	w := newLatencyWindow(4)
	for _, v := range []float64{1, 2, 3, 4, 5, 6} {
		w.record(v)
	}
	p50, _, _, ok := w.percentiles()
	require.True(t, ok)
	// window holds 3,4,5,6 — nearest-rank p50 is index 1
	assert.Equal(t, 4.0, p50)
}

func TestNearestRankEmpty(t *testing.T) {
	assert.Equal(t, 0.0, nearestRank(nil, 0.5))
	assert.Equal(t, 1.5, nearestRank([]float64{1.5}, 0.99))
}
