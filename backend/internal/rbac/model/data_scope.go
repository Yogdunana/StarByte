package model

// 数据权限范围常量
const (
	DataScopeAll              = "all"                // 全部数据
	DataScopeDepartment       = "department"         // 本部门数据
	DataScopeDepartmentAndSub = "department_and_sub" // 本部门及下级部门
	DataScopeSelf             = "self"               // 仅自己的数据
	DataScopeCustom           = "custom"             // 自定义（指定部门）
)

// IsValidDataScope 检查数据权限范围是否有效
func IsValidDataScope(scope string) bool {
	switch scope {
	case DataScopeAll, DataScopeDepartment, DataScopeDepartmentAndSub, DataScopeSelf, DataScopeCustom:
		return true
	default:
		return false
	}
}
