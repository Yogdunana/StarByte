package model

import (
	"time"

	"github.com/google/uuid"
)

// Evaluation 按维度评分。interviewer_id 沿用 000011。
type Evaluation struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	InterviewID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_interview_eval_dim" json:"interview_id"`
	InterviewerID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_interview_eval_dim" json:"evaluator_id"`
	Dimension      string    `gorm:"type:varchar(50);not null;uniqueIndex:uq_interview_eval_dim" json:"dimension"`
	Score          float64   `gorm:"type:decimal(5,2);not null" json:"score"`
	Comment        string    `gorm:"type:text" json:"comment"`
	Recommendation int16     `gorm:"type:smallint;not null;default:3" json:"recommendation"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Evaluation) TableName() string { return "interview_evaluations" }

// EvaluationNamed 评分带面试官姓名。
type EvaluationNamed struct {
	Evaluation
	EvaluatorName string `gorm:"column:evaluator_name"`
}

// Dimension 评分维度配置。
type Dimension struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"name"`
	Weight    float64   `gorm:"type:numeric(5,2);not null;default:1" json:"weight"`
	MaxScore  float64   `gorm:"type:numeric(5,2);not null;default:100" json:"max_score"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Dimension) TableName() string { return "interview_dimensions" }

// StatsRow 统计聚合。
type StatsRow struct {
	Total        int64
	PassCount    int64
	FailCount    int64
	PendingCount int64
}

// ScoreBucket 评分区间。
type ScoreBucket struct {
	Range string
	Count int64
}

// DeptStat 部门统计。
type DeptStat struct {
	Department string
	Count      int64
	PassCount  int64
}
