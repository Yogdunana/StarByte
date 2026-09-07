package main

import (
	"fmt"

	"gorm.io/gorm"
)

func seedOps(db *gorm.DB) error {
	cats := []struct {
		name, code string
		dir        int
		desc       string
	}{
		{"会费收入", "dues", 2, "会员会费"},
		{"赞助收入", "sponsor", 2, "企业/活动赞助"},
		{"活动支出", "activity", 1, "活动举办费用"},
		{"采购支出", "purchase", 1, "物资采购"},
		{"其他收入", "other_in", 2, "其他收入"},
		{"其他支出", "other_out", 1, "其他支出"},
	}
	for _, c := range cats {
		if err := db.Exec(`
			INSERT INTO finance_categories (id, name, code, direction, description, status)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, 0)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description
		`, c.name, c.code, c.dir, c.desc).Error; err != nil {
			return fmt.Errorf("seed finance category %s: %w", c.code, err)
		}
	}

	tpls := []struct{ name, code, content string }{
		{"赞助合同", "sponsor", "赞助合同模板：甲方提供赞助，乙方提供曝光与活动执行。"},
		{"活动合同", "event", "活动合同模板：约定场地、档期与双方责任。"},
		{"采购合同", "purchase", "采购合同模板：约定物资规格、数量与验收。"},
		{"通用合同", "other", "通用合同模板。"},
	}
	for _, t := range tpls {
		if err := db.Exec(`
			INSERT INTO contract_templates (id, name, code, content, status)
			VALUES (uuid_generate_v4(), ?, ?, ?, 0)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, content = EXCLUDED.content
		`, t.name, t.code, t.content).Error; err != nil {
			return fmt.Errorf("seed contract template %s: %w", t.code, err)
		}
	}

	if err := db.Exec(`
		INSERT INTO scheduler_tasks (
			id, name, code, cron_expr, timezone, handler_key, payload, depends_on,
			status, max_retries, timeout_sec, next_run_at
		)
		SELECT
			uuid_generate_v4(), '合同到期扫描', 'contract_expiry',
			'0 0 9 * * *', 'Asia/Shanghai', 'contract_expiry', '', '[]',
			0, 3, 60, NOW() + INTERVAL '1 day'
		WHERE NOT EXISTS (
			SELECT 1 FROM scheduler_tasks WHERE code = 'contract_expiry' AND status <> 2
		)
	`).Error; err != nil {
		return fmt.Errorf("seed contract expiry job: %w", err)
	}
	return nil
}
