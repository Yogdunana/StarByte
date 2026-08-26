package handler

import (
	"strconv"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/dto"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler 流程任务处理器
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler 创建流程任务处理器
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// ListTodoTasks 获取我的待办任务
// @Summary 获取我的待办任务
// @Description 分页查询当前用户的待办任务列表
// @Tags 流程任务
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.TaskResponse}}
// @Router /workflow/tasks/todo [get]
func (h *TaskHandler) ListTodoTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	userID := uuid.MustParse(auth.GetUserID(c))

	tasks, total, err := h.taskService.ListTodoTasks(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]dto.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, dto.ToTaskResponse(&task))
	}

	response.Page(c, result, total, page, pageSize)
}

// ListDoneTasks 获取我的已办任务
// @Summary 获取我的已办任务
// @Description 分页查询当前用户的已办任务列表
// @Tags 流程任务
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.TaskResponse}}
// @Router /workflow/tasks/done [get]
func (h *TaskHandler) ListDoneTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	userID := uuid.MustParse(auth.GetUserID(c))

	tasks, total, err := h.taskService.ListDoneTasks(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]dto.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, dto.ToTaskResponse(&task))
	}

	response.Page(c, result, total, page, pageSize)
}

// GetByID 获取任务详情
// @Summary 获取任务详情
// @Description 根据ID获取任务详细信息
// @Tags 流程任务
// @Produce json
// @Security Bearer
// @Param id path string true "任务ID"
// @Success 200 {object} response.Response{data=dto.TaskResponse}
// @Router /workflow/tasks/{id} [get]
func (h *TaskHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	task, err := h.taskService.GetTaskByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.ToTaskResponse(task))
}

// Approve 审批通过
// @Summary 审批通过
// @Description 审批通过指定任务
// @Tags 流程任务
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "任务ID"
// @Param request body dto.CompleteTaskRequest true "审批信息"
// @Success 200 {object} response.Response
// @Router /workflow/tasks/{id}/approve [post]
func (h *TaskHandler) Approve(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req dto.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	userID := uuid.MustParse(auth.GetUserID(c))

	if err := h.taskService.CompleteTask(c.Request.Context(), id, userID, "approve", req.Comment, req.Variables); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Reject 审批驳回
// @Summary 审批驳回
// @Description 驳回指定任务
// @Tags 流程任务
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "任务ID"
// @Param request body dto.CompleteTaskRequest true "驳回信息"
// @Success 200 {object} response.Response
// @Router /workflow/tasks/{id}/reject [post]
func (h *TaskHandler) Reject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req dto.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	userID := uuid.MustParse(auth.GetUserID(c))

	if err := h.taskService.CompleteTask(c.Request.Context(), id, userID, "reject", req.Comment, req.Variables); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Transfer 转办任务
// @Summary 转办任务
// @Description 将任务转交给其他用户处理
// @Tags 流程任务
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "任务ID"
// @Param request body dto.TransferTaskRequest true "转办信息"
// @Success 200 {object} response.Response
// @Router /workflow/tasks/{id}/transfer [post]
func (h *TaskHandler) Transfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req dto.TransferTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	fromUserID := uuid.MustParse(auth.GetUserID(c))

	if err := h.taskService.TransferTask(c.Request.Context(), id, fromUserID, req.TargetUserID, req.Comment); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Rollback 退回任务
// @Summary 退回任务
// @Description 将任务退回到指定节点
// @Tags 流程任务
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "任务ID"
// @Param request body dto.RollbackTaskRequest true "退回信息"
// @Success 200 {object} response.Response
// @Router /workflow/tasks/{id}/rollback [post]
func (h *TaskHandler) Rollback(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req dto.RollbackTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	userID := uuid.MustParse(auth.GetUserID(c))

	if err := h.taskService.RollbackTask(c.Request.Context(), id, userID, req.TargetNodeID, req.Comment); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}
