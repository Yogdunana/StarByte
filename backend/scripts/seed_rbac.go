package main

import (
	"fmt"

	"gorm.io/gorm"
)

type namedCode struct {
	Name        string
	Code        string
	Description string
	Sort        int
	IsSystem    bool
}

var seedRolesData = []namedCode{
	{Name: "超级管理员", Code: "super_admin", Description: "系统内置超管", Sort: 0, IsSystem: true},
	{Name: "社长", Code: "president", Description: "协会社长", Sort: 1, IsSystem: true},
	{Name: "副社长", Code: "vice_president", Description: "协会副社长", Sort: 2},
	{Name: "部长", Code: "minister", Description: "部门部长", Sort: 3},
	{Name: "副部长", Code: "vice_minister", Description: "部门副部长", Sort: 4},
	{Name: "干事", Code: "officer", Description: "部门干事", Sort: 5},
	{Name: "会员", Code: "member", Description: "普通会员", Sort: 6},
}

var seedDepartmentsData = []namedCode{
	{Name: "技术部", Code: "tech", Sort: 1},
	{Name: "活动部", Code: "activity", Sort: 2},
	{Name: "宣传部", Code: "publicity", Sort: 3},
	{Name: "外联部", Code: "liaison", Sort: 4},
}

var seedPositionsData = []namedCode{
	{Name: "社长", Code: "president", Sort: 1},
	{Name: "副社长", Code: "vice_president", Sort: 2},
	{Name: "部长", Code: "minister", Sort: 3},
	{Name: "副部长", Code: "vice_minister", Sort: 4},
	{Name: "干事", Code: "officer", Sort: 5},
}

type seedPerm struct {
	Name     string
	Code     string
	Resource string
	Action   string
}

func moduleCRUD(resource, label string) []seedPerm {
	return []seedPerm{
		{Name: label + "查看", Code: resource + ":read", Resource: resource, Action: "read"},
		{Name: label + "创建", Code: resource + ":create", Resource: resource, Action: "create"},
		{Name: label + "更新", Code: resource + ":update", Resource: resource, Action: "update"},
		{Name: label + "删除", Code: resource + ":delete", Resource: resource, Action: "delete"},
	}
}

func allSeedPermissions() []seedPerm {
	perms := []seedPerm{}
	perms = append(perms, moduleCRUD("user", "用户")...)
	perms = append(perms, moduleCRUD("role", "角色")...)
	perms = append(perms, moduleCRUD("department", "部门")...)
	perms = append(perms, moduleCRUD("position", "职位")...)
	perms = append(perms, moduleCRUD("member", "会员")...)
	perms = append(perms,
		seedPerm{Name: "会员审核", Code: "member:approve", Resource: "member", Action: "approve"},
		seedPerm{Name: "档案导出", Code: "member:export", Resource: "member", Action: "export"},
		seedPerm{Name: "会员管理", Code: "member:manage", Resource: "member", Action: "manage"},
	)
	perms = append(perms, moduleCRUD("interview", "面试")...)
	perms = append(perms,
		seedPerm{Name: "面试管理", Code: "interview:manage", Resource: "interview", Action: "manage"},
		seedPerm{Name: "面试评分", Code: "interview:evaluate", Resource: "interview", Action: "evaluate"},
	)
	perms = append(perms, moduleCRUD("meeting", "会议")...)
	perms = append(perms, seedPerm{Name: "会议管理", Code: "meeting:manage", Resource: "meeting", Action: "manage"})
	perms = append(perms, moduleCRUD("task", "任务")...)
	perms = append(perms,
		seedPerm{Name: "任务分配", Code: "task:assign", Resource: "task", Action: "assign"},
		seedPerm{Name: "任务转办", Code: "task:transfer", Resource: "task", Action: "transfer"},
		seedPerm{Name: "任务评论", Code: "task:comment", Resource: "task", Action: "comment"},
	)
	perms = append(perms, moduleCRUD("internship", "实习")...)
	perms = append(perms, moduleCRUD("workflow", "流程")...)
	perms = append(perms,
		seedPerm{Name: "权限查看", Code: "permission:read", Resource: "permission", Action: "read"},
		seedPerm{Name: "统计查看", Code: "stats:read", Resource: "stats", Action: "read"},
		seedPerm{Name: "系统配置", Code: "system:config", Resource: "system", Action: "config"},
		seedPerm{Name: "通知模板查看", Code: "notification:template:read", Resource: "notification", Action: "read"},
		seedPerm{Name: "发送通知", Code: "notification:send", Resource: "notification", Action: "create"},
	)
	return perms
}

func seedRoles(db *gorm.DB) error {
	for _, r := range seedRolesData {
		if err := db.Exec(`
			INSERT INTO roles (id, name, code, description, sort_order, status, is_system)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, 0, ?)
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				sort_order = EXCLUDED.sort_order,
				is_system = EXCLUDED.is_system`,
			r.Name, r.Code, r.Description, r.Sort, r.IsSystem,
		).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedPermissions(db *gorm.DB) error {
	for _, p := range allSeedPermissions() {
		if err := db.Exec(`
			INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, 3, true, 0)
			ON CONFLICT (code) DO NOTHING`,
			p.Name, p.Code, p.Resource, p.Action, p.Name,
		).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedDepartments(db *gorm.DB) error {
	for _, d := range seedDepartmentsData {
		if err := db.Exec(`
			INSERT INTO departments (id, name, code, sort_order, status)
			VALUES (uuid_generate_v4(), ?, ?, ?, 0)
			ON CONFLICT (code) DO NOTHING`,
			d.Name, d.Code, d.Sort,
		).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedPositions(db *gorm.DB) error {
	for i, p := range seedPositionsData {
		if err := db.Exec(`
			INSERT INTO positions (id, name, code, level, vote_weight, sort_order, status)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, 0)
			ON CONFLICT (code) DO NOTHING`,
			p.Name, p.Code, 10-i, float64(5-i), p.Sort,
		).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedRolePermissions(db *gorm.DB) error {
	if err := db.Exec(`
		INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
		SELECT uuid_generate_v4(), r.id, p.id, 'all'
		FROM roles r
		CROSS JOIN permissions p
		WHERE r.code IN ('president', 'super_admin')
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("assign all perms to president: %w", err)
	}

	// 副社长：除系统配置外的全部权限
	if err := db.Exec(`
		INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
		SELECT uuid_generate_v4(), r.id, p.id, 'all'
		FROM roles r CROSS JOIN permissions p
		WHERE r.code = 'vice_president' AND p.code <> 'system:config'
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("assign vice_president perms: %w", err)
	}

	// 部长：全部 read + 业务 create/update
	if err := db.Exec(`
		INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
		SELECT uuid_generate_v4(), r.id, p.id, 'department'
		FROM roles r CROSS JOIN permissions p
		WHERE r.code = 'minister' AND (
			p.action = 'read'
			OR (p.resource IN ('member','interview','meeting','task','internship','file','workflow','notification')
			    AND p.action IN ('create','update'))
			OR (p.resource = 'member' AND p.action IN ('approve','export','manage'))
			OR (p.resource = 'interview' AND p.action IN ('manage','evaluate'))
			OR (p.resource = 'meeting' AND p.action = 'manage')
			OR (p.resource = 'task' AND p.action IN ('assign','transfer','comment'))
		)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("assign minister perms: %w", err)
	}

	// 副部长：全部 read + 业务 create
	if err := db.Exec(`
		INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
		SELECT uuid_generate_v4(), r.id, p.id, 'department'
		FROM roles r CROSS JOIN permissions p
		WHERE r.code = 'vice_minister' AND (
			p.action = 'read'
			OR (p.resource IN ('member','interview','meeting','task','internship','file')
			    AND p.action = 'create')
			OR (p.resource = 'task' AND p.action IN ('update','comment'))
		)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("assign vice_minister perms: %w", err)
	}

	if err := assignPermCodes(db, "officer", "department", officerPermCodes()); err != nil {
		return err
	}
	return assignPermCodes(db, "member", "self", memberPermCodes())
}

func officerPermCodes() []string {
	return []string{
		"user:read", "member:read", "member:create",
		"interview:read", "interview:evaluate", "meeting:read",
		"task:read", "task:create", "task:update", "task:comment",
		"file:read", "file:create",
		"internship:read", "internship:create",
	}
}

func memberPermCodes() []string {
	return []string{
		"user:read", "member:read", "meeting:read", "task:read",
		"file:read", "file:create", "internship:read",
	}
}

func assignPermCodes(db *gorm.DB, roleCode, dataScope string, codes []string) error {
	for _, code := range codes {
		if err := db.Exec(`
			INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
			SELECT uuid_generate_v4(), r.id, p.id, ?
			FROM roles r
			CROSS JOIN permissions p
			WHERE r.code = ? AND p.code = ?
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`, dataScope, roleCode, code).Error; err != nil {
			return fmt.Errorf("assign %s perm %s: %w", roleCode, code, err)
		}
	}
	return nil
}
