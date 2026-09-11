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
	{Name: "副部长", Code: "vice_minister", Description: "部门副部长（章程未单列，系统预留）", Sort: 4},
	{Name: "干事", Code: "officer", Description: "部门干事", Sort: 5},
	{Name: "会员", Code: "member", Description: "普通会员", Sort: 6},
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
	perms = append(perms, moduleCRUD("schedule", "日程")...)
	perms = append(perms, seedPerm{Name: "实习评价", Code: "internship:evaluate", Resource: "internship", Action: "evaluate"})
	perms = append(perms, moduleCRUD("workflow", "流程")...)
	perms = append(perms, moduleCRUD("dict", "字典")...)
	perms = append(perms,
		seedPerm{Name: "权限查看", Code: "permission:read", Resource: "permission", Action: "read"},
		seedPerm{Name: "统计查看", Code: "stats:read", Resource: "stats", Action: "read"},
		seedPerm{Name: "统计导出", Code: "stats:export", Resource: "stats", Action: "export"},
		seedPerm{Name: "系统配置", Code: "system:config", Resource: "system", Action: "config"},
		seedPerm{Name: "通知模板查看", Code: "notification:template:read", Resource: "notification", Action: "read"},
		seedPerm{Name: "发送通知", Code: "notification:send", Resource: "notification", Action: "create"},
	)
	perms = append(perms, moduleCRUD("config", "运行时配置")...)
	perms = append(perms,
		seedPerm{Name: "会话查看", Code: "session:read", Resource: "session", Action: "read"},
		seedPerm{Name: "强制下线", Code: "session:delete", Resource: "session", Action: "delete"},
	)
	perms = append(perms,
		seedPerm{Name: "导出Excel", Code: "export:excel", Resource: "export", Action: "excel"},
		seedPerm{Name: "导出CSV", Code: "export:csv", Resource: "export", Action: "csv"},
		seedPerm{Name: "导出PDF", Code: "export:pdf", Resource: "export", Action: "pdf"},
		seedPerm{Name: "导出JSON", Code: "export:json", Resource: "export", Action: "json"},
		seedPerm{Name: "模板打印", Code: "export:template", Resource: "export", Action: "template"},
		seedPerm{Name: "导出查看", Code: "export:read", Resource: "export", Action: "read"},
		seedPerm{Name: "导出下载", Code: "export:download", Resource: "export", Action: "download"},
	)
	perms = append(perms,
		seedPerm{Name: "缓存查看", Code: "cache:read", Resource: "cache", Action: "read"},
		seedPerm{Name: "缓存清除", Code: "cache:delete", Resource: "cache", Action: "delete"},
		seedPerm{Name: "缓存预热", Code: "cache:manage", Resource: "cache", Action: "manage"},
	)
	perms = append(perms,
		seedPerm{Name: "调度查看", Code: "scheduler:read", Resource: "scheduler", Action: "read"},
		seedPerm{Name: "调度创建", Code: "scheduler:create", Resource: "scheduler", Action: "create"},
		seedPerm{Name: "调度更新", Code: "scheduler:update", Resource: "scheduler", Action: "update"},
		seedPerm{Name: "调度删除", Code: "scheduler:delete", Resource: "scheduler", Action: "delete"},
		seedPerm{Name: "调度执行", Code: "scheduler:run", Resource: "scheduler", Action: "run"},
		seedPerm{Name: "调度启停", Code: "scheduler:manage", Resource: "scheduler", Action: "manage"},
	)
	perms = append(perms,
		seedPerm{Name: "统一搜索", Code: "search:read", Resource: "search", Action: "read"},
	)
	perms = append(perms,
		seedPerm{Name: "财务查看", Code: "finance:read", Resource: "finance", Action: "read"},
		seedPerm{Name: "财务创建", Code: "finance:create", Resource: "finance", Action: "create"},
		seedPerm{Name: "财务管理", Code: "finance:manage", Resource: "finance", Action: "manage"},
		seedPerm{Name: "处分查看", Code: "discipline:read", Resource: "discipline", Action: "read"},
		seedPerm{Name: "处分登记", Code: "discipline:create", Resource: "discipline", Action: "create"},
		seedPerm{Name: "处分审批", Code: "discipline:approve", Resource: "discipline", Action: "approve"},
		seedPerm{Name: "处分撤销", Code: "discipline:revoke", Resource: "discipline", Action: "revoke"},
		seedPerm{Name: "合同查看", Code: "contract:read", Resource: "contract", Action: "read"},
		seedPerm{Name: "合同创建", Code: "contract:create", Resource: "contract", Action: "create"},
		seedPerm{Name: "合同管理", Code: "contract:manage", Resource: "contract", Action: "manage"},
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
 AND (r.code = 'president' OR p.resource <> 'interview_private')
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("assign all perms to president: %w", err)
	}

	// 副社长：可读运行时配置，但不能改/删（与 system:config 对齐，避免绕过实习/投票开关）
	if err := db.Exec(`
		INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
		SELECT uuid_generate_v4(), r.id, p.id, CASE WHEN p.resource = 'interview_private' THEN 'department_and_sub' ELSE 'all' END
		FROM roles r CROSS JOIN permissions p
		WHERE r.code = 'vice_president'
		  AND p.code NOT IN ('system:config', 'config:create', 'config:update', 'config:delete', 'cache:delete', 'cache:manage', 'scheduler:create', 'scheduler:update', 'scheduler:delete', 'scheduler:run', 'scheduler:manage')
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("assign vice_president perms: %w", err)
	}
	if err := db.Exec(`
		DELETE FROM role_permissions rp
		USING roles r, permissions p
		WHERE rp.role_id = r.id AND rp.permission_id = p.id
		  AND r.code = 'vice_president'
		  AND p.code IN ('config:create', 'config:update', 'config:delete', 'cache:delete', 'cache:manage', 'scheduler:create', 'scheduler:update', 'scheduler:delete', 'scheduler:run', 'scheduler:manage')
	`).Error; err != nil {
		return fmt.Errorf("revoke vice_president config writes: %w", err)
	}

	// 部长：全部 read + 业务 create/update
	if err := db.Exec(`
		INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
		SELECT uuid_generate_v4(), r.id, p.id, 'department'
		FROM roles r CROSS JOIN permissions p
		WHERE r.code = 'minister' AND (
			p.action = 'read'
			OR (p.resource IN ('member','interview','meeting','task','internship','schedule','file','workflow','notification','finance','discipline','contract')
			    AND p.action IN ('create','update'))
			OR (p.resource = 'member' AND p.action IN ('approve','export','manage'))
			OR (p.resource = 'interview' AND p.action IN ('manage','evaluate'))
			OR (p.resource = 'meeting' AND p.action = 'manage')
			OR (p.resource = 'task' AND p.action IN ('assign','transfer','comment'))
			OR (p.resource = 'internship' AND p.action = 'evaluate')
			OR (p.resource = 'finance' AND p.action = 'manage')
			OR (p.resource = 'discipline' AND p.action IN ('approve','revoke'))
			OR (p.resource = 'contract' AND p.action = 'manage')
			OR (p.resource = 'export' AND p.action IN ('excel','csv','pdf','json','template','download'))
			OR (p.resource = 'stats' AND p.action = 'export')
			OR (p.resource = 'form' AND p.action IN ('write', 'submit'))
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
		WHERE r.code = 'vice_minister' AND p.resource <> 'interview_private' AND (
			p.action = 'read'
			OR (p.resource IN ('member','interview','meeting','task','internship','schedule','file','finance','discipline','contract')
			    AND p.action = 'create')
			OR (p.resource = 'task' AND p.action IN ('update','comment'))
			OR (p.resource = 'form' AND p.action IN ('write', 'submit'))
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
		"internship:read", "internship:create", "internship:update", "internship:delete",
		"schedule:read", "schedule:create", "schedule:update", "schedule:delete",
		"form:read", "form:submit",
		"discipline:read",
	}
}

func memberPermCodes() []string {
	return []string{
		"user:read", "member:read", "meeting:read", "task:read",
		"file:read", "file:create", "internship:read", "internship:create", "internship:update", "internship:delete",
		"schedule:read", "schedule:create", "schedule:update", "schedule:delete",
		"form:submit", "discipline:read",
	}
}

func vicePresidentExcludedPerms() []string {
	return []string{
		"system:config", "config:create", "config:update", "config:delete",
		"cache:delete", "cache:manage",
		"scheduler:create", "scheduler:update", "scheduler:delete", "scheduler:run", "scheduler:manage",
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
