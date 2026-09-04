package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *memberService) ExportProfiles(ctx context.Context, viewer uuid.UUID, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]byte, error) {
	if req.PageSize < 1 || req.PageSize > 200 {
		req.PageSize = 200
	}
	if req.Page < 1 {
		req.Page = 1
	}
	rows, _, err := s.ListProfiles(ctx, viewer, req, scope)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, response.NewError(response.CodeMemberExportFail, "没有可导出的档案")
	}
	pdf, err := buildProfilePDF(rows)
	if err != nil {
		return nil, response.NewError(response.CodeMemberExportFail, "导出失败: "+err.Error())
	}
	return pdf, nil
}

func buildProfilePDF(rows []*dto.ProfileResponse) ([]byte, error) {
	var lines []string
	lines = append(lines, "StarByte 人员档案")
	lines = append(lines, "")
	for i, p := range rows {
		if i > 0 {
			lines = append(lines, "--------------------------------")
		}
		dept := ""
		if p.Department != nil {
			dept = p.Department.Name
		}
		pos := ""
		if p.Position != nil {
			pos = p.Position.Name
		}
		lines = append(lines,
			fmt.Sprintf("姓名: %s    学号: %s", p.RealName, p.StudentNo),
			fmt.Sprintf("部门: %s    职位: %s", dept, pos),
			fmt.Sprintf("类型: %s    状态: %s", memberTypeLabel(p.MemberType), profileStatusLabel(p.Status)),
			fmt.Sprintf("年级: %s    专业: %s", p.Grade, p.Major),
			fmt.Sprintf("电话: %s    邮箱: %s", p.ContactPhone, p.ContactEmail),
			fmt.Sprintf("技能: %s", strings.Join(p.Skills, ", ")),
			fmt.Sprintf("简介: %s", p.Bio),
		)
	}
	return renderSimplePDF(lines), nil
}

func memberTypeLabel(t int16) string {
	switch t {
	case 1:
		return "会员"
	case 2:
		return "干事"
	case 3:
		return "部长"
	case 4:
		return "社长"
	default:
		return "未知"
	}
}

func profileStatusLabel(s int16) string {
	switch s {
	case 0:
		return "正常"
	case 1:
		return "禁用"
	case 2:
		return "已退出"
	default:
		return "未知"
	}
}

// renderSimplePDF 生成带 Adobe CJK 标准字体的单页 PDF，避免额外依赖。
func renderSimplePDF(lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT\n/F1 12 Tf\n50 800 Td\n14 TL\n")
	for _, line := range lines {
		content.WriteString("<")
		content.WriteString(utf16Hex(line))
		content.WriteString("> '\n")
	}
	content.WriteString("ET\n")
	stream := content.Bytes()

	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, 8)
	writeObj := func(id int, body string) {
		offsets[id] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", id, body)
	}
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>")
	offsets[4] = b.Len()
	fmt.Fprintf(&b, "4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(stream), stream)
	writeObj(5, "<< /Type /Font /Subtype /Type0 /BaseFont /STSong-Light /Encoding /UniGB-UCS2-H /DescendantFonts [6 0 R] >>")
	writeObj(6, "<< /Type /Font /Subtype /CIDFontType0 /BaseFont /STSong-Light /CIDSystemInfo << /Registry (Adobe) /Ordering (GB1) /Supplement 2 >> /FontDescriptor 7 0 R >>")
	writeObj(7, "<< /Type /FontDescriptor /FontName /STSong-Light /Flags 4 /FontBBox [-200 -200 1200 900] /ItalicAngle 0 /Ascent 800 /Descent -200 /CapHeight 800 /StemV 80 >>")

	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 8\n0000000000 65535 f \n")
	for i := 1; i <= 7; i++ {
		fmt.Fprintf(&b, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 8 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return b.Bytes()
}

func utf16Hex(s string) string {
	codes := utf16.Encode([]rune(s))
	var b strings.Builder
	for _, c := range codes {
		fmt.Fprintf(&b, "%04X", c)
	}
	return b.String()
}
