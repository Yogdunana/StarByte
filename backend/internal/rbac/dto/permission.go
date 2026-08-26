package dto

import "time"

// ========== 权限 DTO ==========

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Code        string `json:"code" binding:"required,max=100"`
	ParentID    string `json:"parent_id" binding:"omitempty"`
	Type        string `json:"type" binding:"required,oneof=menu button api"`
	Resource    string `json:"resource" binding:"omitempty,max=50"`
	Action      string `json:"action" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
	Path        string `json:"path" binding:"omitempty,max=255"`
	Icon        string `json:"icon" binding:"omitempty,max=50"`
	APIMethod   string `json:"api_method" binding:"omitempty,max=10"`
	APIPath     string `json:"api_path" binding:"omitempty,max=255"`
}

// UpdatePermissionRequest 更新权限请求
// Description 使用 *string：nil 表示不修改，空字符串表示清空
type UpdatePermissionRequest struct {
	Name        string  `json:"name" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	Path        string  `json:"path" binding:"omitempty,max=255"`
	Icon        string  `json:"icon" binding:"omitempty,max=50"`
	SortOrder   *int    `json:"sort_order"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Code        string                   `json:"code"`
	Resource    string                   `json:"resource"`
	Action      string                   `json:"action"`
	Description string                   `json:"description"`
	ParentID    *string                  `json:"parent_id"`
	SortOrder   int                      `json:"sort_order"`
	Type        string                   `json:"type"`
	Path        string                   `json:"path"`
	Icon        string                   `json:"icon"`
	APIMethod   string                   `json:"api_method"`
	APIPath     string                   `json:"api_path"`
	IsSystem    bool                     `json:"is_system"`
	Status      int                      `json:"status"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Children    []PermissionTreeResponse `json:"children,omitempty"`
}

// PermissionTreeResponse 权限树响应
type PermissionTreeResponse = PermissionResponse
