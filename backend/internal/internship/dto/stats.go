package dto

type DurationStatsRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	GroupBy   string `form:"group_by"`
}

type RankingRequest struct {
	DepartmentID string `form:"department_id"`
	Limit        int    `form:"limit"`
	SortBy       string `form:"sort_by"`
}

type DepartmentStatsRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type DurationItem struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	DurationDays int    `json:"duration_days"`
	Count        int64  `json:"count"`
}

type DurationStatsResponse struct {
	GroupBy string         `json:"group_by"`
	Items   []DurationItem `json:"items"`
}

type RankingItem struct {
	Rank              int     `json:"rank"`
	User              Person  `json:"user"`
	Department        *Person `json:"department,omitempty"`
	TotalDurationDays int     `json:"total_duration_days"`
	InternshipCount   int64   `json:"internship_count"`
	LatestInternship  string  `json:"latest_internship"`
}

type RankingResponse struct {
	Rankings []RankingItem `json:"rankings"`
	Total    int           `json:"total"`
}

type DepartmentStatItem struct {
	Department   *Person `json:"department,omitempty"`
	DurationDays int     `json:"duration_days"`
	Count        int64   `json:"count"`
	Ongoing      int64   `json:"ongoing"`
}

type DepartmentStatsResponse struct {
	Items []DepartmentStatItem `json:"items"`
}
