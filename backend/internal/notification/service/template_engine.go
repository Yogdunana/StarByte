package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// TemplateEngine 通知模板引擎接口
type TemplateEngine interface {
	Render(ctx context.Context, templateCode string, variables map[string]interface{}) (*dto.TestTemplateResponse, error)
	Validate(ctx context.Context, templateCode string, variables map[string]interface{}) error
}

type templateEngine struct {
	templateRepo repo.NotificationTemplateRepo
}

// NewTemplateEngine 创建模板引擎
func NewTemplateEngine(templateRepo repo.NotificationTemplateRepo) TemplateEngine {
	return &templateEngine{templateRepo: templateRepo}
}

// Render 渲染模板，返回标题和内容
func (e *templateEngine) Render(ctx context.Context, templateCode string, variables map[string]interface{}) (*dto.TestTemplateResponse, error) {
	tpl, err := e.templateRepo.GetByCode(ctx, templateCode)
	if err != nil {
		return nil, response.NewError(response.CodeNotificationTplNotFound, "通知模板不存在")
	}

	if tpl.Status != 0 {
		return nil, response.NewError(response.CodeNotificationTplNotFound, "通知模板已禁用")
	}

	// 渲染标题
	titleBuf, err := renderTemplate(tpl.TitleTemplate, variables)
	if err != nil {
		return nil, response.NewError(response.CodeNotificationRenderFail,
			fmt.Sprintf("模板渲染失败（标题）: %v", err))
	}

	// 渲染内容
	bodyBuf, err := renderTemplate(tpl.BodyTemplate, variables)
	if err != nil {
		return nil, response.NewError(response.CodeNotificationRenderFail,
			fmt.Sprintf("模板渲染失败（内容）: %v", err))
	}

	return &dto.TestTemplateResponse{
		Title:   titleBuf.String(),
		Content: bodyBuf.String(),
	}, nil
}

// Validate 校验模板变量是否完整
func (e *templateEngine) Validate(ctx context.Context, templateCode string, variables map[string]interface{}) error {
	_, err := e.Render(ctx, templateCode, variables)
	return err
}

// renderTemplate 使用 text/template 渲染模板字符串
func renderTemplate(tplText string, variables map[string]interface{}) (*bytes.Buffer, error) {
	t, err := template.New("notification").Parse(tplText)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, variables); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return &buf, nil
}

// TemplateToResponse 将模板模型转为响应 DTO
func TemplateToResponse(t *model.NotificationTemplate) *dto.NotificationTemplateResponse {
	var channels []string
	_ = json.Unmarshal([]byte(t.Channels), &channels)

	var schema map[string]string
	_ = json.Unmarshal([]byte(t.VariablesSchema), &schema)

	return &dto.NotificationTemplateResponse{
		ID:              t.ID.String(),
		Code:            t.Code,
		Name:            t.Name,
		TitleTemplate:   t.TitleTemplate,
		BodyTemplate:    t.BodyTemplate,
		Channels:        channels,
		Category:        t.Category,
		VariablesSchema: schema,
		Status:          t.Status,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
