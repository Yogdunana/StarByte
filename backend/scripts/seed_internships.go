package main

import (
	"fmt"

	"gorm.io/gorm"
)

func seedInternships(db *gorm.DB) error {
	if err := db.Exec(`
		UPDATE users u
		SET department_id = d.id
		FROM departments d
		WHERE d.code = 'tech' AND u.username IN ('admin', 'test') AND u.department_id IS NULL
	`).Error; err != nil {
		return fmt.Errorf("assign internship seed department: %w", err)
	}

	stmts := []string{
		`
		INSERT INTO internships (
			id, user_id, department_id, start_date, end_date, status,
			title, organization, description, type, skills, achievements, created_by
		)
		SELECT uuid_generate_v4(), u.id, d.id, DATE '2026-07-01', NULL, 0,
			'StarByte 后端开发实习', '计算机协会技术部',
			'参与招新系统与协会平台后端开发', 0,
			'["Go","PostgreSQL","Docker"]', '', u.id
		FROM users u
		LEFT JOIN departments d ON d.code = 'tech'
		WHERE u.username = 'admin'
		AND NOT EXISTS (
			SELECT 1 FROM internships i WHERE i.user_id = u.id AND i.title = 'StarByte 后端开发实习'
		)
		`,
		`
		INSERT INTO internships (
			id, user_id, department_id, start_date, end_date, status,
			title, organization, description, type, skills, achievements, report, created_by
		)
		SELECT uuid_generate_v4(), u.id, d.id, DATE '2026-03-01', DATE '2026-06-30', 1,
			'宣传部招新运营实习', '计算机协会宣传部',
			'协助招新宣发与面试接待', 0,
			'["运营","文案"]', '完成招新海报与社群运营', '已完成招新宣发复盘', u.id
		FROM users u
		LEFT JOIN departments d ON d.code = 'publicity'
		WHERE u.username = 'test'
		AND NOT EXISTS (
			SELECT 1 FROM internships i WHERE i.user_id = u.id AND i.title = '宣传部招新运营实习'
		)
		`,
	}
	for _, sql := range stmts {
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("seed internships: %w", err)
		}
	}
	return nil
}
