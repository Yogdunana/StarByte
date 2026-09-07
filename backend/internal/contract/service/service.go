package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/contract/dto"
	"github.com/Yogdunana/StarByte/backend/internal/contract/model"
	"github.com/Yogdunana/StarByte/backend/internal/contract/repo"
	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

type notificationAdapter struct{ inner notifsvc.NotificationService }

func NewNotifier(inner notifsvc.NotificationService) Notifier {
	if inner == nil {
		return nil
	}
	return &notificationAdapter{inner: inner}
}

func (a *notificationAdapter) Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error {
	if len(userIDs) == 0 {
		return nil
	}
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs: userIDs, TemplateCode: template, Variables: vars, Channels: []string{"in_app", "websocket"},
	})
}

type Service interface {
	Create(ctx context.Context, operator uuid.UUID, req *dto.CreateContractRequest) (*dto.ContractResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateContractRequest) (*dto.ContractResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (*dto.ContractResponse, error)
	List(ctx context.Context, req *dto.ListContractRequest) ([]*dto.ContractResponse, int64, int, int, error)
	Templates(ctx context.Context) ([]dto.TemplateResponse, error)
	Expiring(ctx context.Context, days int) ([]*dto.ContractResponse, error)
	ExpiryJob(ctx context.Context, payload string, logf func(string)) error
}

type contractService struct {
	rows   repo.Repository
	notify Notifier
}

func New(rows repo.Repository, notify Notifier) Service {
	return &contractService{rows: rows, notify: notify}
}

func (s *contractService) Create(ctx context.Context, operator uuid.UUID, req *dto.CreateContractRequest) (*dto.ContractResponse, error) {
	if !model.ValidType(req.ContractType) {
		return nil, response.NewError(response.CodeContractInvalidType, "合同类型不合法")
	}
	if err := validatePeriod(req.StartAt, req.ExpiredAt); err != nil {
		return nil, err
	}
	now := time.Now()
	row := &model.Contract{
		ID: uuid.New(), UserID: operator, Title: strings.TrimSpace(req.Title),
		ContractType: req.ContractType, PartyName: strings.TrimSpace(req.PartyName),
		Amount: req.Amount, StartAt: req.StartAt, ExpiredAt: req.ExpiredAt,
		Status: model.StatusDraft, CreatedAt: now, UpdatedAt: now,
	}
	if req.Status != nil {
		row.Status = *req.Status
	}
	if row.Status == model.StatusActive {
		row.SignedAt = &now
	}
	if err := s.bindRefs(ctx, row, req.TemplateID, req.FileID); err != nil {
		return nil, err
	}
	if err := s.rows.Create(ctx, row); err != nil {
		return nil, fmt.Errorf("create contract: %w", err)
	}
	return s.Get(ctx, row.ID)
}

func (s *contractService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateContractRequest) (*dto.ContractResponse, error) {
	row, err := s.must(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status == model.StatusEnded {
		return nil, response.NewError(response.CodeContractInvalidState, "已终止合同不可修改")
	}
	if req.Title != nil {
		row.Title = strings.TrimSpace(*req.Title)
	}
	if req.ContractType != nil {
		if !model.ValidType(*req.ContractType) {
			return nil, response.NewError(response.CodeContractInvalidType, "合同类型不合法")
		}
		row.ContractType = *req.ContractType
	}
	if req.PartyName != nil {
		row.PartyName = strings.TrimSpace(*req.PartyName)
	}
	if req.Amount != nil {
		row.Amount = req.Amount
	}
	if req.StartAt != nil {
		row.StartAt = req.StartAt
	}
	if req.ExpiredAt != nil {
		row.ExpiredAt = req.ExpiredAt
	}
	if err := validatePeriod(row.StartAt, row.ExpiredAt); err != nil {
		return nil, err
	}
	tpl, file := "", ""
	if req.TemplateID != nil {
		tpl = *req.TemplateID
	}
	if req.FileID != nil {
		file = *req.FileID
	}
	if req.TemplateID != nil || req.FileID != nil {
		if err := s.bindRefs(ctx, row, tpl, file); err != nil {
			return nil, err
		}
	}
	if req.Status != nil {
		row.Status = *req.Status
		if row.Status == model.StatusActive && row.SignedAt == nil {
			now := time.Now()
			row.SignedAt = &now
		}
	}
	row.UpdatedAt = time.Now()
	s.touchExpired(row)
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("update contract: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *contractService) Delete(ctx context.Context, id uuid.UUID) error {
	row, err := s.must(ctx, id)
	if err != nil {
		return err
	}
	if row.Status == model.StatusActive {
		return response.NewError(response.CodeContractInvalidState, "生效中的合同请先终止再删除")
	}
	if err := s.rows.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}
	return nil
}

func (s *contractService) Get(ctx context.Context, id uuid.UUID) (*dto.ContractResponse, error) {
	row, err := s.rows.GetNamed(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get contract: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeContractNotFound, "合同不存在")
	}
	s.persistExpired(ctx, &row.Contract)
	return mapContract(row), nil
}

func (s *contractService) List(ctx context.Context, req *dto.ListContractRequest) ([]*dto.ContractResponse, int64, int, int, error) {
	rows, total, err := s.rows.List(ctx, req)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("list contracts: %w", err)
	}
	page, size := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	out := make([]*dto.ContractResponse, 0, len(rows))
	for i := range rows {
		s.persistExpired(ctx, &rows[i].Contract)
		out = append(out, mapContract(&rows[i]))
	}
	return out, total, page, size, nil
}

func (s *contractService) Templates(ctx context.Context) ([]dto.TemplateResponse, error) {
	rows, err := s.rows.ListTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	out := make([]dto.TemplateResponse, 0, len(rows))
	for _, t := range rows {
		out = append(out, dto.TemplateResponse{ID: t.ID.String(), Name: t.Name, Code: t.Code, Content: t.Content})
	}
	return out, nil
}

func (s *contractService) Expiring(ctx context.Context, days int) ([]*dto.ContractResponse, error) {
	if days <= 0 {
		days = 30
	}
	until := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	rows, err := s.rows.ListExpiring(ctx, until)
	if err != nil {
		return nil, fmt.Errorf("list expiring: %w", err)
	}
	out := make([]*dto.ContractResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapContract(&rows[i]))
	}
	return out, nil
}

func (s *contractService) ExpiryJob(ctx context.Context, _ string, logf func(string)) error {
	n, err := s.rows.MarkExpired(ctx, time.Now())
	if err != nil {
		return err
	}
	logf(fmt.Sprintf("marked %d expired contracts", n))
	rows, err := s.rows.ListExpiring(ctx, time.Now().Add(7*24*time.Hour))
	if err != nil {
		return err
	}
	for i := range rows {
		s.notifyOwner(ctx, &rows[i])
	}
	logf(fmt.Sprintf("notified %d expiring contracts", len(rows)))
	return nil
}
