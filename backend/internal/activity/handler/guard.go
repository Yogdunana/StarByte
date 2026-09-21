package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mustAccessActivity 校验当前用户对该活动是否有数据权限；无权限时已写入响应并返回 false。
func (h *ActivityHandler) mustAccessActivity(c *gin.Context, id uuid.UUID) bool {
	viewer, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return false
	}
	scope := middleware.GetDataScopeFromContext(c)
	ok, err := h.svc.CanAccessActivity(c.Request.Context(), viewer, id, scope)
	if err != nil {
		response.Error(c, err)
		return false
	}
	if !ok {
		response.Error(c, response.NewForbiddenError("无权访问该活动"))
		return false
	}
	return true
}
