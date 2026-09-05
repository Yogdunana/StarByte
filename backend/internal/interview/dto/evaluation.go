package dto

import "time"

type EvaluationItem struct {
	Dimension string  `json:"dimension" binding:"required,max=50"`
	Score     float64 `json:"score"`
	Comment   string  `json:"comment"`
}

type SubmitEvaluationsRequest struct {
	Evaluations []EvaluationItem `json:"evaluations" binding:"required,min=1,dive"`
}

type UpdateEvaluationRequest struct {
	Score   *float64 `json:"score"`
	Comment *string  `json:"comment"`
}

type SubmitResultRequest struct {
	Result  int16  `json:"result" binding:"required,oneof=1 2 3"`
	Comment string `json:"comment"`
}

type DimensionScore struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"`
	Comment   string  `json:"comment,omitempty"`
}

type EvaluatorScores struct {
	Evaluator  Person           `json:"evaluator"`
	Scores     []DimensionScore `json:"scores"`
	TotalScore float64          `json:"total_score"`
}

type EvaluationSummary struct {
	InterviewID   string            `json:"interview_id"`
	Applicant     Person            `json:"applicant"`
	Evaluations   []EvaluatorScores `json:"evaluations"`
	AverageScore  float64           `json:"average_score"`
	WeightedScore float64           `json:"weighted_score"`
}

type EvaluationResponse struct {
	ID          string    `json:"id"`
	InterviewID string    `json:"interview_id"`
	Evaluator   Person    `json:"evaluator"`
	Dimension   string    `json:"dimension"`
	Score       float64   `json:"score"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateDimensionRequest struct {
	Name      string  `json:"name" binding:"required,max=50"`
	Weight    float64 `json:"weight" binding:"required,gt=0"`
	MaxScore  float64 `json:"max_score" binding:"required,gt=0"`
	SortOrder int     `json:"sort_order"`
}

type UpdateDimensionRequest struct {
	Name      *string  `json:"name"`
	Weight    *float64 `json:"weight"`
	MaxScore  *float64 `json:"max_score"`
	SortOrder *int     `json:"sort_order"`
}

type DimensionResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Weight    float64   `json:"weight"`
	MaxScore  float64   `json:"max_score"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StatsQuery struct {
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
	DepartmentID string `form:"department_id"`
	Round        *int16 `form:"round"`
}

type ScoreBucketVO struct {
	Range string `json:"range"`
	Count int64  `json:"count"`
}

type DeptStatVO struct {
	Department string `json:"department"`
	Count      int64  `json:"count"`
	PassCount  int64  `json:"pass_count"`
}

type StatsResponse struct {
	Total        int64           `json:"total"`
	PassCount    int64           `json:"pass_count"`
	FailCount    int64           `json:"fail_count"`
	PendingCount int64           `json:"pending_count"`
	PassRate     float64         `json:"pass_rate"`
	ScoreBuckets []ScoreBucketVO `json:"score_buckets"`
	ByDepartment []DeptStatVO    `json:"by_department"`
}
