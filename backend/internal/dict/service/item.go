package service

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/dict/dto"
	"github.com/Yogdunana/StarByte/backend/internal/dict/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *dictService) CreateItem(ctx context.Context, req dto.CreateItemRequest) (*dto.ItemResponse, error) {
	typ, err := s.requireTypeByCode(ctx, strings.TrimSpace(req.TypeCode))
	if err != nil {
		return nil, err
	}
	value := strings.TrimSpace(req.ItemValue)
	if err := s.ensureValueFree(ctx, typ.ID, value, uuid.Nil); err != nil {
		return nil, err
	}
	rec := &model.DictItem{
		ID: uuid.New(), TypeID: typ.ID, ItemValue: value,
		ItemLabel: strings.TrimSpace(req.ItemLabel), SortOrder: req.SortOrder, Status: req.Status,
	}
	if err := s.repo.CreateItem(ctx, rec); err != nil {
		return nil, err
	}
	s.invalidate(ctx, typ.Code)
	out := toItemDTO(rec, typ.Code)
	return &out, nil
}

func (s *dictService) UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateItemRequest) (*dto.ItemResponse, error) {
	rec, err := s.requireItem(ctx, id)
	if err != nil {
		return nil, err
	}
	typ, err := s.requireTypeByID(ctx, rec.TypeID)
	if err != nil {
		return nil, err
	}
	if req.ItemLabel != nil {
		rec.ItemLabel = strings.TrimSpace(*req.ItemLabel)
	}
	if req.SortOrder != nil {
		rec.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		rec.Status = *req.Status
	}
	if err := s.repo.UpdateItem(ctx, rec); err != nil {
		return nil, err
	}
	s.invalidate(ctx, typ.Code)
	out := toItemDTO(rec, typ.Code)
	return &out, nil
}

func (s *dictService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	rec, err := s.requireItem(ctx, id)
	if err != nil {
		return err
	}
	typ, _ := s.repo.GetTypeByID(ctx, rec.TypeID)
	if err := s.repo.DeleteItem(ctx, id); err != nil {
		return err
	}
	if typ != nil {
		s.invalidate(ctx, typ.Code)
	}
	return nil
}

func (s *dictService) ensureValueFree(ctx context.Context, typeID uuid.UUID, value string, skip uuid.UUID) error {
	items, err := s.repo.ListItemsByTypeID(ctx, typeID, false)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].ItemValue == value && items[i].ID != skip {
			return response.NewError(response.CodeDictItemExists, "字典项值已存在")
		}
	}
	return nil
}
