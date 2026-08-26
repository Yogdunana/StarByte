package rbac

// Error 是 RBAC 模块的领域错误类型，不依赖表现层的 response 包。
type Error struct {
	code    int
	message string
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	return e.message
}

// Code 返回错误码。
func (e *Error) Code() int {
	return e.code
}

// Message 返回错误消息。
func (e *Error) Message() string {
	return e.message
}

// Is 实现 errors.Is 接口，基于错误码判断是否为同一类错误。
// 这样可以使用 errors.Is(err, NewRoleNotFoundError()) 进行错误类型判断。
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.code == e.code
}

// RBAC 模块错误码（3000-3999）
//
// 错误码段分配说明：
//   - 3001-3004: 角色相关（不存在、编码重复、已使用、系统内置不可删除）
//   - 3005-3008: 权限相关（不存在、编码重复、系统内置不可删除、有子节点）
//   - 3009-3011: 部门相关（不存在、编码重复、有子节点）
//   - 3012-3013: 职位相关（不存在、编码重复）
//   - 3014:      数据权限相关（无效范围）
//   - 3015:      权限校验（权限不足）
//   - 3016-3020: 扩展错误码（系统角色不可编辑、系统权限不可编辑、部门已使用、职位已使用、权限已禁用）
const (
	// ===== 角色相关 (3001-3004, 3016) =====
	ErrCodeRoleNotFound       = 3001 // 角色不存在
	ErrCodeRoleCodeExists     = 3002 // 角色编码已存在
	ErrCodeRoleInUse          = 3003 // 角色已被用户关联，不能删除
	ErrCodeSystemRoleNoDelete = 3004 // 系统内置角色不可删除
	ErrCodeSystemRoleNoEdit   = 3016 // 系统内置角色不可编辑权限

	// ===== 权限相关 (3005-3008, 3017, 3020) =====
	ErrCodePermissionNotFound       = 3005 // 权限不存在
	ErrCodePermissionCodeExists     = 3006 // 权限编码已存在
	ErrCodeSystemPermissionNoDelete = 3007 // 系统内置权限不可删除
	ErrCodePermissionHasChildren    = 3008 // 权限有子节点，不能删除
	ErrCodeSystemPermissionNoEdit   = 3017 // 系统内置权限不可编辑状态
	ErrCodePermissionDisabled       = 3020 // 权限已禁用，不能分配

	// ===== 部门相关 (3009-3011, 3018) =====
	ErrCodeDeptNotFound    = 3009 // 部门不存在
	ErrCodeDeptCodeExists  = 3010 // 部门编码已存在
	ErrCodeDeptHasChildren = 3011 // 部门有子部门，不能删除
	ErrCodeDeptInUse       = 3018 // 部门已被用户关联，不能删除

	// ===== 职位相关 (3012-3013, 3019) =====
	ErrCodePositionNotFound   = 3012 // 职位不存在
	ErrCodePositionCodeExists = 3013 // 职位编码已存在
	ErrCodePositionInUse      = 3019 // 职位已被用户关联，不能删除

	// ===== 数据权限相关 (3014) =====
	ErrCodeInvalidDataScope = 3014 // 无效的数据权限范围

	// ===== 权限校验 (3015) =====
	ErrCodePermissionDenied = 3015 // 权限不足
)

// NewRoleNotFoundError 角色不存在错误
func NewRoleNotFoundError() *Error {
	return &Error{code: ErrCodeRoleNotFound, message: "角色不存在"}
}

// NewRoleCodeExistsError 角色编码已存在错误
func NewRoleCodeExistsError(code string) *Error {
	return &Error{code: ErrCodeRoleCodeExists, message: "角色编码已存在: " + code}
}

// NewRoleInUseError 角色已被关联错误
func NewRoleInUseError() *Error {
	return &Error{code: ErrCodeRoleInUse, message: "角色已关联用户，不能删除"}
}

// NewSystemRoleNoDeleteError 系统内置角色不可删除错误
func NewSystemRoleNoDeleteError() *Error {
	return &Error{code: ErrCodeSystemRoleNoDelete, message: "系统内置角色不可删除"}
}

// NewSystemRoleNoEditError 系统内置角色不可编辑权限错误
func NewSystemRoleNoEditError() *Error {
	return &Error{code: ErrCodeSystemRoleNoEdit, message: "系统内置角色不可编辑权限"}
}

// NewPermissionNotFoundError 权限不存在错误
func NewPermissionNotFoundError() *Error {
	return &Error{code: ErrCodePermissionNotFound, message: "权限不存在"}
}

// NewPermissionCodeExistsError 权限编码已存在错误
func NewPermissionCodeExistsError(code string) *Error {
	return &Error{code: ErrCodePermissionCodeExists, message: "权限编码已存在: " + code}
}

// NewSystemPermissionNoDeleteError 系统内置权限不可删除错误
func NewSystemPermissionNoDeleteError() *Error {
	return &Error{code: ErrCodeSystemPermissionNoDelete, message: "系统内置权限不可删除"}
}

// NewPermissionHasChildrenError 权限有子节点错误
func NewPermissionHasChildrenError() *Error {
	return &Error{code: ErrCodePermissionHasChildren, message: "权限有子节点，不能删除"}
}

// NewSystemPermissionNoEditError 系统内置权限不可编辑错误
func NewSystemPermissionNoEditError() *Error {
	return &Error{code: ErrCodeSystemPermissionNoEdit, message: "系统内置权限不可编辑状态"}
}

// NewPermissionDisabledError 权限已禁用错误
func NewPermissionDisabledError(code string) *Error {
	return &Error{code: ErrCodePermissionDisabled, message: "权限已禁用，不能分配: " + code}
}

// NewDeptNotFoundError 部门不存在错误
func NewDeptNotFoundError() *Error {
	return &Error{code: ErrCodeDeptNotFound, message: "部门不存在"}
}

// NewDeptCodeExistsError 部门编码已存在错误
func NewDeptCodeExistsError(code string) *Error {
	return &Error{code: ErrCodeDeptCodeExists, message: "部门编码已存在: " + code}
}

// NewDeptHasChildrenError 部门有子部门错误
func NewDeptHasChildrenError() *Error {
	return &Error{code: ErrCodeDeptHasChildren, message: "部门有子部门，不能删除"}
}

// NewDeptInUseError 部门已被用户关联错误
func NewDeptInUseError() *Error {
	return &Error{code: ErrCodeDeptInUse, message: "部门已关联用户，不能删除"}
}

// NewPositionNotFoundError 职位不存在错误
func NewPositionNotFoundError() *Error {
	return &Error{code: ErrCodePositionNotFound, message: "职位不存在"}
}

// NewPositionCodeExistsError 职位编码已存在错误
func NewPositionCodeExistsError(code string) *Error {
	return &Error{code: ErrCodePositionCodeExists, message: "职位编码已存在: " + code}
}

// NewPositionInUseError 职位已被用户关联错误
func NewPositionInUseError() *Error {
	return &Error{code: ErrCodePositionInUse, message: "职位已关联用户，不能删除"}
}

// NewInvalidDataScopeError 无效的数据权限范围错误
func NewInvalidDataScopeError(scope string) *Error {
	return &Error{code: ErrCodeInvalidDataScope, message: "无效的数据权限范围: " + scope}
}

// NewPermissionDeniedError 权限不足错误
func NewPermissionDeniedError() *Error {
	return &Error{code: ErrCodePermissionDenied, message: "权限不足"}
}
