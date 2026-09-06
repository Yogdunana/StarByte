package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的ID")
	}
	return id, nil
}

func (h *DictHandler) canReadAll(c *gin.Context) (bool, error) {
	userIDStr := auth.GetUserID(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false, response.NewUnauthorizedError("无效的用户身份")
	}
	perms, isSuper, err := h.cache.GetUserPermissionsAndSuperAdmin(c.Request.Context(), userID)
	if err != nil {
		return false, err
	}
	if isSuper {
		return true, nil
	}
	for _, p := range perms {
		if p == "dict:read" {
			return true, nil
		}
	}
	return false, nil
}
