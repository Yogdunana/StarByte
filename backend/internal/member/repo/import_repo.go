package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

// ImportUserRow 是批量录入时需要读写的 users 行。
//
// 这里不直接复用 user 模块的模型：member 只需要这张表的几个字段，
// 而且 user/model 一旦将来引用 member 的东西就会成环。
type ImportUserRow struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"password_hash"`
	RealName     string     `json:"real_name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Gender       int        `json:"gender"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Status       int        `json:"status"`
}

// MemberImportRepo 支撑批量录入的数据库读写。
type MemberImportRepo interface {
	DepartmentByCodeOrName(ctx context.Context, value string) (*rbacModel.Department, error)
	RoleIDByCode(ctx context.Context, code string) (uuid.UUID, error)
	UserByUsername(ctx context.Context, username string) (*ImportUserRow, error)
	UserByID(ctx context.Context, id uuid.UUID) (*ImportUserRow, error)
	HasUserRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error)
	CreateUser(ctx context.Context, row *ImportUserRow) error
	SetUserDepartment(ctx context.Context, userID uuid.UUID, department *uuid.UUID) error
	ReplaceUserRole(ctx context.Context, userID, roleID uuid.UUID, scopes []uuid.UUID) error
	QueuePermissionRefresh(ctx context.Context, userID uuid.UUID) error
	DueImportedProbations(ctx context.Context, now time.Time) ([]ImportedProbation, error)
	ClearProbationUntil(ctx context.Context, userID uuid.UUID) error
}

type memberImportRepo struct{ db *gorm.DB }

// NewMemberImportRepo 创建批量录入仓库。
func NewMemberImportRepo(db *gorm.DB) MemberImportRepo {
	return &memberImportRepo{db: db}
}

// DepartmentByCodeOrName 按部门编码或名称定位一个启用中的部门。
// 编码优先，其次名称；重名时返回多条中的第一条，并在调用方按数量提示。
func (r *memberImportRepo) DepartmentByCodeOrName(ctx context.Context, value string) (*rbacModel.Department, error) {
	var department rbacModel.Department
	err := r.db.WithContext(ctx).Where("status = 0 AND lower(code) = lower(?)", value).Take(&department).Error
	if err == nil {
		return &department, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	err = r.db.WithContext(ctx).Where("status = 0 AND name = ?", value).Take(&department).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &department, nil
}

func (r *memberImportRepo) RoleIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.WithContext(ctx).Table("roles").Where("code = ? AND status = 0", code).Limit(1).Pluck("id", &id).Error
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *memberImportRepo) UserByUsername(ctx context.Context, username string) (*ImportUserRow, error) {
	var row ImportUserRow
	err := r.db.WithContext(ctx).Table("users").
		Select("id, username, password_hash, real_name, email, phone, gender, department_id, status").
		Where("username = ? AND deleted_at IS NULL", username).
		Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *memberImportRepo) UserByID(ctx context.Context, id uuid.UUID) (*ImportUserRow, error) {
	var row ImportUserRow
	err := r.db.WithContext(ctx).Table("users").
		Select("id, username, password_hash, real_name, email, phone, gender, department_id, status").
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// HasUserRole 判断这个人是不是已经有这个角色了，用来区分「更新」和「无变化」。
// 过期未续的角色不算拥有，会被当成要重新写一次。
func (r *memberImportRepo) HasUserRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("user_roles").
		Where("user_id = ? AND role_id = ? AND expired_at IS NULL", userID, roleID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *memberImportRepo) CreateUser(ctx context.Context, row *ImportUserRow) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("users").Create(map[string]interface{}{
		"id":            row.ID,
		"username":      row.Username,
		"password_hash": row.PasswordHash,
		"real_name":     row.RealName,
		"email":         row.Email,
		"phone":         row.Phone,
		"gender":        row.Gender,
		"department_id": row.DepartmentID,
		"status":        0,
		"created_at":    now,
		"updated_at":    now,
	}).Error
}

func (r *memberImportRepo) SetUserDepartment(ctx context.Context, userID uuid.UUID, department *uuid.UUID) error {
	return r.db.WithContext(ctx).Table("users").Where("id = ?", userID).
		Update("department_id", department).Error
}

// ReplaceUserRole 写入替任：先 upsert user_roles，再把任职范围整表替换成 scopes。
//
// 这一步不能省。admission_maintenance.GrantRole 只插 user_roles，
// 不写 user_role_departments，带部门范围的工作流节点因此永远选不到人，
// 任命了部长却没人能审批就是这么来的。
func (r *memberImportRepo) ReplaceUserRole(ctx context.Context, userID, roleID uuid.UUID, scopes []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"expired_at": nil}),
	}).Table("user_roles").Create(map[string]interface{}{
		"id":      uuid.New(),
		"user_id": userID,
		"role_id": roleID,
	}).Error; err != nil {
		return err
	}
	var assignment struct{ ID uuid.UUID }
	if err := r.db.WithContext(ctx).Table("user_roles").
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Select("id").Take(&assignment).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Exec("DELETE FROM user_role_departments WHERE user_role_id = ?", assignment.ID).Error; err != nil {
		return err
	}
	for _, dept := range scopes {
		if err := r.db.WithContext(ctx).Exec(
			"INSERT INTO user_role_departments(user_role_id, department_id) VALUES (?,?) ON CONFLICT DO NOTHING",
			assignment.ID, dept).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *memberImportRepo) QueuePermissionRefresh(ctx context.Context, userID uuid.UUID) error {
	return NewAdmissionJobsRepo(r.db).QueuePermissionRefresh(ctx, userID)
}

// ImportedProbation 是到期该转正式干事的批量录入人员。
type ImportedProbation struct {
	UserID       uuid.UUID
	DepartmentID *uuid.UUID
}

// DueImportedProbations 列出预备期已经届满、档案还在预备期的人。
// 只挑 probation_until 有值且小于等于 now 的，正常入会流程的申请不受影响。
func (r *memberImportRepo) DueImportedProbations(ctx context.Context, now time.Time) ([]ImportedProbation, error) {
	var rows []ImportedProbation
	err := r.db.WithContext(ctx).Table("member_profiles").
		Select("user_id, department_id").
		Where("status = ? AND probation_until IS NOT NULL AND probation_until <= ?", model.ProfileProbation, now).
		Order("probation_until, user_id").
		Limit(100).
		Scan(&rows).Error
	return rows, err
}

func (r *memberImportRepo) ClearProbationUntil(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("member_profiles").
		Where("user_id = ?", userID).Update("probation_until", nil).Error
}
