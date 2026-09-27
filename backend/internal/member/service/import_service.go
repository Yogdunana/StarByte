package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
)

// ProbationMonths 是预备期长度，跟现有入会流程保持一致。
const ProbationMonths = 1

// MemberImportService 批量录入成员。
//
// 设计前提：这些人从来没在系统里注册过，也不需要再走进会审批。
// 「注册」和「录用」在这里一次做完：账号 + 档案 + 角色同时落库，学号是关键。
// CAS 第一次登录时会按学号回查 member_profiles（见 auth/service/cas.go 的
// resolveCASUser），命中就自动绑定身份并以该账号登录 —— 所以只要录进来的学号
// 跟统一认证一致，之后给他们发一条登录链接就够了。
type MemberImportService interface {
	Preview(ctx context.Context, canAppoint bool, req *dto.MemberImportRequest) (*dto.MemberImportResponse, error)
	Import(ctx context.Context, canAppoint bool, req *dto.MemberImportRequest) (*dto.MemberImportResponse, error)
}

type memberImportService struct {
	db       *gorm.DB
	profs    repo.ProfileRepo
	importer repo.MemberImportRepo
	now      func() time.Time
}

// NewMemberImportService 创建批量录入服务。
func NewMemberImportService(db *gorm.DB, profs repo.ProfileRepo) MemberImportService {
	return &memberImportService{
		db:       db,
		profs:    profs,
		importer: repo.NewMemberImportRepo(db),
		now:      time.Now,
	}
}

// importRowPlan 是一行的归因结果，预览和落库共用。
type importRowPlan struct {
	row     dto.MemberImportRow
	dept    *rbacModel.Department
	roleID  uuid.UUID
	user    *repo.ImportUserRow
	profile *model.MemberProfile
	create  bool // 需要新建 users
	hasRole bool
	err     string
	changes []string
}

// Preview 只做归因不写库，用来给前端「先看一眼再确认」。
func (s *memberImportService) Preview(ctx context.Context, canAppoint bool, req *dto.MemberImportRequest) (*dto.MemberImportResponse, error) {
	rows, err := s.plan(ctx, canAppoint, req)
	if err != nil {
		return nil, err
	}
	return summarize(rows), nil
}

// Import 逐行落库。每一行一个独立事务：一个人填错了不影响同批其他人。
func (s *memberImportService) Import(ctx context.Context, canAppoint bool, req *dto.MemberImportRequest) (*dto.MemberImportResponse, error) {
	rows, err := s.plan(ctx, canAppoint, req)
	if err != nil {
		return nil, err
	}
	result := &dto.MemberImportResponse{Total: len(rows), Results: make([]dto.MemberImportRowResult, 0, len(rows))}
	for i := range rows {
		plan := &rows[i]
		out := summarizePlan(plan)
		if plan.err == "" {
			if writeErr := s.writeRow(ctx, plan); writeErr != nil {
				// 归因通过但写库失败，只作废这一行。
				out.Status = dto.ImportRowFailed
				out.Message = writeErr.Error()
			} else {
				out.Status, out.Message = planVerdict(plan)
			}
		}
		result.Results = append(result.Results, out)
		switch out.Status {
		case dto.ImportRowCreated:
			result.Created++
		case dto.ImportRowUpdated:
			result.Updated++
		case dto.ImportRowSkipped:
			result.Skipped++
		default:
			result.Failed++
		}
	}
	return result, nil
}

func summarize(rows []importRowPlan) *dto.MemberImportResponse {
	result := &dto.MemberImportResponse{Total: len(rows), Results: make([]dto.MemberImportRowResult, 0, len(rows))}
	for i := range rows {
		out := summarizePlan(&rows[i])
		result.Results = append(result.Results, out)
		switch out.Status {
		case dto.ImportRowCreated:
			result.Created++
		case dto.ImportRowUpdated:
			result.Updated++
		case dto.ImportRowSkipped:
			result.Skipped++
		default:
			result.Failed++
		}
	}
	return result
}

func planVerdict(plan *importRowPlan) (dto.MemberImportRowStatus, string) {
	if plan.create || plan.profile == nil {
		return dto.ImportRowCreated, "新建账号与档案"
	}
	if len(plan.changes) > 0 {
		return dto.ImportRowUpdated, "更新：" + strings.Join(plan.changes, "、")
	}
	return dto.ImportRowSkipped, "内容一致，未重复写入"
}

func summarizePlan(plan *importRowPlan) dto.MemberImportRowResult {
	out := dto.MemberImportRowResult{
		Row:        plan.row.Row,
		StudentNo:  plan.row.StudentNo,
		RealName:   plan.row.RealName,
		Department: plan.row.Department,
		Role:       plan.row.Role,
	}
	if plan.err != "" {
		out.Status = dto.ImportRowFailed
		out.Message = plan.err
		return out
	}
	out.Status, out.Message = planVerdict(plan)
	return out
}

// plan 把每一行翻译成「要写什么」。单行问题归因到 plan.err，不会让整批报废；
// 返回 error 只用于请求本身不合法（超量）或数据库故障。
func (s *memberImportService) plan(ctx context.Context, canAppoint bool, req *dto.MemberImportRequest) ([]importRowPlan, error) {
	if req == nil || len(req.Rows) == 0 {
		return nil, nil
	}
	if len(req.Rows) > MaxImportRows {
		return nil, fmt.Errorf("一次最多录入 %d 人，请分批导入", MaxImportRows)
	}
	plans := make([]importRowPlan, 0, len(req.Rows))
	seen := make(map[string]int, len(req.Rows))
	for index := range req.Rows {
		row := req.Rows[index]
		if row.Row <= 0 {
			row.Row = index + 1
		}
		plan := importRowPlan{row: row}
		rule, msg := pureImportChecks(&row, seen, canAppoint)
		plan.row = row
		if msg != "" {
			plan.err = msg
			plans = append(plans, plan)
			continue
		}
		seen[row.StudentNo] = row.Row
		if row.Department != "" {
			dept, err := s.importer.DepartmentByCodeOrName(ctx, row.Department)
			if err != nil {
				return nil, err
			}
			if dept == nil {
				plan.err = "找不到启用中的部门或中心：" + row.Department
				plans = append(plans, plan)
				continue
			}
			if rule.Office == officeScopeCenter && dept.ParentID != nil {
				plan.err = dept.Name + " 是职能部门，中心副主任请填所在中心"
				plans = append(plans, plan)
				continue
			}
			if rule.Office == officeScopeDepartment && dept.ParentID == nil {
				plan.err = dept.Name + " 是中心，部长请填具体部门"
				plans = append(plans, plan)
				continue
			}
			plan.dept = dept
		}

		roleID, err := s.importer.RoleIDByCode(ctx, row.Role)
		if err != nil {
			return nil, err
		}
		if roleID == uuid.Nil {
			plan.err = "角色不存在或已禁用：" + row.Role
			plans = append(plans, plan)
			continue
		}
		plan.roleID = roleID

		profile, err := s.profs.GetByStudentNo(ctx, row.StudentNo, nil)
		if err != nil {
			return nil, err
		}
		plan.profile = profile

		if profile != nil {
			if profile.Status == model.ProfileDisabled {
				plan.err = "档案已停用，请先在会员档案页处理"
				plans = append(plans, plan)
				continue
			}
			user, err := s.importer.UserByID(ctx, profile.UserID)
			if err != nil {
				return nil, err
			}
			if user == nil {
				plan.err = "档案指向的账号不存在"
				plans = append(plans, plan)
				continue
			}
			if user.Status != 0 {
				plan.err = "账号不可用（已禁用或锁定）"
				plans = append(plans, plan)
				continue
			}
			plan.user = user
		} else {
			// 档案是唯一的身份凭据，所以按学号找账号：找得到就复用，
			// 找不到才新建，重跑同一份名单不会造重复账号。
			user, err := s.importer.UserByUsername(ctx, row.StudentNo)
			if err != nil {
				return nil, err
			}
			if user != nil && user.Status != 0 {
				plan.err = "同名账号已存在且不可用"
				plans = append(plans, plan)
				continue
			}
			if user == nil {
				user = &repo.ImportUserRow{ID: uuid.New(), Username: row.StudentNo}
				plan.create = true
			}
			plan.user = user
		}

		hasRole, err := s.importer.HasUserRole(ctx, plan.user.ID, roleID)
		if err != nil {
			return nil, err
		}
		plan.hasRole = hasRole
		plan.changes = plannedChanges(&plan, rule)
		plans = append(plans, plan)
	}
	return plans, nil
}

// plannedChanges 列出这次会改哪些字段，用来区分「更新」和「无变化」。
func plannedChanges(plan *importRowPlan, rule importRoleRule) []string {
	if plan.profile == nil {
		return []string{"新建档案"}
	}
	changes := []string{}
	profile, row := plan.profile, plan.row
	if profile.RealName != row.RealName {
		changes = append(changes, "姓名")
	}
	if row.Gender != nil && profile.Gender != *row.Gender {
		changes = append(changes, "性别")
	}
	if row.Grade != "" && profile.Grade != row.Grade {
		changes = append(changes, "年级")
	}
	if row.Major != "" && profile.Major != row.Major {
		changes = append(changes, "专业")
	}
	if row.Phone != "" && profile.ContactPhone != row.Phone {
		changes = append(changes, "手机号")
	}
	if row.Email != "" && profile.ContactEmail != row.Email {
		changes = append(changes, "邮箱")
	}
	if plan.dept != nil && (profile.DepartmentID == nil || *profile.DepartmentID != plan.dept.ID) {
		changes = append(changes, "部门")
	}
	if profile.MemberType < rule.MemberType {
		changes = append(changes, "成员类型")
	}
	if profileStatusRank(importProfileStatus(row.Role)) > profileStatusRank(profile.Status) {
		changes = append(changes, "状态")
	}
	if !plan.hasRole {
		changes = append(changes, "角色")
	}
	return changes
}

// profileStatusRank 给档案状态排序，避免把已经是正式成员的人拉回预备期。
func profileStatusRank(status int16) int {
	switch status {
	case model.ProfileActive:
		return 3
	case model.ProfileProbation:
		return 2
	case model.ProfileLeft:
		return 1
	default:
		return 0
	}
}

// writeRow 在独立事务里写完一个人。
func (s *memberImportService) writeRow(ctx context.Context, plan *importRowPlan) error {
	rule, ok := importRoleRuleOf(plan.row.Role)
	if !ok {
		return fmt.Errorf("不支持的角色：%s", plan.row.Role)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		importer := repo.NewMemberImportRepo(tx)
		profs := repo.NewProfileRepo(tx)
		now := s.now()

		if plan.create {
			hash, err := randomPasswordHash()
			if err != nil {
				return err
			}
			plan.user.PasswordHash = hash
			plan.user.RealName = plan.row.RealName
			plan.user.Email = plan.row.Email
			plan.user.Phone = plan.row.Phone
			plan.user.Status = 0
			if plan.row.Gender != nil {
				plan.user.Gender = int(*plan.row.Gender)
			}
			if err := importer.CreateUser(ctx, plan.user); err != nil {
				return err
			}
		}
		if plan.dept != nil {
			if err := importer.SetUserDepartment(ctx, plan.user.ID, &plan.dept.ID); err != nil {
				return err
			}
		}

		targetStatus := importProfileStatus(plan.row.Role)
		if plan.profile == nil {
			profile := &model.MemberProfile{
				ID:           uuid.New(),
				UserID:       plan.user.ID,
				RealName:     plan.row.RealName,
				StudentNo:    plan.row.StudentNo,
				Grade:        plan.row.Grade,
				Major:        plan.row.Major,
				MemberType:   rule.MemberType,
				Status:       targetStatus,
				ContactPhone: plan.row.Phone,
				ContactEmail: plan.row.Email,
				Skills:       model.JSONStrings{},
				Projects:     model.JSONProjects{},
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if plan.row.Gender != nil {
				profile.Gender = *plan.row.Gender
			}
			if plan.dept != nil {
				id := plan.dept.ID
				profile.DepartmentID = &id
			}
			join := now
			profile.JoinDate = &join
			if targetStatus == model.ProfileProbation {
				until := now.AddDate(0, ProbationMonths, 0)
				profile.ProbationUntil = &until
			}
			if err := profs.Create(ctx, profile); err != nil {
				return err
			}
			plan.profile = profile
		} else {
			profile := plan.profile
			profile.RealName = plan.row.RealName
			if plan.row.Gender != nil {
				profile.Gender = *plan.row.Gender
			}
			if plan.row.Grade != "" {
				profile.Grade = plan.row.Grade
			}
			if plan.row.Major != "" {
				profile.Major = plan.row.Major
			}
			if plan.row.Phone != "" {
				profile.ContactPhone = plan.row.Phone
			}
			if plan.row.Email != "" {
				profile.ContactEmail = plan.row.Email
			}
			if plan.dept != nil {
				id := plan.dept.ID
				profile.DepartmentID = &id
			}
			if profile.MemberType < rule.MemberType {
				profile.MemberType = rule.MemberType
			}
			if profileStatusRank(targetStatus) > profileStatusRank(profile.Status) {
				profile.Status = targetStatus
			}
			if profile.JoinDate == nil {
				join := now
				profile.JoinDate = &join
			}
			if profile.Skills == nil {
				profile.Skills = model.JSONStrings{}
			}
			if profile.Projects == nil {
				profile.Projects = model.JSONProjects{}
			}
			profile.UpdatedAt = now
			if err := profs.Update(ctx, profile); err != nil {
				return err
			}
		}

		scopes := []uuid.UUID{}
		if plan.dept != nil && (rule.Office == officeScopeCenter || rule.Office == officeScopeDepartment) {
			scopes = append(scopes, plan.dept.ID)
		}
		if err := importer.ReplaceUserRole(ctx, plan.user.ID, plan.roleID, scopes); err != nil {
			return err
		}
		return importer.QueuePermissionRefresh(ctx, plan.user.ID)
	})
}

// randomPasswordHash 给预建账号一个猜不到的口令哈希。
//
// 这些人只走 CAS 登录，口令本身没人会用；留强随机值是避免出现
// 「批量录入的账号共用一个已知口令」这种能被直接拿来本地登录的情况。
func randomPasswordHash() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	hash, err := utils.HashPassword(hex.EncodeToString(buf))
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return hash, nil
}
