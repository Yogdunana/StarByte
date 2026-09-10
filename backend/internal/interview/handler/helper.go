package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/interview/service"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := auth.GetUserID(c)
	if userIDStr == "" {
		return uuid.Nil, response.NewUnauthorizedError("用户未认证")
	}
	id, err := uuid.Parse(userIDStr)
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
	if value, exists := c.Get("interview_scope"); exists {
		scope, _ := value.(*rbacModel.DataScopeCondition)
		return scope
	}
	return middleware.GetDataScopeFromContext(c)
}

func viewer(c *gin.Context) service.Viewer {
	id, _ := getUserID(c) // Nil ID fails closed in the service.
	var reviewScope *rbacModel.DataScopeCondition
	if _, exists := c.Get("interview_scope"); exists {
		reviewScope = middleware.GetDataScopeFromContext(c)
	}
	return service.Viewer{ID: id, Scope: dataScope(c), ReviewScope: reviewScope}
}

func defaultPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return page, pageSize
}
