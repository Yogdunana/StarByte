package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func captureRoleBefore(c *gin.Context, id uuid.UUID, load func() (any, error)) {
	if c == nil || load == nil {
		return
	}
	val, err := load()
	if err != nil {
		return
	}
	middleware.SetAuditSnapshot(c, "role", id.String(), val)
}

func captureDeptBefore(c *gin.Context, id uuid.UUID, load func() (any, error)) {
	if c == nil || load == nil {
		return
	}
	val, err := load()
	if err != nil {
		return
	}
	middleware.SetAuditSnapshot(c, "department", id.String(), val)
}
