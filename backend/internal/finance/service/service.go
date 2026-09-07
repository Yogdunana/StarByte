package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/finance/dto"
	"github.com/Yogdunana/StarByte/backend/internal/finance/model"
	"github.com/Yogdunana/StarByte/backend/internal/finance/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, operator uuid.UUID, req *dto.CreateRecordRequest) (*dto.RecordResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRecordRequest) (*dto.RecordResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (*dto.RecordResponse, error)
	List(ctx context.Context, req *dto.ListRecordRequest) ([]*dto.RecordResponse, int64, int, int, error)
	Categories(ctx context.Context) ([]dto.CategoryResponse, error)
	Summary(ctx context.Context, q dto.SummaryQuery) (*dto.SummaryResponse, error)
}

type financeService struct{ rows repo.Repository }

func New(rows repo.Repository) Service { return &financeService{rows: rows} }

func (s *financeService) Create(ctx context.Context, operator uuid.UUID, req *dto.CreateRecordRequest) (*dto.RecordResponse, error) {
	if req.Amount <= 0 {
		return nil, response.NewError(response.CodeFinanceInvalidAmount, "金额必须大于 0")
	}
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "分类 ID 无效")
	}
	cat, err := s.rows.GetCategory(ctx, catID)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	if cat == nil {
		return nil, response.NewError(response.CodeFinanceCategoryGone, "收支分类不存在")
	}
	now := time.Now()
	row := &model.Record{
		ID: uuid.New(), CategoryID: catID, Amount: req.Amount, Direction: req.Direction,
		OccurredAt: dateOnly(req.OccurredAt), Title: strings.TrimSpace(req.Title),
		Remark: req.Remark, CreatedBy: &operator, CreatedAt: now,
	}
	if dept := parseUUID(req.DepartmentID); dept != nil {
		row.DepartmentID = dept
	}
	if err := s.rows.Create(ctx, row); err != nil {
		return nil, fmt.Errorf("create finance record: %w", err)
	}
	return s.Get(ctx, row.ID)
}

func (s *financeService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRecordRequest) (*dto.RecordResponse, error) {
	row, err := s.must(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, response.NewError(response.CodeFinanceInvalidAmount, "金额必须大于 0")
		}
		row.Amount = *req.Amount
	}
	if req.CategoryID != nil {
		catID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "分类 ID 无效")
		}
		cat, err := s.rows.GetCategory(ctx, catID)
		if err != nil {
			return nil, fmt.Errorf("get category: %w", err)
		}
		if cat == nil {
			return nil, response.NewError(response.CodeFinanceCategoryGone, "收支分类不存在")
		}
		row.CategoryID = catID
	}
	if req.Direction != nil {
		row.Direction = *req.Direction
	}
	if req.OccurredAt != nil {
		row.OccurredAt = dateOnly(*req.OccurredAt)
	}
	if req.Title != nil {
		row.Title = strings.TrimSpace(*req.Title)
	}
	if req.Remark != nil {
		row.Remark = *req.Remark
	}
	if req.DepartmentID != nil {
		row.DepartmentID = parseUUID(*req.DepartmentID)
	}
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("update finance record: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *financeService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.must(ctx, id); err != nil {
		return err
	}
	if err := s.rows.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete finance record: %w", err)
	}
	return nil
}

func (s *financeService) Get(ctx context.Context, id uuid.UUID) (*dto.RecordResponse, error) {
	row, err := s.rows.GetNamed(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get finance record: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeFinanceNotFound, "财务记录不存在")
	}
	return mapRecord(row), nil
}

func (s *financeService) List(ctx context.Context, req *dto.ListRecordRequest) ([]*dto.RecordResponse, int64, int, int, error) {
	rows, total, err := s.rows.List(ctx, req)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("list finance records: %w", err)
	}
	page, size := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	out := make([]*dto.RecordResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapRecord(&rows[i]))
	}
	return out, total, page, size, nil
}

func (s *financeService) Categories(ctx context.Context) ([]dto.CategoryResponse, error) {
	rows, err := s.rows.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list finance categories: %w", err)
	}
	out := make([]dto.CategoryResponse, 0, len(rows))
	for _, c := range rows {
		out = append(out, dto.CategoryResponse{
			ID: c.ID.String(), Name: c.Name, Code: c.Code, Direction: c.Direction, Description: c.Description,
		})
	}
	return out, nil
}

func (s *financeService) Summary(ctx context.Context, q dto.SummaryQuery) (*dto.SummaryResponse, error) {
	totals, cats, err := s.rows.Summary(ctx, q.From, q.To, q.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("finance summary: %w", err)
	}
	out := &dto.SummaryResponse{ByCategory: []dto.CategorySum{}}
	for _, row := range totals {
		if row.Direction == model.DirectionIncome {
			out.IncomeTotal = row.Total
			out.IncomeCount = row.Count
		} else {
			out.ExpenseTotal = row.Total
			out.ExpenseCount = row.Count
		}
	}
	out.Balance = out.IncomeTotal - out.ExpenseTotal
	for _, c := range cats {
		out.ByCategory = append(out.ByCategory, dto.CategorySum{
			CategoryID: c.CategoryID.String(), CategoryName: c.CategoryName,
			Direction: c.Direction, Total: c.Total, Count: c.Count,
		})
	}
	return out, nil
}

func (s *financeService) must(ctx context.Context, id uuid.UUID) (*model.Record, error) {
	row, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get finance record: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeFinanceNotFound, "财务记录不存在")
	}
	return row, nil
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func parseUUID(raw string) *uuid.UUID {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil
	}
	return &id
}

func mapRecord(row *model.RecordNamed) *dto.RecordResponse {
	out := &dto.RecordResponse{
		ID: row.ID.String(), CategoryID: row.CategoryID.String(), CategoryName: row.CategoryName,
		Amount: row.Amount, Direction: row.Direction, OccurredAt: row.OccurredAt,
		Title: row.Title, Remark: row.Remark, CreatedAt: row.CreatedAt,
	}
	if row.DepartmentID != nil {
		out.DepartmentID = row.DepartmentID.String()
		out.DepartmentName = row.DepartmentName
	}
	if row.CreatedBy != nil {
		out.CreatedBy = &dto.Person{ID: row.CreatedBy.String(), Name: row.CreatorName}
	}
	return out
}
