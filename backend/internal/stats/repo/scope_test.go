package repo

import (
	"strings"
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newDryDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(postgres.Open("host=localhost dbname=dry user=dry"),
		&gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry db: %v", err)
	}
	return db
}

// TestApplyDeptScope_NilAndEmpty nil / 空 scope 不加任何额外过滤。
func TestApplyDeptScope_NilAndEmpty(t *testing.T) {
	db := newDryDB(t)
	var total int64

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return applyDeptScope(tx.Table("member_profiles").Where("status = 0"), nil, "department_id IN ?").Count(&total)
	})
	assert.NotContains(t, strings.ToLower(sql), "1 = 0")
	assert.NotContains(t, strings.ToLower(sql), "department_id in")

	sql = db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return applyDeptScope(tx.Table("member_profiles").Where("status = 0"), &rbacModel.DataScopeCondition{}, "department_id IN ?").Count(&total)
	})
	assert.NotContains(t, strings.ToLower(sql), "1 = 0")
	assert.NotContains(t, strings.ToLower(sql), "department_id in")
}

// TestApplyDeptScope_IsSelf IsSelf → 生成 1 = 0。
func TestApplyDeptScope_IsSelf(t *testing.T) {
	db := newDryDB(t)
	var total int64
	scope := &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return applyDeptScope(tx.Table("member_profiles").Where("status = 0"), scope, "department_id IN ?").Count(&total)
	})
	assert.Contains(t, strings.ToLower(sql), "1 = 0")
}

// TestApplyDeptScope_Department 部门范围 → SQL 含 department_id IN，会议表含 organizer_id 子查询。
func TestApplyDeptScope_Department(t *testing.T) {
	db := newDryDB(t)
	var total int64
	d1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	d2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	scope := &rbacModel.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{d1, d2}}}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return applyDeptScope(tx.Table("member_profiles").Where("status = 0"), scope, "department_id IN ?").Count(&total)
	})
	assert.Contains(t, strings.ToLower(sql), "department_id")

	sql = db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return applyDeptScope(tx.Table("meetings").Where("start_time >= ?", "x"), scope,
			"organizer_id IN (SELECT id FROM users WHERE department_id IN ?)").Count(&total)
	})
	assert.Contains(t, strings.ToLower(sql), "organizer_id")
	assert.Contains(t, strings.ToLower(sql), "select")
	assert.Contains(t, strings.ToLower(sql), "department_id")
}

// TestApplyDeptScope_Denied "1 = 0"（非 IsSelf）→ 保持拒绝。
func TestApplyDeptScope_Denied(t *testing.T) {
	db := newDryDB(t)
	var total int64
	scope := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return applyDeptScope(tx.Table("member_profiles").Where("status = 0"), scope, "department_id IN ?").Count(&total)
	})
	assert.Contains(t, strings.ToLower(sql), "1 = 0")
}
