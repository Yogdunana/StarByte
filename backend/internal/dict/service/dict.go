package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/dict/dto"
	"github.com/Yogdunana/StarByte/backend/internal/dict/model"
	"github.com/Yogdunana/StarByte/backend/internal/dict/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const cacheTTL = 10 * time.Minute

type DictService interface {
	ListTypes(ctx context.Context) ([]dto.TypeResponse, error)
	CreateType(ctx context.Context, req dto.CreateTypeRequest) (*dto.TypeResponse, error)
	UpdateType(ctx context.Context, id uuid.UUID, req dto.UpdateTypeRequest) (*dto.TypeResponse, error)
	DeleteType(ctx context.Context, id uuid.UUID) error
	ListItems(ctx context.Context, typeCode string, enabledOnly bool) ([]dto.ItemResponse, error)
	CreateItem(ctx context.Context, req dto.CreateItemRequest) (*dto.ItemResponse, error)
	UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateItemRequest) (*dto.ItemResponse, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
}

type dictService struct {
	repo  repo.DictRepository
	cache DictCache
}

func NewDictService(r repo.DictRepository, cache DictCache) DictService {
	return &dictService{repo: r, cache: cache}
}

func cacheKey(typeCode string) string {
	return "dict:" + strings.ToLower(strings.TrimSpace(typeCode))
}

func (s *dictService) invalidate(ctx context.Context, typeCode string) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Del(ctx, cacheKey(typeCode))
}

func (s *dictService) ListTypes(ctx context.Context) ([]dto.TypeResponse, error) {
	list, err := s.repo.ListTypes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.TypeResponse, 0, len(list))
	for i := range list {
		out = append(out, toTypeDTO(&list[i]))
	}
	return out, nil
}

func (s *dictService) CreateType(ctx context.Context, req dto.CreateTypeRequest) (*dto.TypeResponse, error) {
	code := strings.TrimSpace(req.Code)
	if existing, err := s.repo.GetTypeByCode(ctx, code); err == nil && existing != nil {
		return nil, response.NewError(response.CodeDictTypeExists, "字典类型编码已存在")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	rec := &model.DictType{
		ID: uuid.New(), Code: code, Name: strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description), SortOrder: req.SortOrder, Status: req.Status,
	}
	if err := s.repo.CreateType(ctx, rec); err != nil {
		return nil, err
	}
	out := toTypeDTO(rec)
	return &out, nil
}

func (s *dictService) UpdateType(ctx context.Context, id uuid.UUID, req dto.UpdateTypeRequest) (*dto.TypeResponse, error) {
	rec, err := s.requireTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		rec.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		rec.Description = strings.TrimSpace(*req.Description)
	}
	if req.SortOrder != nil {
		rec.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		rec.Status = *req.Status
	}
	if err := s.repo.UpdateType(ctx, rec); err != nil {
		return nil, err
	}
	s.invalidate(ctx, rec.Code)
	out := toTypeDTO(rec)
	return &out, nil
}

func (s *dictService) DeleteType(ctx context.Context, id uuid.UUID) error {
	rec, err := s.requireTypeByID(ctx, id)
	if err != nil {
		return err
	}
	if rec.IsSystem {
		return response.NewError(response.CodeDictSystemLocked, "系统字典类型不可删除")
	}
	if err := s.repo.DeleteType(ctx, id); err != nil {
		return err
	}
	s.invalidate(ctx, rec.Code)
	return nil
}

func (s *dictService) ListItems(ctx context.Context, typeCode string, enabledOnly bool) ([]dto.ItemResponse, error) {
	code := strings.TrimSpace(typeCode)
	if enabledOnly {
		if cached, ok := s.readCache(ctx, code); ok {
			return cached, nil
		}
	}
	typ, err := s.requireTypeByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListItemsByTypeID(ctx, typ.ID, enabledOnly)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ItemResponse, 0, len(items))
	for i := range items {
		out = append(out, toItemDTO(&items[i], typ.Code))
	}
	if enabledOnly {
		s.writeCache(ctx, code, out)
	}
	return out, nil
}

func (s *dictService) readCache(ctx context.Context, code string) ([]dto.ItemResponse, bool) {
	if s.cache == nil {
		return nil, false
	}
	raw, err := s.cache.Get(ctx, cacheKey(code))
	if err != nil || raw == "" {
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, false
		}
		return nil, false
	}
	var cached []dto.ItemResponse
	if json.Unmarshal([]byte(raw), &cached) != nil {
		return nil, false
	}
	return cached, true
}

func (s *dictService) writeCache(ctx context.Context, code string, out []dto.ItemResponse) {
	if s.cache == nil {
		return
	}
	if raw, err := json.Marshal(out); err == nil {
		_ = s.cache.Set(ctx, cacheKey(code), string(raw), cacheTTL)
	}
}

func toTypeDTO(rec *model.DictType) dto.TypeResponse {
	return dto.TypeResponse{
		ID: rec.ID.String(), Code: rec.Code, Name: rec.Name, Description: rec.Description,
		SortOrder: rec.SortOrder, Status: rec.Status, IsSystem: rec.IsSystem,
		CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
}

func toItemDTO(rec *model.DictItem, typeCode string) dto.ItemResponse {
	return dto.ItemResponse{
		ID: rec.ID.String(), TypeID: rec.TypeID.String(), TypeCode: typeCode,
		ItemValue: rec.ItemValue, ItemLabel: rec.ItemLabel, SortOrder: rec.SortOrder,
		Status: rec.Status, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
}
