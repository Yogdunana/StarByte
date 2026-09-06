package model

// DataScopeCondition 表示数据权限过滤的 SQL 条件
// 当 Query 为空时表示不附加任何过滤（用户可访问全部数据，如超级管理员或 all 范围）
// 当 Query 为 "1 = 0" 时表示不允许访问任何记录
// 所有条件均通过 Args 参数化传递，避免 SQL 注入风险
type DataScopeCondition struct {
	Query string
	Args  []interface{}
	// IsSelf 为 true 时表示 data_scope=self；Query 仍为 "1 = 0"，由能表达“仅本人”的业务改写。
	// 其他 "1 = 0" 表示拒绝全部（无范围、部门无法解析、fail-closed）。
	IsSelf bool
}

// IsEmpty 返回是否为空条件（不限制数据范围）
func (c *DataScopeCondition) IsEmpty() bool {
	return c == nil || c.Query == ""
}

// 数据权限范围常量
const (
	DataScopeAll              = "all"
	DataScopeDepartment       = "department"
	DataScopeDepartmentAndSub = "department_and_sub"
	DataScopeSelf             = "self"
	DataScopeCustom           = "custom"
)
