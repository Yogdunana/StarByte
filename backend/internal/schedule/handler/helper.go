package handler

import (
	"strings"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func currentUser(c *gin.Context) (uuid.UUID, error) {
	raw := auth.GetUserID(c)
	if raw == "" {
		return uuid.Nil, response.NewUnauthorizedError("用户未认证")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的用户ID")
	}
	return id, nil
}

func parseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的ID")
	}
	return id, nil
}

func parseNamedID(c *gin.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的ID")
	}
	return id, nil
}

func dataScope(c *gin.Context) *rbacModel.DataScopeCondition {
	return middleware.GetDataScopeFromContext(c)
}

func requestOrigin(c *gin.Context) string {
	host := c.Request.Host
	if host == "" {
		return ""
	}
	proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	return proto + "://" + host
}
