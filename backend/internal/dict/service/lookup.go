package service

import (
	"context"
	"errors"

	"github.com/Yogdunana/StarByte/backend/internal/dict/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *dictService) requireTypeByID(ctx context.Context, id uuid.UUID) (*model.DictType, error) {
	rec, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewError(response.CodeDictTypeNotFound, "字典类型不存在")
		}
		return nil, err
	}
	return rec, nil
}

func (s *dictService) requireTypeByCode(ctx context.Context, code string) (*model.DictType, error) {
	rec, err := s.repo.GetTypeByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewError(response.CodeDictTypeNotFound, "字典类型不存在")
		}
		return nil, err
	}
	return rec, nil
}

func (s *dictService) requireItem(ctx context.Context, id uuid.UUID) (*model.DictItem, error) {
	rec, err := s.repo.GetItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewError(response.CodeDictItemNotFound, "字典项不存在")
		}
		return nil, err
	}
	return rec, nil
}
