package response

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response is the standard JSON envelope returned by all API endpoints.
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
	Timestamp int64       `json:"timestamp"`
}

// PageResponse wraps a paginated list together with paging metadata.
type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// AppError is an application-level error that carries a response code,
// a human-readable message, and an explicit HTTP status code.
// It implements the error interface.
type AppError struct {
	Code       int    // business error code (see error_codes.go)
	Message    string // user-facing message
	HTTPStatus int    // corresponding HTTP status code
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// NewError constructs a new AppError with the given code and message.
// The HTTP status is inferred from the code via httpStatusFromCode().
func NewError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatusFromCode(code)}
}

// NewAppError is an alias for NewError, provided for ergonomic usage
// in domain code (e.g., response.NewAppError(response.CodeWorkflowNotFound, "msg")).
func NewAppError(code int, message string) *AppError {
	return NewError(code, message)
}

// NewAppErrorf constructs a new AppError with a formatted message.
func NewAppErrorf(code int, format string, args ...interface{}) *AppError {
	return NewError(code, fmt.Sprintf(format, args...))
}

// NewValidationError constructs a validation error for a specific field.
func NewValidationError(field, message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    field + ": " + message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewNotFoundError constructs a not-found error for the given resource.
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    resource + " 不存在",
		HTTPStatus: http.StatusNotFound,
	}
}

// NewUnauthorizedError constructs an unauthorized error.
func NewUnauthorizedError(msg string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    msg,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewForbiddenError constructs a forbidden error.
func NewForbiddenError(msg string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    msg,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewConflictError constructs a conflict error.
func NewConflictError(msg string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    msg,
		HTTPStatus: http.StatusConflict,
	}
}

// DomainError is an interface for module-specific errors that provide
// their own Code, Message, and optionally HTTPStatus.
// This allows service-layer errors (e.g. rbac.Error) to be mapped to
// HTTP responses without importing the response package.
type DomainError interface {
	error
	Code() int
	Message() string
}

// HTTPStatusError is an optional interface that DomainError implementations
// can satisfy to provide an explicit HTTP status. When not implemented,
// the status is inferred from the error code via httpStatusFromCode().
type HTTPStatusError interface {
	HTTPStatus() int
}

// OK sends a successful response with the given data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      data,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// OKWithoutData sends a successful response with no data payload.
func OKWithoutData(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// BadRequest sends a 400 response with the given message.
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:      CodeBadRequest,
		Message:   msg,
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// Unauthorized sends a 401 response with the given message.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:      CodeUnauthorized,
		Message:   msg,
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// Forbidden sends a 403 response with the given message.
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code:      CodeForbidden,
		Message:   msg,
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// NotFound sends a 404 response with the given message.
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Response{
		Code:      CodeNotFound,
		Message:   msg,
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// Conflict sends a 409 response with the given message.
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, Response{
		Code:      CodeConflict,
		Message:   msg,
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// NotImplemented sends a 501 response with the unified envelope (request_id included).
func NotImplemented(c *gin.Context, msg string) {
	c.JSON(http.StatusNotImplemented, Response{
		Code:      CodeNotImplemented,
		Message:   msg,
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// Error sends an appropriate response based on the error type.
//
// Error resolution order:
//  1. If the error implements DomainError (Code + Message), use those
//     values. If it also implements HTTPStatusError, use that status;
//     otherwise infer via httpStatusFromCode().
//  2. If the error is *AppError, use its Code, Message, and HTTPStatus.
//  3. Otherwise, return 500 Internal Server Error.
func Error(c *gin.Context, err error) {
	// 1. Check domain error (module-specific, e.g. rbac.Error)
	var domainErr DomainError
	if errors.As(err, &domainErr) {
		httpStatus := httpStatusFromCode(domainErr.Code())
		// Allow the domain error to override the inferred status
		var statusErr HTTPStatusError
		if errors.As(err, &statusErr) {
			httpStatus = statusErr.HTTPStatus()
		}
		c.JSON(httpStatus, Response{
			Code:      domainErr.Code(),
			Message:   domainErr.Message(),
			Data:      nil,
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 2. Check AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		status := appErr.HTTPStatus
		if status == 0 {
			status = httpStatusFromCode(appErr.Code)
		}
		c.JSON(status, Response{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Data:      nil,
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 3. Unknown error — treat as internal server error
	c.JSON(http.StatusInternalServerError, Response{
		Code:      CodeInternalError,
		Message:   "内部错误",
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// Page sends a successful paginated response.
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageResponse{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

// httpStatusFromCode maps an application error code to an HTTP status code.
//
// Mapping rules:
//   - 0            → 200 OK
//   - 1001         → 400 Bad Request
//   - 1002         → 401 Unauthorized
//   - 1003         → 403 Forbidden
//   - 1004         → 404 Not Found
//   - 1005         → 409 Conflict
//   - 1006         → 429 Too Many Requests
//   - 1500         → 500 Internal Server Error
//   - 1501         → 501 Not Implemented
//   - Other codes  → 400 Bad Request (business-level validation/flow errors)
func httpStatusFromCode(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeTooManyReq:
		return http.StatusTooManyRequests
	case CodeInternalError:
		return http.StatusInternalServerError
	case CodeNotImplemented:
		return http.StatusNotImplemented
	default:
		// All other codes (2xxx-12xxx) are module-specific business
		// errors and default to 400 Bad Request.
		return http.StatusBadRequest
	}
}
