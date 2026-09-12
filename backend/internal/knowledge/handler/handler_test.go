package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubSvc struct {
	doc *dto.DocResponse
	err error
}

func (s stubSvc) CreateCategory(context.Context, service.Viewer, *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	return &dto.CategoryResponse{Name: "c"}, s.err
}
func (s stubSvc) UpdateCategory(context.Context, service.Viewer, uuid.UUID, *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	return &dto.CategoryResponse{Name: "c"}, s.err
}
func (s stubSvc) DeleteCategory(context.Context, service.Viewer, uuid.UUID) error { return s.err }
func (s stubSvc) ListCategories(context.Context) ([]*dto.CategoryResponse, error) {
	return []*dto.CategoryResponse{}, s.err
}
func (s stubSvc) Create(context.Context, service.Viewer, *dto.CreateDocRequest) (*dto.DocResponse, error) {
	return s.doc, s.err
}
func (s stubSvc) Update(context.Context, service.Viewer, uuid.UUID, *dto.UpdateDocRequest) (*dto.DocResponse, error) {
	return s.doc, s.err
}
func (s stubSvc) Delete(context.Context, service.Viewer, uuid.UUID) error { return s.err }
func (s stubSvc) Get(context.Context, service.Viewer, uuid.UUID) (*dto.DocResponse, error) {
	return s.doc, s.err
}
func (s stubSvc) GetBySlug(context.Context, service.Viewer, string, string) (*dto.DocResponse, error) {
	return s.doc, s.err
}
func (s stubSvc) List(context.Context, service.Viewer, *dto.ListDocRequest) ([]*dto.DocResponse, int64, error) {
	if s.doc == nil {
		return nil, 0, s.err
	}
	return []*dto.DocResponse{s.doc}, 1, s.err
}
func (s stubSvc) Publish(context.Context, service.Viewer, uuid.UUID) (*dto.DocResponse, error) {
	return s.doc, s.err
}
func (s stubSvc) History(context.Context, service.Viewer, uuid.UUID) ([]*dto.VersionResponse, error) {
	return []*dto.VersionResponse{}, s.err
}
func (s stubSvc) GetVersion(context.Context, service.Viewer, uuid.UUID, int) (*dto.VersionResponse, error) {
	return &dto.VersionResponse{Version: 1}, s.err
}
func (s stubSvc) Rollback(context.Context, service.Viewer, uuid.UUID, int) (*dto.DocResponse, error) {
	return s.doc, s.err
}
func (s stubSvc) Search(context.Context, service.Viewer, *dto.SearchRequest) ([]*dto.DocResponse, int64, error) {
	return nil, 0, s.err
}
func (s stubSvc) Tree(context.Context, service.Viewer) ([]*dto.TreeNode, error) {
	return []*dto.TreeNode{}, s.err
}
func (s stubSvc) Attach(context.Context, service.Viewer, uuid.UUID, uuid.UUID) (*dto.DocResponse, error) {
	return s.doc, s.err
}
func (s stubSvc) Detach(context.Context, service.Viewer, uuid.UUID, uuid.UUID) error {
	return s.err
}

func authReq(method, path string, body []byte) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	c.Request = req
	c.Set("user_id", uuid.New().String())
	c.Set("user_permissions", []string{"*"})
	return c
}

func TestPublicPageAndDocs(t *testing.T) {
	h := New(stubSvc{doc: &dto.DocResponse{Title: "关于我们", Slug: "about-us", Kind: model.KindPage}})
	c := authReq(http.MethodGet, "/knowledge/public/pages/about-us", nil)
	c.Params = gin.Params{{Key: "slug", Value: "about-us"}}
	h.PublicPage(c)
	c2 := authReq(http.MethodGet, "/knowledge/public/docs", nil)
	h.PublicDocs(c2)
	c3 := authReq(http.MethodGet, "/knowledge/public/tree", nil)
	h.PublicTree(c3)
}

func TestAuthenticatedCRUD(t *testing.T) {
	id := uuid.New()
	h := New(stubSvc{doc: &dto.DocResponse{ID: id.String(), Title: "x", Slug: "x", Kind: model.KindDoc}})
	c := authReq(http.MethodGet, "/knowledge/docs", nil)
	h.List(c)
	c = authReq(http.MethodGet, "/knowledge/docs/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.Get(c)
	body, _ := json.Marshal(dto.CreateDocRequest{Kind: model.KindDoc, Slug: "n", Title: "n"})
	c = authReq(http.MethodPost, "/knowledge/docs", body)
	h.Create(c)
	upd, _ := json.Marshal(dto.UpdateDocRequest{})
	c = authReq(http.MethodPut, "/knowledge/docs/"+id.String(), upd)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.Update(c)
	c = authReq(http.MethodPost, "/knowledge/docs/"+id.String()+"/publish", nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.Publish(c)
	c = authReq(http.MethodGet, "/knowledge/docs/"+id.String()+"/history", nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.History(c)
	c = authReq(http.MethodGet, "/knowledge/tree", nil)
	h.Tree(c)
	c = authReq(http.MethodGet, "/knowledge/categories", nil)
	h.ListCategories(c)
	c = authReq(http.MethodDelete, "/knowledge/docs/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.Delete(c)
}

func TestPublicSearchRequiresQuery(t *testing.T) {
	h := New(stubSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/knowledge/public/search", nil)
	h.PublicSearch(c)
	if w.Code == http.StatusOK {
		t.Fatal("expected bind error")
	}
}

func TestRollbackAndAttach(t *testing.T) {
	id := uuid.New()
	fileID := uuid.New()
	h := New(stubSvc{doc: &dto.DocResponse{ID: id.String()}})
	body, _ := json.Marshal(dto.RollbackRequest{Version: 1})
	c := authReq(http.MethodPost, "/knowledge/docs/"+id.String()+"/rollback", body)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.Rollback(c)
	att, _ := json.Marshal(dto.AttachRequest{FileID: fileID.String()})
	c = authReq(http.MethodPost, "/knowledge/docs/"+id.String()+"/attachments", att)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.Attach(c)
	c = authReq(http.MethodDelete, "/knowledge/docs/"+id.String()+"/attachments/"+fileID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}, {Key: "fileId", Value: fileID.String()}}
	h.Detach(c)
	c = authReq(http.MethodGet, "/knowledge/docs/"+id.String()+"/versions/1", nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}, {Key: "version", Value: "1"}}
	h.GetVersion(c)
}

func TestCategoriesAndSearch(t *testing.T) {
	h := New(stubSvc{doc: &dto.DocResponse{Title: "s"}})
	body, _ := json.Marshal(dto.CreateCategoryRequest{Name: "手册", Slug: "handbook"})
	c := authReq(http.MethodPost, "/knowledge/categories", body)
	h.CreateCategory(c)
	id := uuid.New()
	upd, _ := json.Marshal(dto.UpdateCategoryRequest{})
	c = authReq(http.MethodPut, "/knowledge/categories/"+id.String(), upd)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.UpdateCategory(c)
	c = authReq(http.MethodDelete, "/knowledge/categories/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.DeleteCategory(c)
	c = authReq(http.MethodGet, "/knowledge/search?q=章程", nil)
	h.Search(c)
	c = authReq(http.MethodGet, "/knowledge/public/docs/charter", nil)
	c.Params = gin.Params{{Key: "slug", Value: "charter"}}
	h.PublicDoc(c)
	c = authReq(http.MethodGet, "/knowledge/public/search?q=协会", nil)
	h.PublicSearch(c)
}

func TestBadIDs(t *testing.T) {
	h := New(stubSvc{})
	c := authReq(http.MethodGet, "/knowledge/docs/bad", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Get(c)
	c = authReq(http.MethodDelete, "/knowledge/docs/bad", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Delete(c)
	c = authReq(http.MethodPost, "/knowledge/docs", []byte(`{`))
	h.Create(c)
}
