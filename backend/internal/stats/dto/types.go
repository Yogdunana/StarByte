package dto

import (
	"time"

	"github.com/google/uuid"
)

// StatsQuery 统一查询参数。
type StatsQuery struct {
	StartDate    *time.Time  `form:"start_date" json:"start_date"`
	EndDate      *time.Time  `form:"end_date" json:"end_date"`
	DepartmentID *uuid.UUID  `form:"department_id" json:"department_id"`
	GroupBy      string      `form:"group_by" json:"group_by"`
	Granularity  string      `form:"granularity" json:"granularity"`
	Format       string      `form:"format" json:"format"`
	Denied       bool        `json:"-"`
	AllScope     bool        `json:"-"`
	ScopeDeptIDs []uuid.UUID `json:"-"`
}

// ProviderInfo 注册中心对外描述。
type ProviderInfo struct {
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name"`
	ChartConfig *ChartConfig `json:"chart_config"`
}

// ChartConfig 图表建议。
type ChartConfig struct {
	ChartType  string `json:"chart_type"`
	Title      string `json:"title"`
	Stack      bool   `json:"stack"`
	Horizontal bool   `json:"horizontal"`
}

// DataPoint 单点。
type DataPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// DataSeries 数据系列。
type DataSeries struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Data  []DataPoint `json:"data"`
	XAxis []string    `json:"x_axis"`
}

// StatsResult 统计结果。
type StatsResult struct {
	Provider    string         `json:"provider"`
	Summary     map[string]any `json:"summary"`
	Series      []DataSeries   `json:"series"`
	ChartConfig *ChartConfig   `json:"chart_config"`
}

// OverviewMeeting 今日会议。
type OverviewMeeting struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	StartTime string `json:"start_time"`
}

// OverviewMyTasks 我的任务计数。
type OverviewMyTasks struct {
	Todo    int64 `json:"todo"`
	Overdue int64 `json:"overdue"`
}

// OverviewResponse 首页概览。
type OverviewResponse struct {
	TotalMembers           int64             `json:"total_members"`
	TotalMeetingsThisMonth int64             `json:"total_meetings_this_month"`
	TotalTasksInProgress   int64             `json:"total_tasks_in_progress"`
	TotalInternshipsActive int64             `json:"total_internships_active"`
	PendingApprovals       int64             `json:"pending_approvals"`
	TodayMeetings          []OverviewMeeting `json:"today_meetings"`
	MyTasks                OverviewMyTasks   `json:"my_tasks"`
	NotificationsUnread    int64             `json:"notifications_unread"`
}
