package dto

import "time"

type TaskResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Code       string     `json:"code"`
	CronExpr   string     `json:"cron_expr"`
	RunAt      *time.Time `json:"run_at"`
	Timezone   string     `json:"timezone"`
	HandlerKey string     `json:"handler_key"`
	Payload    string     `json:"payload"`
	DependsOn  []string   `json:"depends_on"`
	ShardKey   string     `json:"shard_key"`
	Status     int16      `json:"status"`
	MaxRetries int        `json:"max_retries"`
	TimeoutSec int        `json:"timeout_sec"`
	RetryCount int        `json:"retry_count"`
	NextRunAt  *time.Time `json:"next_run_at"`
	LastRunAt  *time.Time `json:"last_run_at"`
	LastStatus string     `json:"last_status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateTaskRequest struct {
	Name       string   `json:"name" binding:"required"`
	Code       string   `json:"code" binding:"required"`
	CronExpr   string   `json:"cron_expr"`
	RunAt      *string  `json:"run_at"`
	Timezone   string   `json:"timezone"`
	HandlerKey string   `json:"handler_key" binding:"required"`
	Payload    string   `json:"payload"`
	DependsOn  []string `json:"depends_on"`
	ShardKey   string   `json:"shard_key"`
	MaxRetries *int     `json:"max_retries"`
	TimeoutSec *int     `json:"timeout_sec"`
}

type UpdateTaskRequest struct {
	Name       *string  `json:"name"`
	CronExpr   *string  `json:"cron_expr"`
	RunAt      *string  `json:"run_at"`
	Timezone   *string  `json:"timezone"`
	HandlerKey *string  `json:"handler_key"`
	Payload    *string  `json:"payload"`
	DependsOn  []string `json:"depends_on"`
	ShardKey   *string  `json:"shard_key"`
	MaxRetries *int     `json:"max_retries"`
	TimeoutSec *int     `json:"timeout_sec"`
}

type ListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   *int16 `form:"status"`
}

type HandlerInfo struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

type RunResponse struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	Status      string     `json:"status"`
	Attempt     int        `json:"attempt"`
	WorkerID    string     `json:"worker_id"`
	ErrorText   string     `json:"error_text"`
	Output      string     `json:"output"`
}

type LogLine struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`
	Line      string    `json:"line"`
	CreatedAt time.Time `json:"created_at"`
}

type LogsResponse struct {
	Runs []RunResponse `json:"runs"`
	Logs []LogLine     `json:"logs"`
}
