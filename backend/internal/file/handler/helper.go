package handler

import (
	"strconv"

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

func parseFileID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的文件ID")
	}
	return id, nil
}

func parseBoolForm(c *gin.Context, key string) bool {
	v := c.PostForm(key)
	b, err := strconv.ParseBool(v)
	return err == nil && b
}
