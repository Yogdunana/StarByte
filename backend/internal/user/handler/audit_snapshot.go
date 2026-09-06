package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/user/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func captureUserBefore(c *gin.Context, svc service.UserService, id uuid.UUID) {
	if c == nil || svc == nil {
		return
	}
	user, err := svc.GetByID(c.Request.Context(), id)
	if err != nil {
		return
	}
	middleware.SetAuditSnapshot(c, "user", id.String(), user)
}
