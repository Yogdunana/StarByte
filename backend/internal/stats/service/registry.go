package service

import (
	"context"
	"sort"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// StatsProvider 可插拔统计提供者。
type StatsProvider interface {
	Name() string
	DisplayName() string
	GetStats(ctx context.Context, params *dto.StatsQuery) (*dto.StatsResult, error)
	GetChartConfig() *dto.ChartConfig
}

// StatsRegistry 统计注册中心。
type StatsRegistry struct {
	mu        sync.RWMutex
	providers map[string]StatsProvider
}

// NewRegistry 创建空注册中心。
func NewRegistry() *StatsRegistry {
	return &StatsRegistry{providers: map[string]StatsProvider{}}
}

// Register 注册提供者。同名覆盖。
func (r *StatsRegistry) Register(provider StatsProvider) {
	if provider == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// ListProviders 按名称排序返回。
func (r *StatsRegistry) ListProviders() []dto.ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]dto.ProviderInfo, 0, len(names))
	for _, n := range names {
		p := r.providers[n]
		out = append(out, dto.ProviderInfo{Name: p.Name(), DisplayName: p.DisplayName(), ChartConfig: p.GetChartConfig()})
	}
	return out
}

// GetStats 按名称查询。
func (r *StatsRegistry) GetStats(ctx context.Context, provider string, params *dto.StatsQuery) (*dto.StatsResult, error) {
	r.mu.RLock()
	p, ok := r.providers[provider]
	r.mu.RUnlock()
	if !ok {
		return nil, response.NewError(response.CodeStatsProviderNotFound, "统计提供者不存在")
	}
	if params == nil {
		params = &dto.StatsQuery{}
	}
	res, err := p.GetStats(ctx, params)
	if err != nil {
		return nil, err
	}
	if res != nil {
		res.Series = filterSeries(params.GroupBy, res.Series)
	}
	return res, nil
}
