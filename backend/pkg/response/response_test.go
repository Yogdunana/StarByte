package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newContext creates a test gin context with a request_id set.
func newContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-request-id")
	return c, w
}

// parseResponse extracts the Response struct from the recorder body.
func parseResponse(t *testing.T, w *httptest.ResponseRecorder) Response {
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return resp
}

// ========== Constructor tests ==========

func TestNewError(t *testing.T) {
	err := NewError(CodeBadRequest, "something went wrong")
	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "something went wrong", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, "something went wrong", err.Error())
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("username", "is required")
	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "username: is required", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("user")
	assert.Equal(t, CodeNotFound, err.Code)
	assert.Contains(t, err.Message, "user")
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("token expired")
	assert.Equal(t, CodeUnauthorized, err.Code)
	assert.Equal(t, "token expired", err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus)
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("no access")
	assert.Equal(t, CodeForbidden, err.Code)
	assert.Equal(t, http.StatusForbidden, err.HTTPStatus)
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("already exists")
	assert.Equal(t, CodeConflict, err.Code)
	assert.Equal(t, http.StatusConflict, err.HTTPStatus)
}

// ========== Response helper tests ==========

func TestOK(t *testing.T) {
	c, w := newContext()
	OK(c, gin.H{"id": 1})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeSuccess, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, "test-request-id", resp.RequestID)
	assert.NotNil(t, resp.Data)
}

func TestOKWithoutData(t *testing.T) {
	c, w := newContext()
	OKWithoutData(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeSuccess, resp.Code)
	assert.Nil(t, resp.Data)
}

func TestDegraded(t *testing.T) {
	c, w := newContext()
	Degraded(c, gin.H{"fallback": true})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeDegraded, resp.Code)
	assert.Equal(t, "服务降级，已返回兜底数据", resp.Message)
	assert.NotNil(t, resp.Data)
}

func TestBadRequest(t *testing.T) {
	c, w := newContext()
	BadRequest(c, "invalid param")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeBadRequest, resp.Code)
	assert.Equal(t, "invalid param", resp.Message)
}

func TestUnauthorized(t *testing.T) {
	c, w := newContext()
	Unauthorized(c, "not logged in")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeUnauthorized, resp.Code)
}

func TestForbidden(t *testing.T) {
	c, w := newContext()
	Forbidden(c, "no permission")
	assert.Equal(t, http.StatusForbidden, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeForbidden, resp.Code)
}

func TestNotFound(t *testing.T) {
	c, w := newContext()
	NotFound(c, "resource not found")
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeNotFound, resp.Code)
}

func TestConflict(t *testing.T) {
	c, w := newContext()
	Conflict(c, "duplicate entry")
	assert.Equal(t, http.StatusConflict, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeConflict, resp.Code)
}

func TestPage(t *testing.T) {
	c, w := newContext()
	Page(c, []string{"a", "b"}, 100, 1, 10)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code      int          `json:"code"`
		Message   string       `json:"message"`
		Data      PageResponse `json:"data"`
		RequestID string       `json:"request_id"`
		Timestamp int64        `json:"timestamp"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)
	assert.Equal(t, int64(100), resp.Data.Total)
	assert.Equal(t, 1, resp.Data.Page)
	assert.Equal(t, 10, resp.Data.PageSize)
}

// ========== Error() resolution tests ==========

func TestError_WithAppError(t *testing.T) {
	c, w := newContext()
	err := NewNotFoundError("user")
	Error(c, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeNotFound, resp.Code)
	assert.Contains(t, resp.Message, "user")
}

func TestError_WithValidationError(t *testing.T) {
	c, w := newContext()
	err := NewValidationError("email", "invalid format")
	Error(c, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeBadRequest, resp.Code)
	assert.Contains(t, resp.Message, "email")
}

func TestError_WithWrappedAppError(t *testing.T) {
	c, w := newContext()
	original := NewConflictError("role already exists")
	wrapped := errors.New("context: " + original.Error()) // not the wrapped form

	// Use fmt.Errorf with %w to wrap
	wrapped = wrapErr(original)
	Error(c, wrapped)
	assert.Equal(t, http.StatusConflict, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeConflict, resp.Code)
}

func TestError_WithDomainError(t *testing.T) {
	c, w := newContext()
	err := &mockDomainError{code: 3001, message: "角色不存在"}
	Error(c, err)
	assert.Equal(t, http.StatusBadRequest, w.Code) // 3xxx → default 400
	resp := parseResponse(t, w)
	assert.Equal(t, 3001, resp.Code)
	assert.Equal(t, "角色不存在", resp.Message)
}

func TestError_WithDomainErrorAndHTTPStatus(t *testing.T) {
	c, w := newContext()
	err := &mockDomainErrorWithStatus{
		mockDomainError: mockDomainError{code: 3001, message: "角色不存在"},
		httpStatus:      http.StatusNotFound,
	}
	Error(c, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, 3001, resp.Code)
}

func TestError_WithUnknownError(t *testing.T) {
	c, w := newContext()
	err := errors.New("some random error")
	Error(c, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeInternalError, resp.Code)
}

func TestError_WithNilError(t *testing.T) {
	c, w := newContext()
	Error(c, nil)
	// nil doesn't implement any interface, falls through to 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ========== httpStatusFromCode tests ==========

func TestHttpStatusFromCode(t *testing.T) {
	tests := []struct {
		code   int
		expect int
	}{
		{CodeSuccess, http.StatusOK},
		{CodeBadRequest, http.StatusBadRequest},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeWorkflowTaskNoAccess, http.StatusForbidden},
		{CodeVoteNoAccess, http.StatusForbidden},
		{CodeVoteResultPending, http.StatusForbidden},
		{CodeMeetingNotAttendee, http.StatusForbidden},
		{CodeScheduleNoAccess, http.StatusForbidden},
		{CodeCalendarNoAccess, http.StatusForbidden},
		{CodeLeaveNoAccess, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeScheduleNotFound, http.StatusNotFound},
		{CodeCalendarNotFound, http.StatusNotFound},
		{CodeLeaveNotFound, http.StatusNotFound},
		{CodeLeaveTypeNotFound, http.StatusNotFound},
		{CodeBackupNotFound, http.StatusNotFound},
		{CodeConflict, http.StatusConflict},
		{CodeBackupBusy, http.StatusConflict},
		{CodeTooManyReq, http.StatusTooManyRequests},
		{CodeRateLimited, http.StatusTooManyRequests},
		{CodeCircuitOpen, http.StatusServiceUnavailable},
		{CodeBlacklisted, http.StatusForbidden},
		{CodeDegraded, http.StatusOK},
		{CodeInternalError, http.StatusInternalServerError},
		{CodeNotImplemented, http.StatusNotImplemented},
		{5999, http.StatusBadRequest}, // audit module code → 400
		{2001, http.StatusBadRequest}, // module code → default 400
		{3001, http.StatusBadRequest}, // RBAC code → default 400
		{9999, http.StatusBadRequest}, // unknown → default 400
	}

	for _, tt := range tests {
		got := httpStatusFromCode(tt.code)
		assert.Equal(t, tt.expect, got, "code %d", tt.code)
	}
}

// ========== ModuleRanges tests ==========

func TestCodeInternalErrorUsesGeneralRange(t *testing.T) {
	assert.Equal(t, 1500, CodeInternalError)
	assert.Equal(t, 1501, CodeNotImplemented)
	assert.Equal(t, 5001, CodeAuditNotFound)
	assert.Equal(t, 5005, CodeAuditArchiveNotFound)
	assert.Equal(t, 5008, CodeAuditEntityInvalid)
	assert.True(t, CodeInternalError >= 1000 && CodeInternalError <= 1999)
	assert.True(t, CodeAuditNotFound >= 5000 && CodeAuditNotFound <= 5999)
	assert.True(t, CodeAuditEntityInvalid >= 5000 && CodeAuditEntityInvalid <= 5999)
	assert.NotEqual(t, CodeInternalError, CodeAuditNotFound)
}

func TestNotImplemented(t *testing.T) {
	c, w := newContext()
	NotImplemented(c, "微信扫码登录功能暂未开通")
	assert.Equal(t, http.StatusNotImplemented, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, CodeNotImplemented, resp.Code)
	assert.Equal(t, "微信扫码登录功能暂未开通", resp.Message)
	assert.Equal(t, "test-request-id", resp.RequestID)
	assert.Nil(t, resp.Data)
}

func TestModuleRanges(t *testing.T) {
	r, ok := ModuleRanges["user"]
	assert.True(t, ok)
	assert.Equal(t, 2000, r[0])
	assert.Equal(t, 2999, r[1])

	r, ok = ModuleRanges["rbac"]
	assert.True(t, ok)
	assert.Equal(t, 3000, r[0])

	_, ok = ModuleRanges["nonexistent"]
	assert.False(t, ok)

	r, ok = ModuleRanges["export"]
	assert.True(t, ok)
	assert.Equal(t, 17000, r[0])
	assert.Equal(t, 17999, r[1])
	assert.Equal(t, 17001, CodeExportInvalidFormat)
	assert.Equal(t, 17006, CodeExportEmptyData)
	taskRange := ModuleRanges["task"]
	assert.Equal(t, 9000, taskRange[0])
	assert.True(t, r[0] > taskRange[1], "export must not collide with task 9000-9999")

	r, ok = ModuleRanges["cache"]
	assert.True(t, ok)
	assert.Equal(t, 18000, r[0])
	assert.Equal(t, 18999, r[1])
	assert.Equal(t, 18001, CodeCacheKeyNotFound)
	assert.True(t, r[0] > taskRange[1], "cache must not collide with task 9000-9999")

	r, ok = ModuleRanges["scheduler"]
	assert.True(t, ok)
	assert.Equal(t, 19000, r[0])
	assert.Equal(t, 19999, r[1])
	assert.Equal(t, 19001, CodeSchedulerNotFound)
	assert.True(t, r[0] > taskRange[1], "scheduler must not collide with task 9000-9999")

	r, ok = ModuleRanges["search"]
	assert.True(t, ok)
	assert.Equal(t, 20000, r[0])
	assert.Equal(t, 20999, r[1])
	assert.Equal(t, 20001, CodeSearchUnknownResource)
	assert.True(t, r[0] > taskRange[1], "search must not collide with task 9000-9999")

	r, ok = ModuleRanges["traffic"]
	assert.True(t, ok)
	assert.Equal(t, 21000, r[0])
	assert.Equal(t, 21999, r[1])
	assert.Equal(t, 21001, CodeRateLimited)
	assert.Equal(t, 21002, CodeCircuitOpen)
	assert.Equal(t, 21003, CodeBlacklisted)
	assert.Equal(t, 21004, CodeDegraded)
	assert.True(t, r[0] > taskRange[1], "traffic must not collide with task 9000-9999")

	r, ok = ModuleRanges["finance"]
	assert.True(t, ok)
	assert.Equal(t, 23000, r[0])
	assert.Equal(t, 23001, CodeFinanceNotFound)
	assert.Equal(t, 23005, CodeFinanceDirectionMismatch)

	r, ok = ModuleRanges["discipline"]
	assert.True(t, ok)
	assert.Equal(t, 24000, r[0])
	assert.Equal(t, 24001, CodeDisciplineNotFound)

	r, ok = ModuleRanges["contract"]
	assert.True(t, ok)
	assert.Equal(t, 25000, r[0])
	assert.Equal(t, 25001, CodeContractNotFound)

	r, ok = ModuleRanges["duty"]
	assert.True(t, ok)
	assert.Equal(t, 26000, r[0])
	assert.Equal(t, 26999, r[1])

	r, ok = ModuleRanges["activity"]
	assert.True(t, ok)
	assert.Equal(t, 27000, r[0])
	assert.Equal(t, 27999, r[1])
	assert.Equal(t, 27001, CodeActivityNotFound)
	assert.Equal(t, 27013, CodeCheckinGPSNotConfigured)
	assert.True(t, CodeActivityNotFound > ModuleRanges["duty"][1], "activity must not collide with duty 26000-26999")

	r, ok = ModuleRanges["announcement"]
	assert.True(t, ok)
	assert.Equal(t, 28000, r[0])
	assert.Equal(t, 28999, r[1])
	assert.Equal(t, 28001, CodeAnnouncementNotFound)
	assert.Equal(t, 28004, CodeAnnouncementInvalidCat)
	assert.True(t, r[0] > ModuleRanges["internship"][1], "announcement must not collide with internship 10000-10999")
	assert.True(t, r[0] > ModuleRanges["activity"][1], "announcement must not collide with activity 27000-27999")
	assert.True(t, r[1] < ModuleRanges["schedule"][0], "announcement must not collide with schedule 29000-29999")

	r, ok = ModuleRanges["schedule"]
	assert.True(t, ok)
	assert.Equal(t, 29000, r[0])
	assert.Equal(t, 29999, r[1])
	assert.Equal(t, 29001, CodeScheduleNotFound)
	assert.Equal(t, 29002, CodeScheduleNoAccess)
	assert.True(t, r[0] > ModuleRanges["internship"][1], "schedule must not collide with internship 10000-10999")
	assert.True(t, r[0] > ModuleRanges["activity"][1], "schedule must not collide with activity 27000-27999")

	r, ok = ModuleRanges["leave"]
	assert.True(t, ok)
	assert.Equal(t, 30000, r[0])
	assert.Equal(t, 30999, r[1])
	assert.Equal(t, 30001, CodeLeaveNotFound)
	assert.Equal(t, 30003, CodeLeaveNoAccess)
	assert.Equal(t, 30008, CodeLeaveBalanceMissing)
	assert.True(t, r[0] > ModuleRanges["schedule"][1], "leave must not collide with schedule 29000-29999")

	r, ok = ModuleRanges["monitor"]
	assert.True(t, ok)
	assert.Equal(t, 31000, r[0])
	assert.Equal(t, 31999, r[1])
	assert.Equal(t, 31001, CodeMonitorCollectFail)
	assert.Equal(t, 31002, CodeMonitorRedisDown)
	assert.Equal(t, 31003, CodeMonitorDBDown)
	assert.True(t, r[0] > ModuleRanges["leave"][1], "monitor must not collide with leave 30000-30999")

	r, ok = ModuleRanges["backup"]
	assert.True(t, ok)
	assert.Equal(t, 32000, r[0])
	assert.Equal(t, 32999, r[1])
	assert.Equal(t, 32001, CodeBackupNotFound)
	assert.Equal(t, 32003, CodeBackupBusy)
	assert.Equal(t, 32007, CodeBackupConfirmRequired)
	assert.True(t, r[0] > ModuleRanges["monitor"][1], "backup must not collide with monitor 31000-31999")
}

// ========== TranslateGORMError tests ==========

func TestTranslateGORMError_Nil(t *testing.T) {
	err := TranslateGORMError(nil)
	assert.Nil(t, err)
}

func TestTranslateGORMError_RecordNotFound(t *testing.T) {
	err := TranslateGORMError(gorm.ErrRecordNotFound)
	assert.NotNil(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, CodeNotFound, appErr.Code)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPStatus)
}

func TestTranslateGORMError_DuplicatedKey(t *testing.T) {
	err := TranslateGORMError(gorm.ErrDuplicatedKey)
	assert.NotNil(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, CodeConflict, appErr.Code)
	assert.Equal(t, http.StatusConflict, appErr.HTTPStatus)
}

func TestTranslateGORMError_OtherError(t *testing.T) {
	original := errors.New("connection refused")
	err := TranslateGORMError(original)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestTranslateGORMError_WrappedErrors(t *testing.T) {
	wrapped := wrapErr(gorm.ErrRecordNotFound)
	err := TranslateGORMError(wrapped)
	assert.NotNil(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, CodeNotFound, appErr.Code)
}

// ========== Helpers ==========

// wrapErr wraps an error with fmt.Errorf("...: %w", err) to test errors.Is.
func wrapErr(err error) error {
	return &wrappedError{inner: err, msg: "wrapped: " + err.Error()}
}

type wrappedError struct {
	inner error
	msg   string
}

func (w *wrappedError) Error() string { return w.msg }
func (w *wrappedError) Unwrap() error { return w.inner }

// mockDomainError simulates a module-specific domain error (e.g. rbac.Error)
type mockDomainError struct {
	code    int
	message string
}

func (e *mockDomainError) Error() string   { return e.message }
func (e *mockDomainError) Code() int       { return e.code }
func (e *mockDomainError) Message() string { return e.message }

// mockDomainErrorWithStatus adds HTTPStatus() to the domain error
type mockDomainErrorWithStatus struct {
	mockDomainError
	httpStatus int
}

func (e *mockDomainErrorWithStatus) HTTPStatus() int { return e.httpStatus }
