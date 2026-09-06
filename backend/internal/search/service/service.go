package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/search/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/search"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxKeywordRunes = 200

type SearchService interface {
	Resources() []dto.ResourceInfo
	Lookup(code string) (search.Schema, bool)
	Query(ctx context.Context, req dto.QueryRequest, scope *rbacModel.DataScopeCondition, viewer uuid.UUID) (*search.Result, error)
}

type searchService struct {
	db     *gorm.DB
	engine *search.Engine
	byCode map[string]search.Schema
}

func NewSearchService(db *gorm.DB) SearchService {
	list := catalogs()
	by := make(map[string]search.Schema, len(list))
	for _, s := range list {
		by[s.Code] = s
	}
	return &searchService{db: db, engine: search.NewEngine(), byCode: by}
}

func (s *searchService) Resources() []dto.ResourceInfo {
	src := catalogs()
	out := make([]dto.ResourceInfo, 0, len(src))
	for _, sch := range src {
		out = append(out, dto.ResourceInfo{Code: sch.Code, Name: sch.Name, Fields: sch.PublicFields()})
	}
	return out
}

func (s *searchService) Lookup(code string) (search.Schema, bool) {
	sch, ok := s.byCode[strings.TrimSpace(code)]
	return sch, ok
}

func (s *searchService) Query(ctx context.Context, req dto.QueryRequest, scope *rbacModel.DataScopeCondition, viewer uuid.UUID) (*search.Result, error) {
	sch, ok := s.byCode[strings.TrimSpace(req.Resource)]
	if !ok {
		return nil, response.NewError(response.CodeSearchUnknownResource, "未知检索资源")
	}
	if utf8.RuneCountInString(req.Keyword) > maxKeywordRunes {
		return nil, response.NewError(response.CodeSearchInvalidQuery, "关键词过长")
	}
	sch = applyScope(sch, scope, viewer)
	out, err := s.engine.Search(ctx, s.db, sch, req.ToQuery())
	if err != nil {
		return nil, wrapSearchErr(err)
	}
	return out, nil
}

func applyScope(sch search.Schema, scope *rbacModel.DataScopeCondition, viewer uuid.UUID) search.Schema {
	if scope == nil || scope.IsEmpty() {
		return sch
	}
	return sch.ApplyDataScope(scope.Query, scope.Args, viewer, scope.IsSelf)
}

func wrapSearchErr(err error) error {
	switch {
	case errors.Is(err, search.ErrUnknownField):
		return response.NewError(response.CodeSearchUnknownField, err.Error())
	case errors.Is(err, search.ErrInvalidOp):
		return response.NewError(response.CodeSearchInvalidOp, err.Error())
	case errors.Is(err, search.ErrInvalidCursor):
		return response.NewError(response.CodeSearchInvalidCursor, err.Error())
	case errors.Is(err, search.ErrInvalidAgg):
		return response.NewError(response.CodeSearchInvalidAgg, err.Error())
	case errors.Is(err, search.ErrDeepPagination):
		return response.NewError(response.CodeSearchDeepPage, err.Error())
	case errors.Is(err, search.ErrInvalidQuery), errors.Is(err, search.ErrInvalidIdent):
		return response.NewError(response.CodeSearchInvalidQuery, err.Error())
	default:
		return fmtSearch(err)
	}
}

func fmtSearch(err error) error {
	return response.NewError(response.CodeSearchInvalidQuery, err.Error())
}
