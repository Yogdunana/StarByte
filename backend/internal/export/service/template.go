package service

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

//go:embed templates/*.html
var templateFS embed.FS

type builtinTpl struct {
	ID          string
	Name        string
	Description string
	File        string
}

func builtins() []builtinTpl {
	return []builtinTpl{
		{ID: "member_application", Name: "入会申请表", File: "templates/member_application.html", Description: "入会申请打印件"},
		{ID: "meeting_notice", Name: "会议通知", File: "templates/meeting_notice.html", Description: "会议召开通知"},
		{ID: "internship_record", Name: "实习记录", File: "templates/internship_record.html", Description: "IT 实习过程记录"},
	}
}

func listBuiltinTemplates() []dto.TemplateInfo {
	list := builtins()
	out := make([]dto.TemplateInfo, 0, len(list))
	for _, t := range list {
		out = append(out, dto.TemplateInfo{ID: t.ID, Name: t.Name, Description: t.Description})
	}
	return out
}

func findBuiltin(id string) *builtinTpl {
	id = strings.TrimSpace(id)
	for _, t := range builtins() {
		if t.ID == id {
			item := t
			return &item
		}
	}
	return nil
}

func renderTemplate(id string, vars map[string]string) (string, error) {
	meta := findBuiltin(id)
	if meta == nil {
		return "", response.NewError(response.CodeExportTplNotFound, "导出模板不存在")
	}
	raw, err := templateFS.ReadFile(meta.File)
	if err != nil {
		return "", response.NewError(response.CodeExportTplNotFound, "导出模板不存在")
	}
	tpl, err := template.New(meta.ID).Parse(string(raw))
	if err != nil {
		return "", response.NewError(response.CodeInternalError, "模板解析失败")
	}
	if vars == nil {
		vars = map[string]string{}
	}
	if _, ok := vars["Title"]; !ok {
		vars["Title"] = meta.Name
	}
	if _, ok := vars["Date"]; !ok {
		vars["Date"] = time.Now().Format("2006-01-02")
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, vars); err != nil {
		return "", response.NewError(response.CodeInternalError, "模板渲染失败")
	}
	return buf.String(), nil
}
