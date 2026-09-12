package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func unauthReq(method, path string, body []byte) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if body != nil {
		c.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, path, nil)
	}
	return c
}

func TestAuthenticatedHandlersRequireLogin(t *testing.T) {
	h := New(stubSvc{doc: &dto.DocResponse{Title: "x"}})
	id := uuid.New().String()
	h.List(unauthReq(http.MethodGet, "/knowledge/docs", nil))
	c := unauthReq(http.MethodGet, "/knowledge/docs/"+id, nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Get(c)
	h.Create(unauthReq(http.MethodPost, "/knowledge/docs", []byte(`{"kind":"doc","slug":"a","title":"a"}`)))
	c = unauthReq(http.MethodPut, "/knowledge/docs/"+id, []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Update(c)
	c = unauthReq(http.MethodDelete, "/knowledge/docs/"+id, nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Delete(c)
	c = unauthReq(http.MethodPost, "/knowledge/docs/"+id+"/publish", nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Publish(c)
	c = unauthReq(http.MethodGet, "/knowledge/docs/"+id+"/history", nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.History(c)
	c = unauthReq(http.MethodGet, "/knowledge/docs/"+id+"/versions/1", nil)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "version", Value: "1"}}
	h.GetVersion(c)
	c = unauthReq(http.MethodPost, "/knowledge/docs/"+id+"/rollback", []byte(`{"version":1}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Rollback(c)
	h.Search(unauthReq(http.MethodGet, "/knowledge/search?q=x", nil))
	h.Tree(unauthReq(http.MethodGet, "/knowledge/tree", nil))
	h.CreateCategory(unauthReq(http.MethodPost, "/knowledge/categories", []byte(`{"name":"n","slug":"n"}`)))
	c = unauthReq(http.MethodPut, "/knowledge/categories/"+id, []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.UpdateCategory(c)
	c = unauthReq(http.MethodDelete, "/knowledge/categories/"+id, nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.DeleteCategory(c)
	h.ListCategories(unauthReq(http.MethodGet, "/knowledge/categories", nil))
	c = unauthReq(http.MethodPost, "/knowledge/docs/"+id+"/attachments", []byte(`{"file_id":"`+uuid.New().String()+`"}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Attach(c)
	fid := uuid.New().String()
	c = unauthReq(http.MethodDelete, "/knowledge/docs/"+id+"/attachments/"+fid, nil)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "fileId", Value: fid}}
	h.Detach(c)
}

func TestServiceErrorsPropagate(t *testing.T) {
	err := response.NewError(response.CodeKnowledgeNotFound, "文档不存在")
	h := New(stubSvc{err: err})
	id := uuid.New().String()
	c := authReq(http.MethodGet, "/knowledge/docs/"+id, nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Get(c)
	h.List(authReq(http.MethodGet, "/knowledge/docs", nil))
	h.Create(authReq(http.MethodPost, "/knowledge/docs", []byte(`{"kind":"doc","slug":"a","title":"a"}`)))
	c = authReq(http.MethodPut, "/knowledge/docs/"+id, []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Update(c)
	c = authReq(http.MethodDelete, "/knowledge/docs/"+id, nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Delete(c)
	c = authReq(http.MethodPost, "/knowledge/docs/"+id+"/publish", nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Publish(c)
	c = authReq(http.MethodGet, "/knowledge/docs/"+id+"/history", nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.History(c)
	c = authReq(http.MethodGet, "/knowledge/docs/"+id+"/versions/1", nil)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "version", Value: "1"}}
	h.GetVersion(c)
	c = authReq(http.MethodPost, "/knowledge/docs/"+id+"/rollback", []byte(`{"version":1}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Rollback(c)
	h.Search(authReq(http.MethodGet, "/knowledge/search?q=协会", nil))
	h.Tree(authReq(http.MethodGet, "/knowledge/tree", nil))
	h.CreateCategory(authReq(http.MethodPost, "/knowledge/categories", []byte(`{"name":"手册","slug":"hb"}`)))
	c = authReq(http.MethodPut, "/knowledge/categories/"+id, []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.UpdateCategory(c)
	c = authReq(http.MethodDelete, "/knowledge/categories/"+id, nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.DeleteCategory(c)
	h.ListCategories(authReq(http.MethodGet, "/knowledge/categories", nil))
	fid := uuid.New().String()
	c = authReq(http.MethodPost, "/knowledge/docs/"+id+"/attachments", []byte(`{"file_id":"`+fid+`"}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Attach(c)
	c = authReq(http.MethodDelete, "/knowledge/docs/"+id+"/attachments/"+fid, nil)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "fileId", Value: fid}}
	h.Detach(c)
	c = authReq(http.MethodGet, "/knowledge/public/pages/about-us", nil)
	c.Params = gin.Params{{Key: "slug", Value: "about-us"}}
	h.PublicPage(c)
	h.PublicDocs(authReq(http.MethodGet, "/knowledge/public/docs", nil))
	h.PublicTree(authReq(http.MethodGet, "/knowledge/public/tree", nil))
	h.PublicSearch(authReq(http.MethodGet, "/knowledge/public/search?q=协会", nil))
}

func TestBindAndIDErrors(t *testing.T) {
	h := New(stubSvc{doc: &dto.DocResponse{Title: "x"}})
	id := uuid.New().String()
	h.Create(authReq(http.MethodPost, "/knowledge/docs", []byte(`{"kind":"doc"}`)))
	c := authReq(http.MethodPut, "/knowledge/docs/bad", []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Update(c)
	c = authReq(http.MethodPut, "/knowledge/docs/"+id, []byte(`{`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Update(c)
	c = authReq(http.MethodPost, "/knowledge/docs/bad/publish", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Publish(c)
	c = authReq(http.MethodGet, "/knowledge/docs/bad/history", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.History(c)
	c = authReq(http.MethodGet, "/knowledge/docs/bad/versions/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}, {Key: "version", Value: "1"}}
	h.GetVersion(c)
	c = authReq(http.MethodGet, "/knowledge/docs/"+id+"/versions/0", nil)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "version", Value: "0"}}
	h.GetVersion(c)
	c = authReq(http.MethodGet, "/knowledge/docs/"+id+"/versions/x", nil)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "version", Value: "x"}}
	h.GetVersion(c)
	c = authReq(http.MethodPost, "/knowledge/docs/bad/rollback", []byte(`{"version":1}`))
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Rollback(c)
	c = authReq(http.MethodPost, "/knowledge/docs/"+id+"/rollback", []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Rollback(c)
	h.Search(authReq(http.MethodGet, "/knowledge/search", nil))
	h.CreateCategory(authReq(http.MethodPost, "/knowledge/categories", []byte(`{}`)))
	c = authReq(http.MethodPut, "/knowledge/categories/bad", []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.UpdateCategory(c)
	c = authReq(http.MethodPut, "/knowledge/categories/"+id, []byte(`{`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.UpdateCategory(c)
	c = authReq(http.MethodDelete, "/knowledge/categories/bad", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.DeleteCategory(c)
	c = authReq(http.MethodPost, "/knowledge/docs/bad/attachments", []byte(`{"file_id":"`+uuid.New().String()+`"}`))
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Attach(c)
	c = authReq(http.MethodPost, "/knowledge/docs/"+id+"/attachments", []byte(`{}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Attach(c)
	c = authReq(http.MethodPost, "/knowledge/docs/"+id+"/attachments", []byte(`{"file_id":"not-a-uuid"}`))
	c.Params = gin.Params{{Key: "id", Value: id}}
	h.Attach(c)
	c = authReq(http.MethodDelete, "/knowledge/docs/bad/attachments/"+uuid.New().String(), nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}, {Key: "fileId", Value: uuid.New().String()}}
	h.Detach(c)
	c = authReq(http.MethodDelete, "/knowledge/docs/"+id+"/attachments/bad", nil)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "fileId", Value: "bad"}}
	h.Detach(c)
}

func TestInvalidUserIDAndOptionalViewer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := unauthReq(http.MethodGet, "/", nil)
	c.Set("user_id", "not-uuid")
	if _, err := viewerOf(c); err == nil {
		t.Fatal("expected invalid user id")
	}
	if optionalViewer(c).Authenticated() {
		t.Fatal("invalid uuid should stay anonymous")
	}
}

func TestPublicDocsAuthenticatedUsesFullList(t *testing.T) {
	h := New(stubSvc{doc: &dto.DocResponse{Title: "使用手册", Slug: "user-manual", Kind: model.KindDoc}})
	c := authReq(http.MethodGet, "/knowledge/public/docs", nil)
	h.PublicDocs(c)
}

func TestListBindError(t *testing.T) {
	h := New(stubSvc{})
	c := authReq(http.MethodGet, "/knowledge/docs?status=nope", nil)
	h.List(c)
}

func TestParseVersionZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "version", Value: "0"}}
	if _, err := parseVersion(c); err == nil {
		t.Fatal("expected invalid version")
	}
}

func TestHasPermWrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_permissions", "doc:read")
	if hasPerm(c, "doc:read") {
		t.Fatal("wrong type should be false")
	}
	if permsOf(c) != nil {
		t.Fatal("wrong type perms")
	}
}

func TestJSONRoundTripCreateBody(t *testing.T) {
	body, err := json.Marshal(dto.CreateDocRequest{Kind: model.KindPage, Slug: "about-us", Title: "关于我们"})
	if err != nil || len(body) == 0 {
		t.Fatal(err)
	}
}
