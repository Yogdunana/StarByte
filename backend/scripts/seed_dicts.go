package main

import (
	"fmt"

	"gorm.io/gorm"
)

type seedDictType struct {
	Code        string
	Name        string
	Description string
	Sort        int
	Items       []seedDictItem
}

type seedDictItem struct {
	Value string
	Label string
	Sort  int
}

var seedDictTypesData = []seedDictType{
	{Code: "applicant_type", Name: "申请类型", Description: "入会申请：会员 / 干事", Sort: 1, Items: []seedDictItem{
		{"1", "会员", 1}, {"2", "干事", 2},
	}},
	{Code: "interview_result", Name: "面试结果", Description: "面试记录结论", Sort: 2, Items: []seedDictItem{
		{"0", "未出", 0}, {"1", "通过", 1}, {"2", "不通过", 2}, {"3", "待定", 3},
	}},
	{Code: "task_status", Name: "任务状态", Description: "任务流转状态", Sort: 3, Items: []seedDictItem{
		{"0", "待处理", 0}, {"1", "进行中", 1}, {"2", "已完成", 2}, {"3", "已取消", 3}, {"4", "已挂起", 4},
	}},
	{Code: "task_priority", Name: "任务优先级", Description: "任务优先级", Sort: 4, Items: []seedDictItem{
		{"0", "低", 0}, {"1", "中", 1}, {"2", "高", 2}, {"3", "紧急", 3},
	}},
	{Code: "internship_status", Name: "实习状态", Description: "IT 实习记录状态", Sort: 5, Items: []seedDictItem{
		{"0", "进行中", 0}, {"1", "已完成", 1}, {"2", "已中止", 2},
	}},
	{Code: "internship_type", Name: "实习类型", Description: "IT 实习类型", Sort: 6, Items: []seedDictItem{
		{"0", "校内社团", 0}, {"1", "校内其他", 1}, {"2", "校外实习", 2},
	}},
	{Code: "meeting_status", Name: "会议状态", Description: "会议生命周期", Sort: 7, Items: []seedDictItem{
		{"0", "待开始", 0}, {"1", "进行中", 1}, {"2", "已结束", 2}, {"3", "已取消", 3},
	}},
}

func seedDicts(db *gorm.DB) error {
	for _, typ := range seedDictTypesData {
		if err := db.Exec(`
			INSERT INTO dict_types (id, code, name, description, sort_order, status, is_system)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, 0, true)
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				sort_order = EXCLUDED.sort_order,
				is_system = true`,
			typ.Code, typ.Name, typ.Description, typ.Sort,
		).Error; err != nil {
			return fmt.Errorf("seed dict type %s: %w", typ.Code, err)
		}
		for _, item := range typ.Items {
			if err := db.Exec(`
				INSERT INTO dict_items (id, type_id, item_value, item_label, sort_order, status)
				SELECT uuid_generate_v4(), t.id, ?, ?, ?, 0
				FROM dict_types t WHERE t.code = ?
				ON CONFLICT (type_id, item_value) DO UPDATE SET
					item_label = EXCLUDED.item_label,
					sort_order = EXCLUDED.sort_order`,
				item.Value, item.Label, item.Sort, typ.Code,
			).Error; err != nil {
				return fmt.Errorf("seed dict item %s/%s: %w", typ.Code, item.Value, err)
			}
		}
	}
	return nil
}
