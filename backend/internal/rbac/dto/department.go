package dto

import "time"

// ========== 部门 DTO ==========

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Code        string `json:"code" binding:"required,max=50"`
	ParentID    string `json:"parent_id" binding:"omitempty"`
	LeaderID    string `json:"leader_id" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int   `json:"sort_order"`
}

// UpdateDepartmentRequest 更新部门请求
// Description 使用 *string：nil 表示不修改，空字符串表示清空
type UpdateDepartmentRequest struct {
	Name        string  `json:"name" binding:"omitempty,max=100"`
	LeaderID    *string `json:"leader_id" binding:"omitempty"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int    `json:"sort_order"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// DepartmentResponse 部门响应
type DepartmentResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Code        string                   `json:"code"`
	ParentID    *string                  `json:"parent_id"`
	LeaderID    *string                  `json:"leader_id"`
	Description string                   `json:"description"`
	SortOrder   int                      `json:"sort_order"`
	Status      int                      `json:"status"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Children    []DepartmentTreeResponse `json:"children,omitempty"`
}

// DepartmentTreeResponse 部门树响应
type DepartmentTreeResponse = DepartmentResponse
