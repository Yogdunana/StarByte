package service

import (
	"strings"
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

// 空 / nil scope 不限制，返回 nil。
func TestActivityScopeClause_NoScope(t *testing.T) {
	if got := activityScopeClause(nil, uuid.New()); got != nil {
		t.Errorf("nil scope: got %v, want nil", got)
	}
	if got := activityScopeClause(&rbacModel.DataScopeCondition{}, uuid.New()); got != nil {
		t.Errorf("empty scope: got %v, want nil", got)
	}
}

// IsSelf 且 organizer == viewer 允许，否则拒绝。
func TestActivityScopeClause_Self(t *testing.T) {
	viewer := uuid.New()
	got := activityScopeClause(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, viewer)
	if got == nil {
		t.Fatal("IsSelf scope: got nil")
	}
	if got.Query != "a.organizer_id = ?" {
		t.Errorf("IsSelf query = %q, want 'a.organizer_id = ?'", got.Query)
	}
	if len(got.Args) != 1 || got.Args[0] != viewer {
		t.Errorf("IsSelf args = %v, want [%s]", got.Args, viewer)
	}

	if !canAccessActivity(got, viewer, &viewer, viewer) {
		t.Error("self: organizer==viewer should be allowed")
	}
	other := uuid.New()
	if canAccessActivity(got, other, &other, viewer) {
		t.Error("self: organizer!=viewer should be denied")
	}
}

// 拒绝条件（1 = 0，非 self）原样返回并拒绝访问。
func TestActivityScopeClause_Denied(t *testing.T) {
	deny := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	got := activityScopeClause(deny, uuid.New())
	if got == nil || got.Query != "1 = 0" || got.IsSelf {
		t.Errorf("deny scope: got %+v, want 1=0 non-self", got)
	}
	if canAccessActivity(got, uuid.New(), nil, uuid.New()) {
		t.Error("denied scope should reject all")
	}
}

// 部门范围命中 / 不命中。
func TestActivityScopeClause_DepartmentHit(t *testing.T) {
	dept := uuid.New()
	otherDept := uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{dept}}}
	got := activityScopeClause(scope, uuid.New())
	if got == nil {
		t.Fatal("department scope: got nil")
	}
	if !strings.Contains(got.Query, "EXISTS") || !strings.Contains(got.Query, "a.organizer_id") {
		t.Errorf("department query = %q, want EXISTS + a.organizer_id", got.Query)
	}
	if !strings.Contains(got.Query, "su.department_id IN ?") {
		t.Errorf("department query = %q, want su.department_id IN ?", got.Query)
	}

	// 命中
	if !canAccessActivity(got, uuid.New(), &dept, uuid.New()) {
		t.Error("department hit should allow")
	}
	// 不命中
	if canAccessActivity(got, uuid.New(), &otherDept, uuid.New()) {
		t.Error("department miss should deny")
	}
}

// 部门范围，但组织者部门无法解析（nil）应拒绝（fail-closed）。
func TestActivityScopeClause_DepartmentNilDept(t *testing.T) {
	dept := uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{dept}}}
	got := activityScopeClause(scope, uuid.New())
	if canAccessActivity(got, uuid.New(), nil, uuid.New()) {
		t.Error("nil organizer dept should be denied under department scope")
	}
}

// 部门范围 Args 为单个 uuid.UUID（非切片）也应命中。
func TestActivityScopeClause_DepartmentSingleUUID(t *testing.T) {
	dept := uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{dept}}
	got := activityScopeClause(scope, uuid.New())
	if !canAccessActivity(got, uuid.New(), &dept, uuid.New()) {
		t.Error("single uuid.UUID arg should hit")
	}
}
