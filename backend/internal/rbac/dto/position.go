package dto

import "time"

// ========== 职位 DTO ==========

// CreatePositionRequest 创建职位请求
type CreatePositionRequest struct {
	Name        string   `json:"name" binding:"required,max=50"`
	Code        string   `json:"code" binding:"required,max=50"`
	Level       *int     `json:"level"`
	VoteWeight  *float64 `json:"vote_weight"`
	Description string   `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int     `json:"sort_order"`
}

// UpdatePositionRequest 更新职位请求
type UpdatePositionRequest struct {
	Name        string   `json:"name" binding:"omitempty,max=50"`
	Level       *int     `json:"level"`
	VoteWeight  *float64 `json:"vote_weight"`
	Description string   `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int     `json:"sort_order"`
	Status      *int     `json:"status" binding:"omitempty,oneof=0 1"`
}

// ListPositionRequest 职位列表请求
type ListPositionRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size" binding:"max=100"`
	Keyword  string `form:"keyword"`
}

// PositionResponse 职位响应
type PositionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Level       int       `json:"level"`
	VoteWeight  float64   `json:"vote_weight"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
