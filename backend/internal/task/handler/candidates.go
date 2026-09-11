package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (h *TaskHandler) Candidates(c *gin.Context) {
	out, err := h.svc.Candidates(c.Request.Context(), c.Query("keyword"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TaskHandler) AssignmentRoles(c *gin.Context) {
	out, err := h.svc.AssignmentRoles(c.Request.Context(), c.Query("keyword"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
