package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/internal/stats/repo"
)

type memberProvider struct{ repo repo.StatsRepo }

func (p *memberProvider) Name() string        { return "member-distribution" }
func (p *memberProvider) DisplayName() string { return "会员分布统计" }
func (p *memberProvider) GetChartConfig() *dto.ChartConfig {
	return &dto.ChartConfig{ChartType: "pie", Title: "会员分布"}
}

func (p *memberProvider) GetStats(ctx context.Context, params *dto.StatsQuery) (*dto.StatsResult, error) {
	q := toQuery(params)
	total, active, neu, err := p.repo.MemberSummary(ctx, q)
	if err != nil {
		return nil, err
	}
	dept, err := p.repo.MemberByDepartment(ctx, q)
	if err != nil {
		return nil, err
	}
	grade, err := p.repo.MemberByGrade(ctx, q)
	if err != nil {
		return nil, err
	}
	trend, err := p.repo.MemberTrend(ctx, q)
	if err != nil {
		return nil, err
	}
	return &dto.StatsResult{
		Provider: p.Name(),
		Summary: map[string]any{
			"total_members":  total,
			"active_members": active,
			"new_this_month": neu,
		},
		Series: []dto.DataSeries{
			toSeries("部门分布", "pie", dept),
			toSeries("年级分布", "bar", grade),
			toSeries("增长趋势", "line", trend),
		},
		ChartConfig: p.GetChartConfig(),
	}, nil
}

type interviewProvider struct{ repo repo.StatsRepo }

func (p *interviewProvider) Name() string        { return "interview-data" }
func (p *interviewProvider) DisplayName() string { return "面试数据统计" }
func (p *interviewProvider) GetChartConfig() *dto.ChartConfig {
	return &dto.ChartConfig{ChartType: "bar", Title: "面试数据"}
}

func (p *interviewProvider) GetStats(ctx context.Context, params *dto.StatsQuery) (*dto.StatsResult, error) {
	q := toQuery(params)
	total, pass, avg, err := p.repo.InterviewSummary(ctx, q)
	if err != nil {
		return nil, err
	}
	dept, err := p.repo.InterviewByDepartment(ctx, q)
	if err != nil {
		return nil, err
	}
	trend, err := p.repo.InterviewTrend(ctx, q)
	if err != nil {
		return nil, err
	}
	hist, err := p.repo.InterviewScoreHist(ctx, q)
	if err != nil {
		return nil, err
	}
	return &dto.StatsResult{
		Provider: p.Name(),
		Summary: map[string]any{
			"total_interviews": total,
			"pass_rate":        round4(pass),
			"avg_score":        round4(avg),
		},
		Series: []dto.DataSeries{
			{Name: "通过率", Type: "gauge", Data: []dto.DataPoint{{Label: "通过率", Value: round4(pass * 100)}}, XAxis: []string{"通过率"}},
			toSeries("各部门面试人数", "bar", dept),
			toSeries("评分分布", "bar", hist),
			toSeries("面试趋势", "line", trend),
		},
		ChartConfig: p.GetChartConfig(),
	}, nil
}
