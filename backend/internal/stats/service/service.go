package service

import (
	"context"
	"fmt"
	"strings"

	exportDTO "github.com/Yogdunana/StarByte/backend/internal/export/dto"
	exportSvc "github.com/Yogdunana/StarByte/backend/internal/export/service"
	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/internal/stats/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

const maxExportPoints = 5000

// StatsService 统一入口。
type StatsService interface {
	ListProviders() []dto.ProviderInfo
	GetStats(ctx context.Context, provider string, q *dto.StatsQuery) (*dto.StatsResult, error)
	Overview(ctx context.Context, userID uuid.UUID) (*dto.OverviewResponse, error)
	Export(ctx context.Context, provider, format string, q *dto.StatsQuery) ([]byte, string, error)
}

type statsService struct {
	reg  *StatsRegistry
	repo repo.StatsRepo
}

// NewStatsService 注册一期 5 个提供者。
func NewStatsService(r repo.StatsRepo) StatsService {
	reg := NewRegistry()
	reg.Register(&memberProvider{repo: r})
	reg.Register(&interviewProvider{repo: r})
	reg.Register(&meetingProvider{repo: r})
	reg.Register(&taskProvider{repo: r})
	reg.Register(&internshipProvider{repo: r})
	return &statsService{reg: reg, repo: r}
}

func (s *statsService) ListProviders() []dto.ProviderInfo {
	return s.reg.ListProviders()
}

func (s *statsService) GetStats(ctx context.Context, provider string, q *dto.StatsQuery) (*dto.StatsResult, error) {
	if err := validateQuery(q); err != nil {
		return nil, err
	}
	return s.reg.GetStats(ctx, provider, q)
}

func (s *statsService) Overview(ctx context.Context, userID uuid.UUID) (*dto.OverviewResponse, error) {
	return s.repo.Overview(ctx, userID)
}

func (s *statsService) Export(ctx context.Context, provider, format string, q *dto.StatsQuery) ([]byte, string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "excel" {
		return nil, "", response.NewError(response.CodeStatsExportFormat, "导出格式不支持")
	}
	res, err := s.GetStats(ctx, provider, q)
	if err != nil {
		return nil, "", err
	}
	rows := make([][]string, 0)
	for _, series := range res.Series {
		for _, pt := range series.Data {
			rows = append(rows, []string{series.Name, pt.Label, fmt.Sprintf("%g", pt.Value)})
			if len(rows) > maxExportPoints {
				return nil, "", response.NewError(response.CodeStatsTooLarge, "数据量过大")
			}
		}
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"summary", "empty", "0"})
	}
	title := provider
	if res.ChartConfig != nil && res.ChartConfig.Title != "" {
		title = res.ChartConfig.Title
	}
	data, filename, err := exportSvc.RenderTable(format, &exportDTO.TableExportRequest{
		Filename: "stats_" + provider,
		Title:    title,
		Columns:  []string{"系列", "标签", "数值"},
		Rows:     rows,
		Sheet:    "stats",
	})
	if err != nil {
		return nil, "", err
	}
	return data, filename, nil
}
