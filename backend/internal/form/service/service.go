package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/form/dto"
	"github.com/Yogdunana/StarByte/backend/internal/form/model"
	"github.com/Yogdunana/StarByte/backend/internal/form/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormService interface {
	List(ctx context.Context, q dto.ListQuery, canReadAll bool) ([]dto.FormListItem, int64, int, int, error)
	Get(ctx context.Context, id uuid.UUID, canReadAll bool) (*dto.FormDetail, error)
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateFormRequest) (*dto.FormDetail, error)
	Update(ctx context.Context, userID, id uuid.UUID, req dto.UpdateFormRequest) (*dto.FormDetail, error)
	Submit(ctx context.Context, userID, id uuid.UUID, values map[string]interface{}) (*dto.SubmitResponse, error)
	ListSubmissions(ctx context.Context, id uuid.UUID, q dto.SubmissionQuery) ([]dto.SubmissionItem, int64, int, int, error)
}

type formService struct{ repo repo.Repository }

func New(r repo.Repository) FormService { return &formService{repo: r} }

func pageOf(page, size int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	return page, size, (page - 1) * size
}

func (s *formService) require(ctx context.Context, id uuid.UUID) (*model.Form, error) {
	rec, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewError(response.CodeFormNotFound, "表单不存在")
		}
		return nil, err
	}
	return rec, nil
}

func (s *formService) toDetail(ctx context.Context, rec *model.Form) (*dto.FormDetail, error) {
	n, err := s.repo.CountSubmissions(ctx, rec.ID)
	if err != nil {
		return nil, err
	}
	fields := []model.FormField(rec.Fields)
	if fields == nil {
		fields = []model.FormField{}
	}
	return &dto.FormDetail{
		ID: rec.ID.String(), Name: rec.Name, Description: rec.Description, Status: rec.Status,
		Fields: fields, SubmissionCount: n, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}, nil
}

func (s *formService) List(ctx context.Context, q dto.ListQuery, canReadAll bool) ([]dto.FormListItem, int64, int, int, error) {
	page, size, offset := pageOf(q.Page, q.PageSize)
	status := q.Status
	if !canReadAll {
		pub := model.StatusPublished
		status = &pub
	}
	list, total, err := s.repo.List(ctx, strings.TrimSpace(q.Keyword), status, offset, size)
	if err != nil {
		return nil, 0, page, size, err
	}
	out := make([]dto.FormListItem, 0, len(list))
	for i := range list {
		n, err := s.repo.CountSubmissions(ctx, list[i].ID)
		if err != nil {
			return nil, 0, page, size, err
		}
		out = append(out, dto.FormListItem{
			ID: list[i].ID.String(), Name: list[i].Name, Description: list[i].Description,
			Status: list[i].Status, SubmissionCount: n, CreatedAt: list[i].CreatedAt, UpdatedAt: list[i].UpdatedAt,
		})
	}
	return out, total, page, size, nil
}

func (s *formService) Get(ctx context.Context, id uuid.UUID, canReadAll bool) (*dto.FormDetail, error) {
	rec, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canReadAll && rec.Status != model.StatusPublished {
		return nil, response.NewError(response.CodeFormNotPublished, "表单未发布")
	}
	return s.toDetail(ctx, rec)
}

func validStatus(st int16) bool {
	return st == model.StatusDraft || st == model.StatusPublished || st == model.StatusDisabled
}

func (s *formService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateFormRequest) (*dto.FormDetail, error) {
	name := strings.TrimSpace(req.Name)
	if err := validateSchema(req.Fields); err != nil {
		return nil, err
	}
	if existing, err := s.repo.GetByName(ctx, name); err == nil && existing != nil {
		return nil, response.NewError(response.CodeFormNameExists, "表单名称已存在")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	st := model.StatusDraft
	if req.Status != nil {
		if !validStatus(*req.Status) {
			return nil, response.NewError(response.CodeFormInvalidStatus, "表单状态不合法")
		}
		st = *req.Status
	}
	uid := userID
	rec := &model.Form{
		ID: uuid.New(), Name: name, Description: strings.TrimSpace(req.Description),
		Status: st, Fields: model.JSONFields(req.Fields), CreatedBy: &uid, UpdatedBy: &uid,
	}
	if rec.Fields == nil {
		rec.Fields = model.JSONFields{}
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, err
	}
	return s.toDetail(ctx, rec)
}

func (s *formService) Update(ctx context.Context, userID, id uuid.UUID, req dto.UpdateFormRequest) (*dto.FormDetail, error) {
	rec, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if existing, err := s.repo.GetByName(ctx, name); err == nil && existing != nil && existing.ID != rec.ID {
			return nil, response.NewError(response.CodeFormNameExists, "表单名称已存在")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		rec.Name = name
	}
	if req.Description != nil {
		rec.Description = strings.TrimSpace(*req.Description)
	}
	if req.Fields != nil {
		if err := validateSchema(*req.Fields); err != nil {
			return nil, err
		}
		rec.Fields = model.JSONFields(*req.Fields)
	}
	if req.Status != nil {
		if !validStatus(*req.Status) {
			return nil, response.NewError(response.CodeFormInvalidStatus, "表单状态不合法")
		}
		rec.Status = *req.Status
	}
	uid := userID
	rec.UpdatedBy = &uid
	if err := s.repo.Update(ctx, rec); err != nil {
		return nil, err
	}
	return s.toDetail(ctx, rec)
}

func (s *formService) Submit(ctx context.Context, userID, id uuid.UUID, values map[string]interface{}) (*dto.SubmitResponse, error) {
	rec, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.Status != model.StatusPublished {
		return nil, response.NewError(response.CodeFormNotPublished, "表单未发布")
	}
	if values == nil {
		values = map[string]interface{}{}
	}
	cleaned := model.JSONMap{}
	for _, f := range rec.Fields {
		if !isVisible(f, values) {
			continue
		}
		v := values[f.Name]
		if err := validateValue(f, v); err != nil {
			return nil, err
		}
		if !isEmpty(v) {
			cleaned[f.Name] = v
		}
	}
	uid := userID
	now := time.Now()
	sub := &model.Submission{ID: uuid.New(), FormID: rec.ID, Data: cleaned, SubmittedBy: &uid, SubmittedAt: now}
	if err := s.repo.CreateSubmission(ctx, sub); err != nil {
		return nil, err
	}
	return &dto.SubmitResponse{SubmissionID: sub.ID.String(), SubmittedAt: now}, nil
}

func (s *formService) ListSubmissions(ctx context.Context, id uuid.UUID, q dto.SubmissionQuery) ([]dto.SubmissionItem, int64, int, int, error) {
	if _, err := s.require(ctx, id); err != nil {
		return nil, 0, 0, 0, err
	}
	page, size, offset := pageOf(q.Page, q.PageSize)
	list, total, err := s.repo.ListSubmissions(ctx, id, offset, size)
	if err != nil {
		return nil, 0, page, size, err
	}
	out := make([]dto.SubmissionItem, 0, len(list))
	for i := range list {
		data := map[string]interface{}(list[i].Data)
		if data == nil {
			data = map[string]interface{}{}
		}
		item := dto.SubmissionItem{ID: list[i].ID.String(), Data: data, SubmittedAt: list[i].SubmittedAt}
		if list[i].SubmittedBy != nil {
			item.SubmittedBy = &dto.UserBrief{ID: list[i].SubmittedBy.String(), Name: list[i].SubmittedBy.String()}
		}
		out = append(out, item)
	}
	return out, total, page, size, nil
}
