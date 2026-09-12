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
	p, s = defaultPage(3, 50)
	if p != 3 || s != 50 {
		t.Fatalf("got %d %d", p, s)
	}
}

func TestIsStaff(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if isStaff(c) {
		t.Fatal("empty context should not be staff")
	}
	c.Set("user_permissions", []string{"announcement:read"})
	if isStaff(c) {
		t.Fatal("read-only should not be staff")
	}
	c.Set("user_permissions", []string{"announcement:publish"})
	if !isStaff(c) {
		t.Fatal("publish should be staff")
	}
	c.Set("user_permissions", []string{"*"})
	if !isStaff(c) {
		t.Fatal("super admin should be staff")
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

func TestViewerOf_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := viewerOf(c); err == nil {
		t.Fatal("expected unauthorized")
	}
}
