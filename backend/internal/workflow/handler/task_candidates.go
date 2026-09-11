package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (h *TaskHandler) TransferCandidates(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的任务ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	if len([]rune(keyword)) > 100 {
		response.BadRequest(c, "搜索条件过长")
		return
	}
	result, err := h.taskService.TransferCandidates(c.Request.Context(), id, keyword)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
