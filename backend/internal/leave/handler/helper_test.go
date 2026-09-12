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
	if p != 1 || s != 10 {
		t.Fatalf("got %d %d", p, s)
	}
	p, s = defaultPage(3, 50)
	if p != 3 || s != 50 {
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

func TestHasPermAndViewer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if hasPerm(c, "leave:read") {
		t.Fatal("empty should not have perm")
	}
	c.Set("user_permissions", []string{"leave:read"})
	if !hasPerm(c, "leave:read") || hasPerm(c, "leave:approve") {
		t.Fatal("read should not imply approve")
	}
	c.Set("user_permissions", []string{"*"})
	if !hasPerm(c, "leave:approve") {
		t.Fatal("super admin should match")
	}
	if _, err := viewerOf(c); err == nil {
		t.Fatal("expected unauthorized without user id")
	}
	id := uuid.New()
	c.Set("user_id", id.String())
	v, err := viewerOf(c)
	if err != nil {
		t.Fatal(err)
	}
	if v.UserID != id || !v.CanRead || !v.CanApprove {
		t.Fatalf("viewer %+v", v)
	}
}
