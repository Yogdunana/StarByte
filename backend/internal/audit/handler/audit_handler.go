package handler

import (
	"net/http"
	"strconv"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService service.AuditService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// List 审计日志列表
// @Summary 获取审计日志列表
// @Description 分页查询审计日志列表，支持按用户名、操作类型、HTTP方法、路径、IP、状态码、时间范围筛选
// @Tags 审计日志
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param username query string false "用户名"
// @Param operation query string false "操作类型"
// @Param method query string false "HTTP方法"
// @Param path query string false "请求路径"
// @Param ip query string false "IP地址"
// @Param request_id query string false "请求ID"
// @Param status_min query int false "状态码最小值"
// @Param status_max query int false "状态码最大值"
// @Param start_time query string false "开始时间(RFC3339)"
// @Param end_time query string false "结束时间(RFC3339)"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.AuditLogListResponse}}
// @Router /audit-logs [get]
func (h *AuditHandler) List(c *gin.Context) {
	var req dto.ListAuditLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.auditService.Query(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, list, total, req.Page, req.PageSize)
}

// GetByID 审计日志详情
// @Summary 获取审计日志详情
// @Description 根据ID获取审计日志详细信息
// @Tags 审计日志
// @Produce json
// @Security Bearer
// @Param id path string true "审计日志ID"
// @Success 200 {object} response.Response{data=dto.AuditLogResponse}
// @Router /audit-logs/{id} [get]
func (h *AuditHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的审计日志ID")
		return
	}

	result, err := h.auditService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Export 导出审计日志
// @Summary 导出审计日志
// @Description 导出审计日志为 CSV 或 JSON 格式
// @Tags 审计日志
// @Produce application/octet-stream
// @Security Bearer
// @Param format query string true "导出格式(csv或json)" Enums(csv, json)
// @Param username query string false "用户名"
// @Param operation query string false "操作类型"
// @Param method query string false "HTTP方法"
// @Param path query string false "请求路径"
// @Param ip query string false "IP地址"
// @Param status_min query int false "状态码最小值"
// @Param status_max query int false "状态码最大值"
// @Param start_time query string false "开始时间(RFC3339)"
// @Param end_time query string false "结束时间(RFC3339)"
// @Success 200 {file} file "导出文件"
// @Router /audit-logs/export [get]
func (h *AuditHandler) Export(c *gin.Context) {
	var req dto.ExportAuditLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	data, filename, err := h.auditService.Export(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// TriggerArchive 手动触发归档（管理员）
// @Summary 手动触发归档
// @Description 手动触发审计日志归档操作
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.ArchiveRequest false "归档参数"
// @Success 200 {object} response.Response{data=dto.ArchiveResponse}
// @Router /audit-logs/archive [post]
func (h *AuditHandler) TriggerArchive(c *gin.Context) {
	var req dto.ArchiveRequest
	// 允许空 body，使用默认值
	_ = c.ShouldBindJSON(&req)

	if req.BeforeDays <= 0 {
		req.BeforeDays = 90
	}

	result, err := h.auditService.Archive(c.Request.Context(), req.BeforeDays)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// ParseQueryInt 安全解析查询参数为 int
func ParseQueryInt(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
