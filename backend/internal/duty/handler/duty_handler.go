package handler

import (
	"net/http"
	"strconv"

	"github.com/Yogdunana/StarByte/backend/internal/duty/dto"
	"github.com/Yogdunana/StarByte/backend/internal/duty/service"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DutyHandler 值班处理器
type DutyHandler struct {
	svc service.DutyService
}

func NewDutyHandler(svc service.DutyService) *DutyHandler {
	return &DutyHandler{svc: svc}
}

// ListSchedules 排班列表
func (h *DutyHandler) ListSchedules(c *gin.Context) {
	var req dto.ListScheduleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListSchedules(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// CreateSchedule 创建排班
func (h *DutyHandler) CreateSchedule(c *gin.Context) {
	var req dto.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	creatorID := authmiddleware.GetUserID(c)
	result, err := h.svc.CreateSchedule(c.Request.Context(), &req, creatorID)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{
		Code:      response.CodeSuccess,
		Message:   "success",
		Data:      result,
		RequestID: c.GetString("request_id"),
	})
}

// BatchCreateSchedules 批量排班
func (h *DutyHandler) BatchCreateSchedules(c *gin.Context) {
	var req dto.BatchCreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	creatorID := authmiddleware.GetUserID(c)
	count, err := h.svc.BatchCreateSchedules(c.Request.Context(), &req, creatorID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"created": count})
}

// UpdateSchedule 调整排班（拖拽）
func (h *DutyHandler) UpdateSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的排班ID")
		return
	}
	var req dto.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateSchedule(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// CreateSwapRequest 调班申请
func (h *DutyHandler) CreateSwapRequest(c *gin.Context) {
	var req dto.SwapRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	requesterID := authmiddleware.GetUserID(c)
	result, err := h.svc.CreateSwapRequest(c.Request.Context(), &req, requesterID)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{
		Code:      response.CodeSuccess,
		Message:   "success",
		Data:      result,
		RequestID: c.GetString("request_id"),
	})
}

// ActionSwap 审批调班
func (h *DutyHandler) ActionSwap(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的调班申请ID")
		return
	}
	var req dto.SwapActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	approverID := authmiddleware.GetUserID(c)
	result, err := h.svc.ActionSwap(c.Request.Context(), id, &req, approverID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ListSwapRequests 调班申请列表
func (h *DutyHandler) ListSwapRequests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	requesterID := c.Query("requester_id")
	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}
	list, total, err := h.svc.ListSwapRequests(c.Request.Context(), page, pageSize, requesterID, status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// GetDutyStats 值班统计
func (h *DutyHandler) GetDutyStats(c *gin.Context) {
	var req dto.DutyStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.GetDutyStats(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
