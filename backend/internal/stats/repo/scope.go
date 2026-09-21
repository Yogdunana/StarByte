package repo

import (
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"gorm.io/gorm"
)

// applyDeptScope 将部门数据范围条件下推到 Overview 的组织级聚合计数查询。
//
// 过滤规则（与 DataScopeCondition 语义一致）：
//   - scope == nil 或 scope.IsEmpty()：不加任何过滤（超管 / 全部数据范围）。
//   - scope.IsSelf：仅本人范围的用户不应看到全组织总量，组织级聚合一律置 0。
//   - scope.Query == "1 = 0"（非 IsSelf）：同样拒绝，聚合置 0。
//   - 部门范围：scope.Query 形如 "department_id IN ?"，Args 为 []uuid.UUID，按 deptExpr 套用。
//
// deptExpr 为调用方传入的、受信的 SQL 条件片段（含 `?` 占位符），例如：
//   - "department_id IN ?"（member_profiles / tasks / internships / member_applications）
//   - "organizer_id IN (SELECT id FROM users WHERE department_id IN ?)"（meetings：按组织者部门过滤）
//
// deptExpr 来自代码常量而非外部输入，且部门 ID 一律通过 scope.Args 参数化传递，
// 不存在 SQL 注入风险。
func applyDeptScope(q *gorm.DB, scope *rbacModel.DataScopeCondition, deptExpr string) *gorm.DB {
	if scope == nil || scope.IsEmpty() {
		return q
	}
	if scope.IsSelf {
		return q.Where("1 = 0")
	}
	if scope.Query == "1 = 0" {
		return q.Where("1 = 0")
	}
	return q.Where(deptExpr, scope.Args...)
}
