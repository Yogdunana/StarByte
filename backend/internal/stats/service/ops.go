package service

import (
	"context"
	"sort"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/internal/stats/repo"
)

type meetingProvider struct{ repo repo.StatsRepo }

func (p *meetingProvider) Name() string        { return "meeting-attendance" }
func (p *meetingProvider) DisplayName() string { return "会议出席统计" }
func (p *meetingProvider) GetChartConfig() *dto.ChartConfig {
	return &dto.ChartConfig{ChartType: "line", Title: "会议出席"}
}

func (p *meetingProvider) GetStats(ctx context.Context, params *dto.StatsQuery) (*dto.StatsResult, error) {
	q := toQuery(params)
	trend, err := p.repo.MeetingAttendanceTrend(ctx, q)
	if err != nil {
		return nil, err
	}
	dept, err := p.repo.MeetingByDepartment(ctx, q)
	if err != nil {
		return nil, err
	}
	cal, err := p.repo.MeetingCalendar(ctx, q)
	if err != nil {
		return nil, err
	}
	avg := 0.0
	if len(trend) > 0 {
		sum := 0.0
		for _, b := range trend {
			sum += b.Value
		}
		avg = round4(sum / float64(len(trend)))
	}
	return &dto.StatsResult{
		Provider: p.Name(),
		Summary:  map[string]any{"avg_attendance_rate": avg, "meeting_days": int64(len(cal))},
		Series: []dto.DataSeries{
			toSeries("出席率", "line", trend),
			toSeries("各部门会议数", "bar", dept),
			toSeries("会议频率", "calendar", cal),
		},
		ChartConfig: p.GetChartConfig(),
	}, nil
}

type taskProvider struct{ repo repo.StatsRepo }

func (p *taskProvider) Name() string        { return "task-progress" }
func (p *taskProvider) DisplayName() string { return "任务进度统计" }
func (p *taskProvider) GetChartConfig() *dto.ChartConfig {
	return &dto.ChartConfig{ChartType: "pie", Title: "任务进度", Stack: true}
}

func (p *taskProvider) GetStats(ctx context.Context, params *dto.StatsQuery) (*dto.StatsResult, error) {
	q := toQuery(params)
	byStatus, err := p.repo.TaskByStatus(ctx, q)
	if err != nil {
		return nil, err
	}
	onTime, err := p.repo.TaskOnTimeRate(ctx, q)
	if err != nil {
		return nil, err
	}
	statuses, points, err := p.repo.TaskTrend(ctx, q)
	if err != nil {
		return nil, err
	}
	sort.Strings(statuses)
	labels := map[string]string{"0": "待处理", "1": "进行中", "2": "已完成", "3": "已取消", "4": "已挂起"}
	series := []dto.DataSeries{toSeries("任务状态", "pie", byStatus)}
	for _, st := range statuses {
		name := labels[st]
		if name == "" {
			name = st
		}
		s := toSeries(name, "bar", points[st])
		series = append(series, s)
	}
	var total float64
	for _, b := range byStatus {
		total += b.Value
	}
	return &dto.StatsResult{
		Provider: p.Name(),
		Summary: map[string]any{
			"total_tasks":  total,
			"on_time_rate": round4(onTime),
			"in_progress":  statusValue(byStatus, "1"),
		},
		Series:      series,
		ChartConfig: p.GetChartConfig(),
	}, nil
}

func statusValue(buckets []repo.Bucket, key string) float64 {
	for _, b := range buckets {
		if b.Key == key {
			return b.Value
		}
	}
	return 0
}

type internshipProvider struct{ repo repo.StatsRepo }

func (p *internshipProvider) Name() string        { return "internship-duration" }
func (p *internshipProvider) DisplayName() string { return "实习时长统计" }
func (p *internshipProvider) GetChartConfig() *dto.ChartConfig {
	return &dto.ChartConfig{ChartType: "bar", Title: "实习时长", Horizontal: true}
}

func (p *internshipProvider) GetStats(ctx context.Context, params *dto.StatsQuery) (*dto.StatsResult, error) {
	q := toQuery(params)
	rank, err := p.repo.InternshipRanking(ctx, q)
	if err != nil {
		return nil, err
	}
	dept, err := p.repo.InternshipDeptAvg(ctx, q)
	if err != nil {
		return nil, err
	}
	trend, err := p.repo.InternshipTrend(ctx, q)
	if err != nil {
		return nil, err
	}
	return &dto.StatsResult{
		Provider: p.Name(),
		Summary:  map[string]any{"ranked_users": int64(len(rank))},
		Series: []dto.DataSeries{
			toSeries("时长排行", "bar", rank),
			toSeries("部门平均时长", "bar", dept),
			toSeries("时长趋势", "line", trend),
		},
		ChartConfig: p.GetChartConfig(),
	}, nil
}
