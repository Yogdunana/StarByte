package service

import (
	"math"
	"sort"

	"github.com/Yogdunana/StarByte/backend/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dtoProm "github.com/prometheus/client_model/go"
)

const httpDurationMetric = "starbyte_http_request_duration_seconds"

func resolvePercentiles(gatherer prometheus.Gatherer) (p50, p95, p99 *float64, source, note string) {
	if a, b, c, ok := metrics.RecentHTTPPercentiles(); ok {
		return ptrSec(a), ptrSec(b), ptrSec(c), "recent_requests",
			"P50/P95/P99 来自进程内最近请求耗时窗口（最多 2048 条）；直方图同时写入 /metrics。"
	}
	if gatherer == nil {
		return nil, nil, nil, "", "尚无请求样本，无法计算分位数。"
	}
	families, err := gatherer.Gather()
	if err != nil {
		return nil, nil, nil, "", "读取 Prometheus 直方图失败，无法计算分位数。"
	}
	bounds, cumul, count := mergeDurationHistograms(families)
	if count == 0 {
		return nil, nil, nil, "", "尚无请求样本，无法计算分位数。"
	}
	return quantile(0.50, bounds, cumul, count),
		quantile(0.95, bounds, cumul, count),
		quantile(0.99, bounds, cumul, count),
		"prometheus_histogram",
		"P50/P95/P99 由 starbyte_http_request_duration_seconds 直方图插值（进程累计）。"
}

func ptrSec(v float64) *float64 {
	x := math.Round(v*1e6) / 1e6
	return &x
}

func mergeDurationHistograms(families []*dtoProm.MetricFamily) (bounds []float64, cumul []uint64, count uint64) {
	increments := map[float64]uint64{}
	for _, mf := range families {
		if mf.GetName() != httpDurationMetric {
			continue
		}
		for _, m := range mf.GetMetric() {
			h := m.GetHistogram()
			if h == nil {
				continue
			}
			count += h.GetSampleCount()
			var prev uint64
			for _, b := range h.GetBucket() {
				inc := b.GetCumulativeCount()
				if inc >= prev {
					inc -= prev
				}
				increments[b.GetUpperBound()] += inc
				prev = b.GetCumulativeCount()
			}
		}
	}
	if len(increments) == 0 {
		return nil, nil, count
	}
	bounds = make([]float64, 0, len(increments))
	for u := range increments {
		bounds = append(bounds, u)
	}
	sort.Float64s(bounds)
	cumul = make([]uint64, len(bounds))
	var running uint64
	for i, u := range bounds {
		running += increments[u]
		cumul[i] = running
	}
	return bounds, cumul, count
}

func quantile(q float64, bounds []float64, cumul []uint64, count uint64) *float64 {
	if count == 0 || len(bounds) == 0 {
		return nil
	}
	rank := q * float64(count)
	var prevBound float64
	var prevCumul uint64
	for i, bound := range bounds {
		c := cumul[i]
		if float64(c) >= rank {
			if math.IsInf(bound, 1) {
				return ptrSec(prevBound)
			}
			width := bound - prevBound
			den := float64(c - prevCumul)
			frac := 0.0
			if den > 0 {
				frac = (rank - float64(prevCumul)) / den
			}
			return ptrSec(prevBound + width*frac)
		}
		prevBound, prevCumul = bound, c
	}
	last := bounds[len(bounds)-1]
	if math.IsInf(last, 1) && len(bounds) > 1 {
		last = bounds[len(bounds)-2]
	}
	return ptrSec(last)
}
