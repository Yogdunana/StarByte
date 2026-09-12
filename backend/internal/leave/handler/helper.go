package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/leave/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func defaultPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func hasPerm(c *gin.Context, code string) bool {
	raw, ok := c.Get("user_permissions")
	if !ok {
		return false
	}
	perms, _ := raw.([]string)
	for _, p := range perms {
		if p == "*" || p == code {
			return true
		}
	}
	return false
}

func viewerOf(c *gin.Context) (service.Viewer, error) {
	uid, err := getUserID(c)
	if err != nil {
		return service.Viewer{}, err
	}
	return service.Viewer{
		UserID:     uid,
		CanRead:    hasPerm(c, "leave:read"),
		CanApprove: hasPerm(c, "leave:approve"),
		Scope:      middleware.GetDataScopeFromContext(c),
	}, nil
}
