package main

import (
	"fmt"

	"gorm.io/gorm"
)

// seedDeptNode 章程组织节点：三大中心为根，七大职能部门挂中心。
type seedDeptNode struct {
	Name        string
	Code        string
	ParentCode  string
	Description string
	Sort        int
}

var seedCentersData = []seedDeptNode{
	{Name: "技术研发中心", Code: "center_rd", Description: "副社长兼任主任；下辖创新创业部、项目开发部、课程研发部", Sort: 1},
	{Name: "运营管理中心", Code: "center_ops", Description: "副社长兼任主任；下辖综合行政部、品牌传播部", Sort: 2},
	{Name: "对外合作中心", Code: "center_ext", Description: "副社长兼任主任；下辖战略关系部、赛事运营部", Sort: 3},
}

var seedDepartmentsData = []seedDeptNode{
	{Name: "创新创业部", Code: "innovation", ParentCode: "center_rd", Description: "大创等赛事组织、项目孵化、竞赛培训", Sort: 11},
	{Name: "项目开发部", Code: "project", ParentCode: "center_rd", Description: "软件系统设计开发、代码规范、技术资产管理", Sort: 12},
	{Name: "课程研发部", Code: "curriculum", ParentCode: "center_rd", Description: "课程体系规划开发、教学实施、课程库建设", Sort: 13},
	{Name: "综合行政部", Code: "admin", ParentCode: "center_ops", Description: "成员招新、档案管理、换届组织、内部协调", Sort: 21},
	{Name: "品牌传播部", Code: "brand", ParentCode: "center_ops", Description: "品牌建设、新媒体运营、宣传物料制作", Sort: 22},
	{Name: "战略关系部", Code: "strategy", ParentCode: "center_ext", Description: "对外合作、赞助拓展、厂商对接、项目承接", Sort: 31},
	{Name: "赛事运营部", Code: "events", ParentCode: "center_ext", Description: "竞赛组织、活动策划、赛事执行、学术交流", Sort: 32},
}

// seedLegacyDeptRemaps 旧四部原地改码，保留 UUID，避免用户/申请/任务外键断裂。
var seedLegacyDeptRemaps = []struct{ Old, New string }{
	{Old: "tech", New: "project"},
	{Old: "activity", New: "events"},
	{Old: "publicity", New: "brand"},
	{Old: "liaison", New: "strategy"},
}

func seedDepartments(db *gorm.DB) error {
	if err := upsertSeedCenters(db); err != nil {
		return err
	}
	if err := remapLegacyDepartments(db); err != nil {
		return err
	}
	return upsertSeedDepartments(db)
}

func upsertSeedCenters(db *gorm.DB) error {
	for _, c := range seedCentersData {
		if err := db.Exec(`
			INSERT INTO departments (id, name, code, description, sort_order, status)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, 0)
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				parent_id = NULL,
				sort_order = EXCLUDED.sort_order,
				status = 0,
				updated_at = CURRENT_TIMESTAMP`,
			c.Name, c.Code, c.Description, c.Sort,
		).Error; err != nil {
			return fmt.Errorf("upsert center %s: %w", c.Code, err)
		}
	}
	return nil
}

func remapLegacyDepartments(db *gorm.DB) error {
	byCode := map[string]seedDeptNode{}
	for _, d := range seedDepartmentsData {
		byCode[d.Code] = d
	}
	for _, m := range seedLegacyDeptRemaps {
		target, ok := byCode[m.New]
		if !ok {
			return fmt.Errorf("legacy remap %s -> missing target %s", m.Old, m.New)
		}
		if err := db.Exec(`
			UPDATE departments d
			SET
				code = ?,
				name = ?,
				description = ?,
				parent_id = p.id,
				sort_order = ?,
				status = 0,
				updated_at = CURRENT_TIMESTAMP
			FROM departments p
			WHERE d.code = ?
			  AND p.code = ?
			  AND NOT EXISTS (SELECT 1 FROM departments x WHERE x.code = ?)`,
			target.Code, target.Name, target.Description, target.Sort,
			m.Old, target.ParentCode, target.Code,
		).Error; err != nil {
			return fmt.Errorf("remap department %s -> %s: %w", m.Old, m.New, err)
		}
	}
	return nil
}

func upsertSeedDepartments(db *gorm.DB) error {
	for _, d := range seedDepartmentsData {
		if err := db.Exec(`
			INSERT INTO departments (id, name, code, parent_id, description, sort_order, status)
			SELECT uuid_generate_v4(), ?, ?, p.id, ?, ?, 0
			FROM departments p
			WHERE p.code = ?
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				parent_id = EXCLUDED.parent_id,
				sort_order = EXCLUDED.sort_order,
				status = 0,
				updated_at = CURRENT_TIMESTAMP`,
			d.Name, d.Code, d.Description, d.Sort, d.ParentCode,
		).Error; err != nil {
			return fmt.Errorf("upsert department %s: %w", d.Code, err)
		}
	}
	return nil
}
