package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/service"
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
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
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

func permsOf(c *gin.Context) []string {
	raw, ok := c.Get("user_permissions")
	if !ok {
		return nil
	}
	perms, _ := raw.([]string)
	return perms
}

func viewerOf(c *gin.Context) (service.Viewer, error) {
	uid, err := getUserID(c)
	if err != nil {
		return service.Viewer{}, err
	}
	return service.Viewer{
		UserID:     uid,
		Perms:      permsOf(c),
		Roles:      rolesOf(c),
		CanCreate:  hasPerm(c, "doc:create"),
		CanUpdate:  hasPerm(c, "doc:update"),
		CanDelete:  hasPerm(c, "doc:delete"),
		CanPublish: hasPerm(c, "doc:publish"),
	}, nil
}

func optionalViewer(c *gin.Context) service.Viewer {
	uidStr := auth.GetUserID(c)
	if uidStr == "" {
		return service.Viewer{}
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return service.Viewer{}
	}
	return service.Viewer{
		UserID:     uid,
		Perms:      permsOf(c),
		Roles:      rolesOf(c),
		CanCreate:  hasPerm(c, "doc:create"),
		CanUpdate:  hasPerm(c, "doc:update"),
		CanDelete:  hasPerm(c, "doc:delete"),
		CanPublish: hasPerm(c, "doc:publish"),
	}
}

func rolesOf(c *gin.Context) []string {
	raw, _ := c.Get("current_roles")
	roles, _ := raw.([]string)
	return roles
}
