// Package handler HTTP 接入层
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 四层架构: handler -> service -> repo -> model
// 本文件定义统一响应封装、错误码到 HTTP 状态映射、JWT 中间件
//
// 设计原则:
//   1. handler 仅做协议转换 (HTTP <-> DTO)，不含业务逻辑
//   2. 统一响应格式，前端按 code 字段判断成功/失败
//   3. service 层 BizError 自动转为 ErrorResponse，handler 无需逐个 try-catch
//   4. actorID 由 JWT 中间件从 token 提取并注入 context.Context
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"schedule-service/internal/service"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// ============================================================================
// 统一响应封装
// ============================================================================

// ApiResponse 统一响应体结构
// 前端按 code 字段判断：0 表示成功，非 0 表示业务错误
type ApiResponse struct {
	// Code 业务码：0=成功，非0=失败（与 service.ErrCode 同值）
	Code int `json:"code"`
	// Message 面向用户的提示信息（脱敏，不含技术细节）
	Message string `json:"message"`
	// Data 业务数据，失败时为 null
	Data any `json:"data"`
	// RequestID 请求追踪ID，便于日志排查
	RequestID string `json:"request_id,omitempty"`
}

// OK 成功响应：封装业务数据
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// OKCreated 201 Created 响应：资源创建成功
func OKCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, ApiResponse{
		Code:    0,
		Message: "created",
		Data:    data,
	})
}

// OKList 分页列表响应
func OKList(c *gin.Context, list any, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"list":      list,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ============================================================================
// 错误响应
// ============================================================================

// errToHTTPStatus 错误码到 HTTP 状态码的映射规则
//   - 10501 参数校验失败     → 400 Bad Request
//   - 10502 未授权           → 401 Unauthorized
//   - 10503 无权限           → 403 Forbidden
//   - 10504/10601/10701/10801 资源不存在 → 404 Not Found
//   - 10603 并发冲突         → 409 Conflict
//   - 10604/10802 重复       → 409 Conflict
//   - 10606 已关联会议       → 409 Conflict
//   - 10607 状态转换非法     → 409 Conflict
//   - 10703 已触发           → 409 Conflict
//   - 其余业务错误           → 500 Internal Server Error
func errToHTTPStatus(code service.ErrCode) int {
	switch code {
	case service.ErrCodeInvalidParam,
		service.ErrCodeEventTimeInvalid,
		service.ErrCodeEventRRULEInvalid,
		service.ErrCodeReminderOffsetInvalid,
		service.ErrCodeReminderMethodUnsupported,
		service.ErrCodeAttendeeResponseInvalid,
		service.ErrCodeRRULEParseFailed:
		return http.StatusBadRequest
	case service.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case service.ErrCodeForbidden,
		service.ErrCodeAttendeeOrganizerRequired:
		return http.StatusForbidden
	case service.ErrCodeNotFound,
		service.ErrCodeEventNotFound,
		service.ErrCodeReminderNotFound,
		service.ErrCodeAttendeeNotFound:
		return http.StatusNotFound
	case service.ErrCodeEventConcurrent,
		service.ErrCodeEventDuplicate,
		service.ErrCodeEventMeetingLinked,
		service.ErrCodeEventStatusTransition,
		service.ErrCodeReminderAlreadyTriggered,
		service.ErrCodeReminderSnoozeExceed,
		service.ErrCodeAttendeeDuplicate,
		service.ErrCodeAttendeeCountExceed,
		service.ErrCodeRecurrenceRangeExceed:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Fail 业务错误响应：根据 BizError 生成对应 HTTP 响应
func Fail(c *gin.Context, err error) {
	// 尝试解包 BizError
	var bizErr *service.BizError
	if errors.As(err, &bizErr) {
		status := errToHTTPStatus(bizErr.Code)
		// 记录原始错误用于排查（包含 Err 字段）
		slog.Error("handler business error",
			"code", bizErr.Code,
			"message", bizErr.Message,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"err", bizErr.Err,
		)
		c.JSON(status, ApiResponse{
			Code:    int(bizErr.Code),
			Message: bizErr.Message,
			Data:    nil,
		})
		return
	}

	// 非 BizError：视为未预期错误，统一 500
	slog.Error("handler unexpected error",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"err", err,
	)
	c.JSON(http.StatusInternalServerError, ApiResponse{
		Code:    int(service.ErrCodeInternal),
		Message: "internal server error",
		Data:    nil,
	})
}

// ============================================================================
// JWT 中间件：从 token 提取 actorID 并注入 context
// ============================================================================

// stringCtxKeyActorID context key 的字符串形式（gin 内部用 map 存储）
const stringCtxKeyActorID = "schedule.actor_id"

// ActorIDFromContext 从 gin.Context 提取 actorID
//   handler 通过此函数获取当前操作者，透传给 service
//   未提取到则返回 0（service 层会拒绝）
func ActorIDFromContext(c *gin.Context) uint {
	if v, ok := c.Get(stringCtxKeyActorID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// JWTClaims 自定义 JWT Claims
//   包含用户ID字段 uid
type JWTClaims struct {
	UID uint `json:"uid"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware JWT 认证中间件
//   - 从 Authorization 头提取 Bearer token
//   - 解析 token 获得 userID
//   - 将 userID 写入 gin.Context，handler 通过 ActorIDFromContext 取出
//   - 解析失败返回 401
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	secret := []byte(jwtSecret)
	return func(c *gin.Context) {
		// 1. 提取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Fail(c, service.ErrUnauthorized)
			c.Abort()
			return
		}

		// 2. 校验 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			Fail(c, service.ErrUnauthorized)
			c.Abort()
			return
		}

		// 3. 解析 token 获得 userID
		userID, err := parseToken(parts[1], secret)
		if err != nil {
			slog.Warn("jwt parse failed",
				"path", c.Request.URL.Path,
				"err", err,
			)
			Fail(c, service.ErrUnauthorized)
			c.Abort()
			return
		}

		// 4. 注入 actorID 到 context
		c.Set(stringCtxKeyActorID, userID)
		c.Next()
	}
}

// parseToken 解析 JWT token 返回 userID
//   接入 github.com/golang-jwt/jwt/v5
//   claims 中应包含 "uid" 字段
func parseToken(tokenString string, secret []byte) (uint, error) {
	// 解析 token，使用自定义 claims
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		// 校验签名算法
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return 0, err
	}

	// 提取 claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token claims")
	}
	return claims.UID, nil
}

// GenerateToken 生成 JWT token（辅助函数，供登录/测试使用）
//   token 有效期默认 24 小时
func GenerateToken(userID uint, jwtSecret string) (string, error) {
	claims := JWTClaims{
		UID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ============================================================================
// 通用绑定辅助
// ============================================================================

// ShouldBindJSON 安全绑定 JSON 请求体
//   绑定失败时自动返回 400 错误响应
//   成功返回 true，失败返回 false（已写入响应，调用方应 return）
func ShouldBindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		slog.Warn("bind json failed",
			"path", c.Request.URL.Path,
			"err", err,
		)
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    int(service.ErrCodeInvalidParam),
			Message: "请求参数格式错误: " + err.Error(),
			Data:    nil,
		})
		return false
	}
	return true
}

// ShouldBindQuery 安全绑定查询参数
//   绑定失败时自动返回 400
func ShouldBindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		slog.Warn("bind query failed",
			"path", c.Request.URL.Path,
			"err", err,
		)
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    int(service.ErrCodeInvalidParam),
			Message: "查询参数格式错误: " + err.Error(),
			Data:    nil,
		})
		return false
	}
	return true
}

// ParseUintParam 从路径参数解析 uint
//   失败自动返回 400 并写入响应
//   与 parsePathUint 功能相同，保留为公开 API
func ParseUintParam(c *gin.Context, key string) (uint, bool) {
	return parsePathUint(c, key)
}

// parsePathUint 解析路径参数为 uint（内部使用）
//   失败自动返回 400 并写入响应
func parsePathUint(c *gin.Context, key string) (uint, bool) {
	raw := c.Param(key)
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    int(service.ErrCodeInvalidParam),
			Message: "路径参数 " + key + " 必须为正整数",
		})
		return 0, false
	}
	return uint(val), true
}

// parseTimeRFC3339 解析 RFC3339 时间字符串
//   用于 ListByTimeRange 的 from/to 查询参数
func parseTimeRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
