package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouter creates a test gin engine with ErrorHandler middleware.
func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(RequestID())
	r.Use(ErrorHandler())
	return r
}

func TestErrorHandler_NormalRequest(t *testing.T) {
	r := setupRouter()
	r.GET("/ok", func(c *gin.Context) {
		response.OK(c, gin.H{"msg": "hello"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeSuccess, resp.Code)
}

func TestErrorHandler_PanicRecovery(t *testing.T) {
	r := setupRouter()
	r.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeInternalError, resp.Code)
	assert.Equal(t, "内部错误", resp.Message)
}

func TestErrorHandler_UnhandledError(t *testing.T) {
	r := setupRouter()
	r.GET("/error", func(c *gin.Context) {
		// Push error without writing response — ErrorHandler should handle it
		_ = c.Error(response.NewNotFoundError("user"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeNotFound, resp.Code)
}

func TestErrorHandler_PreWrittenResponse(t *testing.T) {
	r := setupRouter()
	r.GET("/written", func(c *gin.Context) {
		// Handler already wrote a response
		response.BadRequest(c, "custom bad request")
		// Also push an error — should NOT override the written response
		_ = c.Error(response.NewForbiddenError("should be ignored"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/written", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeBadRequest, resp.Code)
	assert.Equal(t, "custom bad request", resp.Message)
}

func TestErrorHandler_NoErrors(t *testing.T) {
	r := setupRouter()
	r.GET("/clean", func(c *gin.Context) {
		response.OKWithoutData(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clean", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeSuccess, resp.Code)
}

func TestErrorHandler_PanicWithWrittenResponse(t *testing.T) {
	r := setupRouter()
	r.GET("/panic-after-write", func(c *gin.Context) {
		response.OK(c, gin.H{"partial": true})
		panic("panic after write")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-after-write", nil)
	r.ServeHTTP(w, req)

	// The response was already written, so the panic shouldn't override it
	// (the writer was already flushed with 200)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeSuccess, resp.Code)
}

func TestErrorHandler_AppErrorPanic(t *testing.T) {
	r := setupRouter()
	r.GET("/apperror-panic", func(c *gin.Context) {
		panic(response.NewConflictError("test conflict"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/apperror-panic", nil)
	r.ServeHTTP(w, req)

	// Panics are always caught as 500, regardless of the panic value type
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeInternalError, resp.Code)
}

func TestErrorHandler_MultipleErrors(t *testing.T) {
	r := setupRouter()
	r.GET("/multi-error", func(c *gin.Context) {
		_ = c.Error(response.NewValidationError("field1", "required"))
		_ = c.Error(response.NewNotFoundError("user"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/multi-error", nil)
	r.ServeHTTP(w, req)

	// The last error should be used
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeNotFound, resp.Code)
}

func TestErrorHandler_RequestIDInResponse(t *testing.T) {
	r := setupRouter()
	r.GET("/with-rid", func(c *gin.Context) {
		panic("test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/with-rid", nil)
	req.Header.Set("X-Request-Id", "my-custom-id")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "my-custom-id", resp.RequestID)
}
