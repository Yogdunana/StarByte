package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// parseBindingError 解析参数绑定错误，返回具体的字段错误信息。
// 对于 validator 验证错误，提取字段名和失败的 tag；
// 对于其他类型错误，返回通用的"参数错误"。
func parseBindingError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		var msgs []string
		for _, fe := range ve {
			field := fe.Field()
			tag := fe.Tag()
			msgs = append(msgs, field+": "+tag+" 验证失败")
		}
		return strings.Join(msgs, "; ")
	}
	return "参数错误"
}

// getUserID 从 gin context 中获取当前用户 ID 并解析为 UUID。
// 如果用户未认证或 ID 无效，返回错误。
func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := auth.GetUserID(c)
	if userIDStr == "" {
		return uuid.Nil, response.NewAppError(response.CodeUnauthorized, "用户未认证")
	}
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, response.NewAppError(response.CodeBadRequest, "无效的用户ID")
	}
	return id, nil
}

// parsePagination 从 query 参数中解析分页参数。
// page 默认 1，pageSize 默认 10，最小值为 1。
func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

// parseUUIDParam 从路径参数中解析 UUID。
func parseUUIDParam(c *gin.Context, name string, errMsg string) (uuid.UUID, error) {
	idStr := c.Param(name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, response.NewAppError(response.CodeBadRequest, errMsg)
	}
	return id, nil
}
