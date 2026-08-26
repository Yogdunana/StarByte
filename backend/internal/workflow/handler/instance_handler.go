package handler

import (
	"strconv"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/dto"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InstanceHandler 流程实例处理器
type InstanceHandler struct {
	instService service.InstanceService
}

// NewInstanceHandler 创建流程实例处理器
func NewInstanceHandler(instService service.InstanceService) *InstanceHandler {
	return &InstanceHandler{instService: instService}
}

// Start 启动流程实例
// @Summary 启动流程实例
// @Description 根据流程定义启动新的流程实例
// @Tags 流程实例
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.StartInstanceRequest true "启动参数"
// @Success 200 {object} response.Response{data=dto.InstanceResponse}
// @Router /workflow/instances [post]
func (h *InstanceHandler) Start(c *gin.Context) {
	var req dto.StartInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	initiatorID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	inst, err := h.instService.Start(c.Request.Context(), req.DefinitionID, req.BusinessKey, req.BusinessType, initiatorID, req.Variables)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.ToInstanceResponse(inst))
}

// List 流程实例列表
// @Summary 获取流程实例列表
// @Description 分页查询流程实例列表
// @Tags 流程实例
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query int false "状态"
// @Param definition_id query string false "流程定义ID"
// @Param initiator_id query string false "发起人ID"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.InstanceResponse}}
// @Router /workflow/instances [get]
func (h *InstanceHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)

	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = &s
		}
	}

	var definitionID *uuid.UUID
	if defIDStr := c.Query("definition_id"); defIDStr != "" {
		if id, err := uuid.Parse(defIDStr); err == nil {
			definitionID = &id
		}
	}

	var initiatorID *uuid.UUID
	if initIDStr := c.Query("initiator_id"); initIDStr != "" {
		if id, err := uuid.Parse(initIDStr); err == nil {
			initiatorID = &id
		}
	}

	list, total, err := h.instService.List(c.Request.Context(), page, pageSize, status, definitionID, initiatorID)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]dto.InstanceResponse, 0, len(list))
	for _, inst := range list {
		result = append(result, dto.ToInstanceResponse(&inst))
	}

	response.Page(c, result, total, page, pageSize)
}

// GetByID 获取流程实例详情
// @Summary 获取流程实例详情
// @Description 根据ID获取流程实例详细信息
// @Tags 流程实例
// @Produce json
// @Security Bearer
// @Param id path string true "流程实例ID"
// @Success 200 {object} response.Response{data=dto.InstanceResponse}
// @Router /workflow/instances/{id} [get]
func (h *InstanceHandler) GetByID(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程实例ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inst, err := h.instService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.ToInstanceResponse(inst))
}

// Terminate 终止流程实例
// @Summary 终止流程实例
// @Description 强制终止正在运行的流程实例
// @Tags 流程实例
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "流程实例ID"
// @Param request body dto.TerminateInstanceRequest true "终止原因"
// @Success 200 {object} response.Response
// @Router /workflow/instances/{id}/terminate [post]
func (h *InstanceHandler) Terminate(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程实例ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req dto.TerminateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	operatorID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.instService.Terminate(c.Request.Context(), id, operatorID, req.Reason); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Suspend 挂起流程实例
// @Summary 挂起流程实例
// @Description 挂起正在运行的流程实例
// @Tags 流程实例
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "流程实例ID"
// @Param request body dto.SuspendInstanceRequest true "挂起原因"
// @Success 200 {object} response.Response
// @Router /workflow/instances/{id}/suspend [post]
func (h *InstanceHandler) Suspend(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程实例ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req dto.SuspendInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	operatorID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.instService.Suspend(c.Request.Context(), id, operatorID, req.Reason); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Resume 恢复流程实例
// @Summary 恢复流程实例
// @Description 恢复已挂起的流程实例
// @Tags 流程实例
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "流程实例ID"
// @Success 200 {object} response.Response
// @Router /workflow/instances/{id}/resume [post]
func (h *InstanceHandler) Resume(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程实例ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	operatorID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.instService.Resume(c.Request.Context(), id, operatorID); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// ListHistory 获取流程历史记录
// @Summary 获取流程历史记录
// @Description 获取指定流程实例的历史操作记录
// @Tags 流程实例
// @Produce json
// @Security Bearer
// @Param id path string true "流程实例ID"
// @Success 200 {object} response.Response{data=[]dto.HistoryResponse}
// @Router /workflow/instances/{id}/history [get]
func (h *InstanceHandler) ListHistory(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程实例ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	histories, err := h.instService.ListHistory(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]dto.HistoryResponse, 0, len(histories))
	for _, hist := range histories {
		result = append(result, dto.ToHistoryResponse(&hist))
	}

	response.OK(c, result)
}
