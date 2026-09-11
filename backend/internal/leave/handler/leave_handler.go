package handler

import (
	"leave-backend/internal/leave/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LeaveHandler 请假模块HTTP接口处理层
type LeaveHandler struct {
	service *service.LeaveService
}

// NewLeaveHandler 创建handler实例
func NewLeaveHandler(service *service.LeaveService) *LeaveHandler {
	return &LeaveHandler{service: service}
}

// RegisterRoutes 注册路由
func (h *LeaveHandler) RegisterRoutes(r *gin.Engine) {
	leaveGroup := r.Group("/api/leave")
	{
		// 请假类型
		leaveGroup.GET("/types", h.GetLeaveTypes)

		// 请假申请
		leaveGroup.POST("/applications", h.SubmitLeaveApplication)
		leaveGroup.GET("/applications/my", h.GetMyLeaveApplications)
		leaveGroup.GET("/applications/:id", h.GetLeaveApplicationByID)
		leaveGroup.PUT("/applications/:id/approve", h.ApproveLeaveApplication)
		leaveGroup.PUT("/applications/:id/reject", h.RejectLeaveApplication)
		leaveGroup.GET("/applications", h.GetLeaveApplicationsByStatus)

		// 假期余额
		leaveGroup.GET("/balances/:user_id", h.GetLeaveBalances)
	}
}

// ===================== 统一响应结构 =====================

// Response 统一API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// successResponse 构造成功响应
func successResponse(data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// errorResponse 构造失败响应
func errorResponse(msg string) *Response {
	return &Response{
		Code:    -1,
		Message: msg,
	}
}

// ===================== 请求参数结构体 =====================

// SubmitLeaveRequest 提交请假申请请求参数
type SubmitLeaveRequest struct {
	LeaveTypeID uint   `json:"leave_type_id" binding:"required"`           // 请假类型ID
	StartTime  string `json:"start_time" binding:"required"`               // 开始时间（2006-01-02 15:04）
	EndTime    string `json:"end_time" binding:"required"`                 // 结束时间（2006-01-02 15:04）
	Reason     string `json:"reason" binding:"required,min=1,max=500"`     // 请假事由
}

// ApproveRequest 审批请求参数
type ApproveRequest struct {
	ApproverID uint   `json:"approver_id" binding:"required"` // 审批人ID
	Remark     string `json:"remark"`                         // 审批备注
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ===================== 接口实现 =====================

// GetLeaveTypes 获取请假类型列表
// GET /api/leave/types
func (h *LeaveHandler) GetLeaveTypes(c *gin.Context) {
	types, err := h.service.GetAvailableLeaveTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("获取请假类型失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(types))
}

// SubmitLeaveApplication 提交请假申请
// POST /api/leave/applications
func (h *LeaveHandler) SubmitLeaveApplication(c *gin.Context) {
	var req SubmitLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("参数错误: "+err.Error()))
		return
	}

	startTime, err := time.ParseInLocation("2006-01-02 15:04", req.StartTime, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("开始时间格式错误，请使用 2006-01-02 15:04 格式"))
		return
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("结束时间格式错误，请使用 2006-01-02 15:04 格式"))
		return
	}

	// 从请求头获取申请人ID（实际项目中可从JWT token中解析）
	applicantIDStr := c.GetHeader("X-User-ID")
	if applicantIDStr == "" {
		c.JSON(http.StatusUnauthorized, errorResponse("缺少用户身份信息"))
		return
	}
	applicantID, err := strconv.ParseUint(applicantIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("用户ID格式错误"))
		return
	}

	application, err := h.service.SubmitLeaveRequest(uint(applicantID), req.LeaveTypeID, startTime, endTime, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, successResponse(application))
}

// GetLeaveApplicationByID 根据ID获取请假申请详情
// GET /api/leave/applications/:id
func (h *LeaveHandler) GetLeaveApplicationByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("申请ID格式错误"))
		return
	}

	application, err := h.service.GetLeaveApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(application))
}

// GetMyLeaveApplications 查询我的请假记录
// GET /api/leave/applications/my?user_id=1&page=1&page_size=10
func (h *LeaveHandler) GetMyLeaveApplications(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		userIDStr = c.GetHeader("X-User-ID")
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("用户ID格式错误"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	applications, total, err := h.service.GetMyLeaveRecords(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("查询请假记录失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(PaginatedResult{
		List:     applications,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}))
}

// ApproveLeaveApplication 审批通过请假申请
// PUT /api/leave/applications/:id/approve
func (h *LeaveHandler) ApproveLeaveApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("申请ID格式错误"))
		return
	}

	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("参数错误: "+err.Error()))
		return
	}

	if err := h.service.ApproveLeave(uint(id), req.ApproverID, req.Remark); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(nil))
}

// RejectLeaveApplication 驳回请假申请
// PUT /api/leave/applications/:id/reject
func (h *LeaveHandler) RejectLeaveApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("申请ID格式错误"))
		return
	}

	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("参数错误: "+err.Error()))
		return
	}

	if err := h.service.RejectLeave(uint(id), req.ApproverID, req.Remark); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(nil))
}

// GetLeaveApplicationsByStatus 按状态查询请假申请
// GET /api/leave/applications?status=pending&page=1&page_size=10
func (h *LeaveHandler) GetLeaveApplicationsByStatus(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	applications, total, err := h.service.GetLeaveApplicationsByStatus(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("查询失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(PaginatedResult{
		List:     applications,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}))
}

// GetLeaveBalances 获取用户假期余额
// GET /api/leave/balances/:user_id?year=2024
func (h *LeaveHandler) GetLeaveBalances(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("用户ID格式错误"))
		return
	}

	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("年份格式错误"))
		return
	}

	balances, err := h.service.GetMyLeaveBalances(uint(userID), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("查询假期余额失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(balances))
}
