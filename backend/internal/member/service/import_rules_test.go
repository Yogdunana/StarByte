package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

func TestNormalizeImportRole(t *testing.T) {
	for raw, want := range map[string]string{
		"":                     DefaultImportRole,
		"  minister ":          "minister",
		"MINISTER":             "minister",
		"部长":                   "minister",
		"中心副主任":                "vice_center_director",
		"Vice_Center_Director": "vice_center_director",
		"预备干事":                 "probationary",
		"正式干事":                 "officer",
		"会员":                   "member",
	} {
		got, ok := normalizeImportRole(raw)
		require.True(t, ok, raw)
		require.Equal(t, want, got, raw)
	}
	for _, raw := range []string{"president", "center_director", "副部長", "student"} {
		_, ok := normalizeImportRole(raw)
		require.False(t, ok, raw)
	}
}

func TestValidateImportRowTrimsAndRejectsBadCells(t *testing.T) {
	gender := int16(1)
	row := dto.MemberImportRow{StudentNo: " 20240001 ", RealName: " 张\u00a0三 ", Gender: &gender, Email: " a@b.com "}
	require.Empty(t, validateImportRow(&row))
	require.Equal(t, "20240001", row.StudentNo)
	// 单元格内的 NBSP 换成普通空格，不再是不可见字符；首尾空白照常去掉。
	require.NotContains(t, row.RealName, "\u00a0")
	require.Equal(t, "张 三", row.RealName)
	require.Equal(t, "a@b.com", row.Email)
}

func TestValidateImportRowRejectsMissingStudentNo(t *testing.T) {
	row := dto.MemberImportRow{RealName: "李四"}
	require.Equal(t, "学号不能为空", validateImportRow(&row))
}

func TestValidateImportRowRejectsBadEmail(t *testing.T) {
	row := dto.MemberImportRow{StudentNo: "20240002", RealName: "李四", Email: "not-an-email"}
	require.Equal(t, "邮箱格式不正确", validateImportRow(&row))
}

func TestValidateImportRowRejectsBadGender(t *testing.T) {
	gender := int16(9)
	row := dto.MemberImportRow{StudentNo: "20240003", RealName: "王五", Gender: &gender}
	require.Equal(t, "性别只能填 0/1/2（未知/男/女）", validateImportRow(&row))
}

func TestValidateImportRowRejectsUnknownRole(t *testing.T) {
	row := dto.MemberImportRow{StudentNo: "20240004", RealName: "赵六", Role: "会长"}
	require.Contains(t, validateImportRow(&row), "不支持的角色")
}

func TestImportProfileStatusOnlyProbationStaysInProbation(t *testing.T) {
	require.Equal(t, model.ProfileProbation, importProfileStatus("probationary"))
	for _, role := range []string{"member", "officer", "minister", "vice_center_director"} {
		require.Equal(t, model.ProfileActive, importProfileStatus(role), role)
	}
}

func TestImportOfficeRolesRequireSuperAdmin(t *testing.T) {
	for code, rule := range importRoles {
		require.Equal(t, code, trimCell(code))
		if rule.Office == officeScopeNone {
			require.False(t, rule.NeedScope, code)
		} else {
			require.True(t, rule.NeedScope, code)
		}
	}
	require.Contains(t, ImportRoles(), "minister")
	require.NotContains(t, ImportRoles(), "president")
}

func TestPureImportChecksRejectsDuplicateStudentNumbers(t *testing.T) {
	seen := map[string]int{}
	first := dto.MemberImportRow{StudentNo: "20240001", RealName: "张三"}
	_, msg := pureImportChecks(&first, seen, true)
	require.Empty(t, msg)
	seen[first.StudentNo] = 1

	second := dto.MemberImportRow{StudentNo: "20240001", RealName: "李四"}
	_, msg = pureImportChecks(&second, seen, true)
	require.Contains(t, msg, "已经出现过")
}

func TestPureImportChecksGuardsCharterOffices(t *testing.T) {
	minister := dto.MemberImportRow{StudentNo: "20240010", RealName: "王五", Role: "minister"}
	_, msg := pureImportChecks(&minister, map[string]int{}, false)
	require.Equal(t, "协会职务须由系统管理员登记任命", msg)

	_, msg = pureImportChecks(&minister, map[string]int{}, true)
	require.Equal(t, "该角色必须填写部门或中心", msg)

	scoped := minister
	scoped.Department = "技术部"
	rule, msg := pureImportChecks(&scoped, map[string]int{}, true)
	require.Empty(t, msg)
	require.Equal(t, officeScopeDepartment, rule.Office)
}

func TestPureImportChecksAllowsProbationaryWithoutDepartment(t *testing.T) {
	row := dto.MemberImportRow{StudentNo: "20240011", RealName: "赵六"}
	rule, msg := pureImportChecks(&row, map[string]int{}, false)
	require.Empty(t, msg)
	require.Equal(t, model.MemberTypeOfficer, rule.MemberType)
}

func TestRandomPasswordHashIsDifferentEveryTime(t *testing.T) {
	first, err := randomPasswordHash()
	require.NoError(t, err)
	second, err := randomPasswordHash()
	require.NoError(t, err)
	require.NotEmpty(t, first)
	require.NotEqual(t, first, second)
}

func TestPreviewRejectsOversizedBatch(t *testing.T) {
	ctx := context.Background()
	rows := make([]dto.MemberImportRow, MaxImportRows+1)
	svc := NewMemberImportService(nil, nil)
	_, err := svc.Preview(ctx, true, &dto.MemberImportRequest{Rows: rows})
	require.Error(t, err)
	require.Contains(t, err.Error(), "分批导入")
}

func TestPreviewOfEmptyBatchIsEmpty(t *testing.T) {
	ctx := context.Background()
	svc := NewMemberImportService(nil, nil)
	result, err := svc.Preview(ctx, true, &dto.MemberImportRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, result.Total)
}
