package response

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Standard application response codes.
const (
	CodeSuccess       = 0
	CodeBadRequest    = 4000
	CodeUnauthorized  = 4001
	CodeForbidden     = 4003
	CodeNotFound      = 4004
	CodeConflict      = 4090
	CodeInternalError = 5000
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

// AppError is an application-level error that carries a response code and
// message. It implements the error interface.
type AppError struct {
	Code    int
	Message string
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// NewError constructs a new AppError with the given code and message.
func NewError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// NewValidationError constructs a validation error for a specific field.
func NewValidationError(field, message string) *AppError {
	return &AppError{Code: CodeBadRequest, Message: field + ": " + message}
}

// NewNotFoundError constructs a not-found error for the given resource.
func NewNotFoundError(resource string) *AppError {
	return &AppError{Code: CodeNotFound, Message: resource + " not found"}
}

// NewUnauthorizedError constructs an unauthorized error.
func NewUnauthorizedError(msg string) *AppError {
	return &AppError{Code: CodeUnauthorized, Message: msg}
}

// NewForbiddenError constructs a forbidden error.
func NewForbiddenError(msg string) *AppError {
	return &AppError{Code: CodeForbidden, Message: msg}
}

// NewConflictError constructs a conflict error.
func NewConflictError(msg string) *AppError {
	return &AppError{Code: CodeConflict, Message: msg}
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

// Error sends an appropriate response based on the error type. AppErrors are
// mapped to their corresponding HTTP status; any other error is treated as an
// internal server error.
func Error(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(httpStatusFromCode(appErr.Code), Response{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Data:      nil,
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Response{
		Code:      CodeInternalError,
		Message:   "Internal Server Error",
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

// httpStatusFromCode maps an application response code to an HTTP status code.
func httpStatusFromCode(code int) int {
	switch code {
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
	case CodeInternalError:
		return http.StatusInternalServerError
	default:
		// Business-level codes (e.g. the 2xxx range) describe client-side
		// validation/flow issues and default to a 400 response.
		if code >= CodeInternalError {
			return http.StatusInternalServerError
		}
		return http.StatusBadRequest
	}
}
