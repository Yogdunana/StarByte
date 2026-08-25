package dto

import "time"

// ========== 角色 DTO ==========

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Code        string `json:"code" binding:"required,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// ListRoleRequest 角色列表请求
type ListRoleRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
	DataScope    string   `json:"data_scope" binding:"omitempty,oneof=all department department_and_sub self custom"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	ParentID    string    `json:"parent_id"`
	SortOrder   int       `json:"sort_order"`
	Status      int       `json:"status"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleListResponse 角色列表项响应
type RoleListResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Status    int    `json:"status"`
	IsSystem  bool   `json:"is_system"`
	SortOrder int    `json:"sort_order"`
}

// RoleDetailResponse 角色详情响应（含权限列表）
type RoleDetailResponse struct {
	RoleResponse
	PermissionIDs []string `json:"permission_ids"`
}

// RoleUserResponse 角色下用户响应
type RoleUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Status   int    `json:"status"`
}
