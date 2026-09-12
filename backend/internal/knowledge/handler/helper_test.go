package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestDefaultPage(t *testing.T) {
	p, s := defaultPage(0, 0)
	if p != 1 || s != 20 {
		t.Fatalf("got %d %d", p, s)
	}
	p, s = defaultPage(2, 200)
	if p != 2 || s != 100 {
		t.Fatalf("got %d %d", p, s)
	}
}

func TestParseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	if _, err := parseID(c); err == nil {
		t.Fatal("expected invalid id")
	}
	id := uuid.New()
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	got, err := parseID(c)
	if err != nil || got != id {
		t.Fatalf("got %s err=%v", got, err)
	}
}

func TestHasPermAndOptionalViewer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if hasPerm(c, "doc:read") {
		t.Fatal("empty should not have perm")
	}
	if optionalViewer(c).Authenticated() {
		t.Fatal("anon")
	}
	c.Set("user_permissions", []string{"doc:read"})
	if !hasPerm(c, "doc:read") || hasPerm(c, "doc:publish") {
		t.Fatal("perm mismatch")
	}
	c.Set("user_permissions", []string{"*"})
	if !hasPerm(c, "doc:delete") {
		t.Fatal("star should match")
	}
	uid := uuid.New()
	c.Set("user_id", uid.String())
	v := optionalViewer(c)
	if v.UserID != uid || !v.CanDelete {
		t.Fatalf("viewer %+v", v)
	}
}

func TestViewerOfUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := viewerOf(c); err == nil {
		t.Fatal("expected unauthorized")
	}
}

func TestParseVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "version", Value: "2"}}
	n, err := parseVersion(c)
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	c.Params = gin.Params{{Key: "version", Value: "x"}}
	if _, err := parseVersion(c); err == nil {
		t.Fatal("expected bad version")
	}
}
