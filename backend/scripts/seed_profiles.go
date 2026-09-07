package main

import (
	"fmt"

	"gorm.io/gorm"
)

type seedProfileRow struct {
	Username   string
	StudentNo  string
	RealName   string
	Grade      string
	Major      string
	DeptCode   string
	PosCode    string
	MemberType int
}

var seedProfileData = []seedProfileRow{
	{
		Username: "admin", StudentNo: "20210001", RealName: "管理员",
		Grade: "2021", Major: "计算机科学与技术",
		DeptCode: "project", PosCode: "president", MemberType: 4,
	},
	{
		Username: "test", StudentNo: "20210002", RealName: "测试会员",
		Grade: "2021", Major: "软件工程",
		DeptCode: "brand", PosCode: "officer", MemberType: 1,
	},
}

func seedMemberProfiles(db *gorm.DB) error {
	for _, row := range seedProfileData {
		if err := db.Exec(`
			INSERT INTO member_profiles (
				id, user_id, real_name, student_no, gender, grade, major,
				department_id, position_id, member_type, status, contact_email,
				skills, projects, bio
			)
			SELECT uuid_generate_v4(), u.id, ?, ?, 1, ?, ?, d.id, p.id, ?, 0, u.email,
				'[]'::jsonb, '[]'::jsonb, ''
			FROM users u
			LEFT JOIN departments d ON d.code = ?
			LEFT JOIN positions p ON p.code = ?
			WHERE u.username = ?
			  AND NOT EXISTS (SELECT 1 FROM member_profiles mp WHERE mp.user_id = u.id)
		`, row.RealName, row.StudentNo, row.Grade, row.Major, row.MemberType,
			row.DeptCode, row.PosCode, row.Username,
		).Error; err != nil {
			return fmt.Errorf("insert profile %s: %w", row.Username, err)
		}
		if err := db.Exec(`
			UPDATE member_profiles mp
			SET
				student_no = ?,
				real_name = ?,
				grade = CASE WHEN mp.grade = '' THEN ? ELSE mp.grade END,
				major = CASE WHEN mp.major = '' THEN ? ELSE mp.major END,
				department_id = COALESCE(mp.department_id, d.id),
				position_id = COALESCE(mp.position_id, pos.id),
				updated_at = CURRENT_TIMESTAMP
			FROM users u
			LEFT JOIN departments d ON d.code = ?
			LEFT JOIN positions pos ON pos.code = ?
			WHERE mp.user_id = u.id AND u.username = ?
		`, row.StudentNo, row.RealName, row.Grade, row.Major, row.DeptCode, row.PosCode, row.Username,
		).Error; err != nil {
			return fmt.Errorf("backfill profile %s: %w", row.Username, err)
		}
	}
	return nil
}
