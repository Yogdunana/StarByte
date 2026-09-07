package identity

import (
	"context"

	authsvc "github.com/Yogdunana/StarByte/backend/internal/auth/service"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/google/uuid"
)

// Lookup 用人员档案实现认证模块的学号/身份查询，避免 auth 直接依赖 member。
type Lookup struct {
	profiles repo.ProfileRepo
}

// NewLookup 创建档案身份查询。
func NewLookup(profiles repo.ProfileRepo) *Lookup {
	return &Lookup{profiles: profiles}
}

// GetUserIDByStudentNo 按学号解析 user_id；未找到返回 uuid.Nil。
func (l *Lookup) GetUserIDByStudentNo(ctx context.Context, studentNo string) (uuid.UUID, error) {
	if l == nil || l.profiles == nil {
		return uuid.Nil, nil
	}
	p, err := l.profiles.GetByStudentNo(ctx, studentNo, nil)
	if err != nil {
		return uuid.Nil, err
	}
	if p == nil {
		return uuid.Nil, nil
	}
	return p.UserID, nil
}

// GetByUserID 读取档案中的学号、姓名、年级、专业和部门职位名。
func (l *Lookup) GetByUserID(ctx context.Context, userID uuid.UUID) (*authsvc.MemberIdentity, error) {
	if l == nil || l.profiles == nil {
		return nil, nil
	}
	row, err := l.profiles.GetByUserIDWithNames(ctx, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	ident := &authsvc.MemberIdentity{
		StudentNo:      row.StudentNo,
		RealName:       row.RealName,
		Grade:          row.Grade,
		Major:          row.Major,
		DepartmentName: row.DepartmentName,
		PositionName:   row.PositionName,
	}
	if row.DepartmentID != nil {
		ident.DepartmentID = row.DepartmentID.String()
	}
	if row.PositionID != nil {
		ident.PositionID = row.PositionID.String()
	}
	return ident, nil
}
