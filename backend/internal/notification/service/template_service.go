package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// TemplateService 通知模板管理服务接口
type TemplateService interface {
	Create(ctx context.Context, req *dto.CreateTemplateRequest) (*dto.NotificationTemplateResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.NotificationTemplateResponse, error)
	List(ctx context.Context, req *dto.ListTemplatesRequest) ([]*dto.NotificationTemplateResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateTemplateRequest) (*dto.NotificationTemplateResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Test(ctx context.Context, id uuid.UUID, req *dto.TestTemplateRequest) (*dto.TestTemplateResponse, error)
}

type templateService struct {
	templateRepo   repo.NotificationTemplateRepo
	templateEngine TemplateEngine
}

// NewTemplateService 创建模板服务
func NewTemplateService(
	templateRepo repo.NotificationTemplateRepo,
	templateEngine TemplateEngine,
) TemplateService {
	return &templateService{
		templateRepo:   templateRepo,
		templateEngine: templateEngine,
	}
}

func (s *templateService) Create(ctx context.Context, req *dto.CreateTemplateRequest) (*dto.NotificationTemplateResponse, error) {
	// 检查 code 是否已存在
	existing, err := s.templateRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, response.NewError(response.CodeNotificationTplExists, "模板 Code 已存在")
	}

	tpl := &model.NotificationTemplate{
		ID:            uuid.New(),
		Code:          req.Code,
		Name:          req.Name,
		TitleTemplate: req.TitleTemplate,
		BodyTemplate:  req.BodyTemplate,
		Category:      req.Category,
		Status:        0,
	}
	tpl.SetChannels(req.Channels)
	tpl.SetVariablesSchema(req.VariablesSchema)

	if err := s.templateRepo.Create(ctx, tpl); err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}

	return TemplateToResponse(tpl), nil
}

func (s *templateService) GetByID(ctx context.Context, id uuid.UUID) (*dto.NotificationTemplateResponse, error) {
	tpl, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, response.NewError(response.CodeNotificationTplNotFound, "通知模板不存在")
	}
	return TemplateToResponse(tpl), nil
}

func (s *templateService) List(ctx context.Context, req *dto.ListTemplatesRequest) ([]*dto.NotificationTemplateResponse, int64, error) {
	list, total, err := s.templateRepo.List(ctx, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("list templates: %w", err)
	}

	result := make([]*dto.NotificationTemplateResponse, 0, len(list))
	for _, tpl := range list {
		result = append(result, TemplateToResponse(tpl))
	}
	return result, total, nil
}

func (s *templateService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateTemplateRequest) (*dto.NotificationTemplateResponse, error) {
	tpl, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, response.NewError(response.CodeNotificationTplNotFound, "通知模板不存在")
	}

	if req.Name != nil {
		tpl.Name = *req.Name
	}
	if req.TitleTemplate != nil {
		tpl.TitleTemplate = *req.TitleTemplate
	}
	if req.BodyTemplate != nil {
		tpl.BodyTemplate = *req.BodyTemplate
	}
	if req.Category != nil {
		tpl.Category = *req.Category
	}
	if req.Channels != nil {
		tpl.SetChannels(req.Channels)
	}
	if req.VariablesSchema != nil {
		tpl.SetVariablesSchema(req.VariablesSchema)
	}
	if req.Status != nil {
		tpl.Status = *req.Status
	}

	if err := s.templateRepo.Update(ctx, tpl); err != nil {
		return nil, fmt.Errorf("update template: %w", err)
	}

	return TemplateToResponse(tpl), nil
}

func (s *templateService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.templateRepo.Delete(ctx, id); err != nil {
		return response.NewError(response.CodeNotificationTplNotFound, "通知模板不存在")
	}
	return nil
}

func (s *templateService) Test(ctx context.Context, id uuid.UUID, req *dto.TestTemplateRequest) (*dto.TestTemplateResponse, error) {
	tpl, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, response.NewError(response.CodeNotificationTplNotFound, "通知模板不存在")
	}
	return s.templateEngine.Render(ctx, tpl.Code, req.Variables)
}

// channelsToJSON 将渠道列表序列化为 JSON 字符串
func channelsToJSON(channels []string) string {
	data, _ := json.Marshal(channels)
	return string(data)
}

// schemaToJSON 将变量 schema 序列化为 JSON 字符串
func schemaToJSON(schema map[string]string) string {
	data, _ := json.Marshal(schema)
	return string(data)
}
