package dto

import "time"

// ========== 角色 DTO ==========

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Code        string `json:"code" binding:"required,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
	ParentID    string `json:"parent_id" binding:"omitempty"`
	SortOrder   *int   `json:"sort_order"`
}

// UpdateRoleRequest 更新角色请求
// ParentID 使用 *string：nil 表示不修改，空字符串表示设为 null（提升为根节点）
// Description 使用 *string：nil 表示不修改，空字符串表示清空
type UpdateRoleRequest struct {
	Name        string  `json:"name" binding:"omitempty,max=50"`
	Code        string  `json:"code" binding:"omitempty,max=50"`
	ParentID    *string `json:"parent_id" binding:"omitempty"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// ListRoleRequest 角色列表请求
type ListRoleRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size" binding:"max=100"`
	Keyword  string `form:"keyword"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
	DataScope     string   `json:"data_scope" binding:"omitempty,oneof=all department department_and_sub self custom"`
}

// RoleUserListRequest 角色用户列表请求
type RoleUserListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size" binding:"max=100"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	ParentID    *string   `json:"parent_id"`
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
