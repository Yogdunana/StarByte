package service

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

// MaxImportRows 一次最多录多少人。测评系统一届的量级远小于此，
// 这个上限只是防止把整张表塞进一个请求里。
const MaxImportRows = 500

// DefaultImportRole 是没有写明角色时用的角色。批量录入的场景是「测评完之后集体进协会」，
// 这些人还没转正，所以默认按预备干事处理，满一个月后由部长在现有流程里确认转正。
const DefaultImportRole = "probationary"

// 任职范围的三种形态，取值与 rbac/handler.associationOffice 保持一致，
// 免得两条写角色的路径出现「手动任命有范围、批量录入没范围」的分叉。
const (
	officeScopeNone       = ""
	officeScopeUnscoped   = "unscoped"
	officeScopeCenter     = "center"
	officeScopeDepartment = "department"
)

// importRoleRule 描述一个允许批量录入的角色。
type importRoleRule struct {
	MemberType int16  // 落到 member_profiles.member_type
	Office     string // 是否协会职务、以及它绑哪种范围
	NeedScope  bool   // 行里必须给出部门/中心
}

// importRoles 是批量录入的角色白名单。
//
// 只放了「不需要额外背景信息就能登记」的角色：会长、指导老师、中心主任这类要么唯一、
// 要么由现有任命流程产生，不在这里开放；副部长因为 rbac 侧还没归类成协会职务，
// 一旦在这里写范围就会和手动任命产生分叉，也先不放。
// 要扩容时改这张表并同步 docs 即可。
var importRoles = map[string]importRoleRule{
	"member": {
		MemberType: model.MemberTypeMember,
		Office:     officeScopeNone,
	},
	"probationary": {
		MemberType: model.MemberTypeOfficer,
		Office:     officeScopeNone,
	},
	"officer": {
		MemberType: model.MemberTypeOfficer,
		Office:     officeScopeNone,
	},
	"vice_center_director": {
		MemberType: model.MemberTypeMinister,
		Office:     officeScopeCenter,
		NeedScope:  true,
	},
	"minister": {
		MemberType: model.MemberTypeMinister,
		Office:     officeScopeDepartment,
		NeedScope:  true,
	},
}

// ImportRoles 返回允许批量录入的角色编码，供前端下拉直接使用。
func ImportRoles() []string {
	codes := make([]string, 0, len(importRoles))
	for code := range importRoles {
		codes = append(codes, code)
	}
	return codes
}

// normalizeImportRole 把「亚麻不及待」「写错大小写」「没写」都收敛成角色编码。
// 第二个返回值表示能不能认出来。
func normalizeImportRole(role string) (string, bool) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return DefaultImportRole, true
	}
	// 录入表里写中文也能接受，省得每次都要在一堆英文编码里找。
	switch role {
	case "会员", "普通会员":
		role = "member"
	case "预备干事":
		role = "probationary"
	case "正式干事", "干事":
		role = "officer"
	case "中心副主任", "副中心主任":
		role = "vice_center_director"
	case "部长", "部门负责人":
		role = "minister"
	}
	rule, ok := importRoles[role]
	if !ok {
		return "", false
	}
	_ = rule
	return role, true
}

// importRoleRuleOf 取角色规则；不在白名单里返回 nil。
func importRoleRuleOf(role string) (importRoleRule, bool) {
	rule, ok := importRoles[role]
	return rule, ok
}

// importProfileStatus 决定档案状态：只有预备干事处在预备期，
// 其余进来就是正式状态，这样抵押权率等统计口径不用为批量录入单独开分支。
func importProfileStatus(role string) int16 {
	if role == "probationary" {
		return model.ProfileProbation
	}
	return model.ProfileActive
}

// validateImportRow 只做不依赖数据库的校验。返回空串表示通过。
func validateImportRow(row *dto.MemberImportRow) string {
	row.StudentNo = trimCell(row.StudentNo)
	row.RealName = trimCell(row.RealName)
	row.Department = trimCell(row.Department)
	row.Grade = trimCell(row.Grade)
	row.Major = trimCell(row.Major)
	row.Phone = trimCell(row.Phone)
	row.Email = trimCell(row.Email)

	if row.StudentNo == "" {
		return "学号不能为空"
	}
	if len(row.StudentNo) > 30 {
		return "学号过长（超过 30 个字符）"
	}
	if row.RealName == "" {
		return "姓名不能为空"
	}
	if len(row.RealName) > 50 {
		return "姓名过长（超过 50 个字符）"
	}
	if row.Email != "" {
		if _, err := mail.ParseAddress(row.Email); err != nil {
			return "邮箱格式不正确"
		}
		if len(row.Email) > 100 {
			return "邮箱过长（超过 100 个字符）"
		}
	}
	if len(row.Phone) > 20 {
		return "手机号过长（超过 20 个字符）"
	}
	if len(row.Grade) > 20 {
		return "年级过长（超过 20 个字符）"
	}
	if len(row.Major) > 100 {
		return "专业过长（超过 100 个字符）"
	}
	if row.Gender != nil && *row.Gender != 0 && *row.Gender != 1 && *row.Gender != 2 {
		return "性别只能填 0/1/2（未知/男/女）"
	}
	role, ok := normalizeImportRole(row.Role)
	if !ok {
		return "不支持的角色：" + row.Role
	}
	row.Role = role
	return ""
}

// pureImportChecks 是不需要查库的那部分判定：格式、重复、角色是否开放、
// 有没有任命协会职务的权限、协会职务有没有给范围。返回规则与错误文案。
//
// seen 累计「学号 → 首次出现的行号」，用来在一批里自己撞自己的时候给出准确提示。
func pureImportChecks(row *dto.MemberImportRow, seen map[string]int, canAppoint bool) (importRoleRule, string) {
	if msg := validateImportRow(row); msg != "" {
		return importRoleRule{}, msg
	}
	if first, dup := seen[row.StudentNo]; dup {
		return importRoleRule{}, fmt.Sprintf("学号在第 %d 行已经出现过一次", first)
	}
	rule, ok := importRoleRuleOf(row.Role)
	if !ok {
		return importRoleRule{}, "不支持的角色：" + row.Role
	}
	// 与手动任命保持一致：协会职务只能由系统管理员登记。
	if rule.Office != officeScopeNone && !canAppoint {
		return importRoleRule{}, "协会职务须由系统管理员登记任命"
	}
	if rule.NeedScope && row.Department == "" {
		return importRoleRule{}, "该角色必须填写部门或中心"
	}
	return rule, ""
}

// trimCell 去掉 CSV/Excel 常见的多余空白与 Excel 粘来的 NBSP。
func trimCell(value string) string {
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return strings.TrimSpace(value)
}
