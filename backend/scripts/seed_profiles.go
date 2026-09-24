package main

import (
	"fmt"

	membermodel "github.com/Yogdunana/StarByte/backend/internal/member/model"
	"gorm.io/gorm"
)

type seedProfileRow struct {
	Username   string
	StudentNo  string
	RealName   string
	Gender     int // 0=未知 1=男 2=女，与 users.gender 同一套取值
	Grade      string
	Major      string
	DeptCode   string
	PosCode    string
	MemberType int
}

// seedProfileData 是演示/内置账号的档案基线。
//
// 注意 admin 这一行的语义：它是【技术账号】，不是协会会员。
// 所以刻意留空学号、不挂部门（d.code='' 匹配不到行 → department_id 为 NULL），
// member_type 用 MemberTypeAdmin(5)，职位指向“系统管理员”。
// 之前这里是 学号 20210001 / 部门 project / 职位 president / 类型 4（会长），
// 会让管理员在人员档案里冒充会长、并在会长类业务里被算作在任会员。
var seedProfileData = []seedProfileRow{
	{
		Username: "admin", StudentNo: "", RealName: "管理员", Gender: 0,
		Grade: "", Major: "",
		DeptCode: "", PosCode: "system_admin", MemberType: int(membermodel.MemberTypeAdmin),
	},
	{
		Username: "test", StudentNo: "20210002", RealName: "测试会员", Gender: 1,
		Grade: "2021", Major: "软件工程",
		DeptCode: "brand", PosCode: "officer", MemberType: int(membermodel.MemberTypeMember),
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
			SELECT uuid_generate_v4(), u.id, ?, ?, ?, ?, ?, d.id, p.id, ?, 0, u.email,
				'[]'::jsonb, '[]'::jsonb, ''
			FROM users u
			LEFT JOIN departments d ON d.code = ?
			LEFT JOIN positions p ON p.code = ?
			WHERE u.username = ?
			  AND NOT EXISTS (SELECT 1 FROM member_profiles mp WHERE mp.user_id = u.id)
		`, row.RealName, row.StudentNo, row.Gender, row.Grade, row.Major, row.MemberType,
			row.DeptCode, row.PosCode, row.Username,
		).Error; err != nil {
			return fmt.Errorf("insert profile %s: %w", row.Username, err)
		}
		// 第二段是「收敛」而不是「补空」：这几行的归属权在种子手里，
		// 每次 bootstrap 都要把它们拉回上面声明的形态。
		// 之前用 COALESCE(mp.department_id, d.id) 只补不清，结果 admin 一旦被
		// 种成「项目部/会长」就再也改不回去（重跑种子也不清）。这里改成直接赋值。
		if err := db.Exec(`
			UPDATE member_profiles mp
			SET
				student_no = ?,
				real_name = ?,
				gender = ?,
				grade = ?,
				major = ?,
				department_id = d.id,
				position_id = pos.id,
				member_type = ?,
				updated_at = CURRENT_TIMESTAMP
			FROM users u
			LEFT JOIN departments d ON d.code = ?
			LEFT JOIN positions pos ON pos.code = ?
			WHERE mp.user_id = u.id AND u.username = ?
			  AND NOT (
			        mp.student_no IS NOT DISTINCT FROM ?
			    AND mp.real_name IS NOT DISTINCT FROM ?
			    AND mp.gender IS NOT DISTINCT FROM ?
			    AND mp.grade IS NOT DISTINCT FROM ?
			    AND mp.major IS NOT DISTINCT FROM ?
			    AND mp.department_id IS NOT DISTINCT FROM d.id
			    AND mp.position_id IS NOT DISTINCT FROM pos.id
			    AND mp.member_type IS NOT DISTINCT FROM ?
			  )
		`, row.StudentNo, row.RealName, row.Gender, row.Grade, row.Major, row.MemberType,
			row.DeptCode, row.PosCode, row.Username,
			row.StudentNo, row.RealName, row.Gender, row.Grade, row.Major, row.MemberType,
		).Error; err != nil {
			return fmt.Errorf("backfill profile %s: %w", row.Username, err)
		}
	}
	return nil
}
