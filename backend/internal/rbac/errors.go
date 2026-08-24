package rbac

import "github.com/Yogdunana/StarByte/backend/pkg/response"

// RBAC 模块错误码（3000-3999）
const (
	// 角色相关
	ErrCodeRoleNotFound       = 3001 // 角色不存在
	ErrCodeRoleCodeExists     = 3002 // 角色编码已存在
	ErrCodeRoleInUse          = 3003 // 角色已被用户关联，不能删除
	ErrCodeSystemRoleNoDelete = 3004 // 系统内置角色不可删除

	// 权限相关
	ErrCodePermissionNotFound       = 3005 // 权限不存在
	ErrCodePermissionCodeExists     = 3006 // 权限编码已存在
	ErrCodeSystemPermissionNoDelete = 3007 // 系统内置权限不可删除
	ErrCodePermissionHasChildren    = 3008 // 权限有子节点，不能删除

	// 部门相关
	ErrCodeDeptNotFound    = 3009 // 部门不存在
	ErrCodeDeptCodeExists  = 3010 // 部门编码已存在
	ErrCodeDeptHasChildren = 3011 // 部门有子部门，不能删除

	// 职位相关
	ErrCodePositionNotFound   = 3012 // 职位不存在
	ErrCodePositionCodeExists = 3013 // 职位编码已存在

	// 数据权限相关
	ErrCodeInvalidDataScope = 3014 // 无效的数据权限范围

	// 权限校验
	ErrCodePermissionDenied = 3015 // 权限不足
)

// NewRoleNotFoundError 角色不存在错误
func NewRoleNotFoundError() *response.AppError {
	return response.NewError(ErrCodeRoleNotFound, "角色不存在")
}

// NewRoleCodeExistsError 角色编码已存在错误
func NewRoleCodeExistsError(code string) *response.AppError {
	return response.NewError(ErrCodeRoleCodeExists, "角色编码已存在: "+code)
}

// NewRoleInUseError 角色已被关联错误
func NewRoleInUseError() *response.AppError {
	return response.NewError(ErrCodeRoleInUse, "角色已关联用户，不能删除")
}

// NewSystemRoleNoDeleteError 系统内置角色不可删除错误
func NewSystemRoleNoDeleteError() *response.AppError {
	return response.NewError(ErrCodeSystemRoleNoDelete, "系统内置角色不可删除")
}

// NewPermissionNotFoundError 权限不存在错误
func NewPermissionNotFoundError() *response.AppError {
	return response.NewError(ErrCodePermissionNotFound, "权限不存在")
}

// NewPermissionCodeExistsError 权限编码已存在错误
func NewPermissionCodeExistsError(code string) *response.AppError {
	return response.NewError(ErrCodePermissionCodeExists, "权限编码已存在: "+code)
}

// NewSystemPermissionNoDeleteError 系统内置权限不可删除错误
func NewSystemPermissionNoDeleteError() *response.AppError {
	return response.NewError(ErrCodeSystemPermissionNoDelete, "系统内置权限不可删除")
}

// NewPermissionHasChildrenError 权限有子节点错误
func NewPermissionHasChildrenError() *response.AppError {
	return response.NewError(ErrCodePermissionHasChildren, "权限有子节点，不能删除")
}

// NewDeptNotFoundError 部门不存在错误
func NewDeptNotFoundError() *response.AppError {
	return response.NewError(ErrCodeDeptNotFound, "部门不存在")
}

// NewDeptCodeExistsError 部门编码已存在错误
func NewDeptCodeExistsError(code string) *response.AppError {
	return response.NewError(ErrCodeDeptCodeExists, "部门编码已存在: "+code)
}

// NewDeptHasChildrenError 部门有子部门错误
func NewDeptHasChildrenError() *response.AppError {
	return response.NewError(ErrCodeDeptHasChildren, "部门有子部门，不能删除")
}

// NewPositionNotFoundError 职位不存在错误
func NewPositionNotFoundError() *response.AppError {
	return response.NewError(ErrCodePositionNotFound, "职位不存在")
}

// NewPositionCodeExistsError 职位编码已存在错误
func NewPositionCodeExistsError(code string) *response.AppError {
	return response.NewError(ErrCodePositionCodeExists, "职位编码已存在: "+code)
}

// NewInvalidDataScopeError 无效的数据权限范围错误
func NewInvalidDataScopeError(scope string) *response.AppError {
	return response.NewError(ErrCodeInvalidDataScope, "无效的数据权限范围: "+scope)
}

// NewPermissionDeniedError 权限不足错误
func NewPermissionDeniedError() *response.AppError {
	return response.NewError(ErrCodePermissionDenied, "权限不足")
}
