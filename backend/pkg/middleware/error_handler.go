package middleware

import (
	"net/http"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler is the unified global error-recovery middleware.
//
// It provides three layers of protection:
//  1. Panic recovery — catches panics from any downstream handler/middleware,
//     logs the error with request_id, and returns a 500 response.
//  2. Unhandled error collection — after handlers run, if c.Errors contains
//     any errors that were never written to the response, the last one is
//     resolved through response.Error() so the client receives a proper
//     JSON error envelope.
//  3. Error logging — all handled panics and unhandled errors are logged
//     with the request_id for traceability.
//
// Registration order (in main.go):
//
//	r.Use(middleware.RequestID())
//	r.Use(middleware.Logger())
//	r.Use(middleware.ErrorHandler())
//	r.Use(middleware.CORS())
//
// ErrorHandler is the unified global error-recovery middleware. It
// supersedes the removed Recovery() middleware.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID := c.GetString("request_id")

				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)

				// Only write if the response hasn't been written to yet.
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
						Code:      response.CodeInternalError,
						Message:   "内部错误",
						Data:      nil,
						RequestID: requestID,
						Timestamp: time.Now().Unix(),
					})
				}
				c.Abort()
				return
			}
		}()

		// Run the request chain.
		c.Next()

		// After all handlers have run, check for unhandled errors.
		// Handlers that use response.Error(c, err) write the response
		// themselves; handlers that use c.Error(err) push the error
		// without writing, expecting this middleware to handle it.
		if len(c.Errors) > 0 && !c.Writer.Written() {
			lastErr := c.Errors.Last().Err
			requestID := c.GetString("request_id")

			logger.Error("unhandled error",
				zap.Error(lastErr),
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)

			response.Error(c, lastErr)
		}
	}
}
