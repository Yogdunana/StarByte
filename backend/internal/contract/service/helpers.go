package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/contract/dto"
	"github.com/Yogdunana/StarByte/backend/internal/contract/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *contractService) bindRefs(ctx context.Context, row *model.Contract, templateID, fileID string) error {
	if templateID != "" {
		tid, err := uuid.Parse(templateID)
		if err != nil {
			return response.NewError(response.CodeBadRequest, "模板 ID 无效")
		}
		tpl, err := s.rows.GetTemplate(ctx, tid)
		if err != nil {
			return fmt.Errorf("get template: %w", err)
		}
		if tpl == nil {
			return response.NewError(response.CodeContractTemplateGone, "合同模板不存在")
		}
		row.TemplateID = &tid
	}
	if fileID != "" {
		fid, err := uuid.Parse(fileID)
		if err != nil {
			return response.NewError(response.CodeBadRequest, "附件 ID 无效")
		}
		row.FileID = &fid
	}
	return nil
}

func (s *contractService) must(ctx context.Context, id uuid.UUID) (*model.Contract, error) {
	row, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get contract: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeContractNotFound, "合同不存在")
	}
	return row, nil
}

func (s *contractService) touchExpired(row *model.Contract) {
	if row.Status == model.StatusActive && row.ExpiredAt != nil && row.ExpiredAt.Before(time.Now()) {
		row.Status = model.StatusExpired
	}
}

func (s *contractService) persistExpired(ctx context.Context, row *model.Contract) {
	if row.Status != model.StatusActive || row.ExpiredAt == nil || !row.ExpiredAt.Before(time.Now()) {
		return
	}
	row.Status = model.StatusExpired
	row.UpdatedAt = time.Now()
	if err := s.rows.Update(ctx, row); err != nil {
		logger.Warn("mark contract expired failed", zap.Error(err))
	}
}

func (s *contractService) notifyOwner(ctx context.Context, row *model.ContractNamed) {
	if s.notify == nil {
		return
	}
	exp := ""
	if row.ExpiredAt != nil {
		exp = row.ExpiredAt.Format("2006-01-02")
	}
	_ = s.notify.Send(ctx, []uuid.UUID{row.UserID}, "contract_expiring", map[string]interface{}{
		"real_name": row.OwnerName, "title": row.Title, "expired_at": exp,
	})
}

func validatePeriod(start, end *time.Time) error {
	if start != nil && end != nil && end.Before(*start) {
		return response.NewError(response.CodeContractInvalidPeriod, "结束日期不能早于开始日期")
	}
	return nil
}

func mapContract(row *model.ContractNamed) *dto.ContractResponse {
	out := &dto.ContractResponse{
		ID: row.ID.String(), Title: row.Title, Status: row.Status, ContractType: row.ContractType,
		PartyName: row.PartyName, Amount: row.Amount, TemplateName: row.TemplateName, FileName: row.FileName,
		Owner:   dto.Person{ID: row.UserID.String(), Name: row.OwnerName},
		StartAt: row.StartAt, SignedAt: row.SignedAt, ExpiredAt: row.ExpiredAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.TemplateID != nil {
		out.TemplateID = row.TemplateID.String()
	}
	if row.FileID != nil {
		out.FileID = row.FileID.String()
	}
	return out
}
