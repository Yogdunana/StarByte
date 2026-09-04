package handler

import (
	"net/http"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditHandler struct {
	auditService service.AuditService
}

func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// List 审计日志列表
// @Summary 获取审计日志列表
// @Description 分页查询审计日志，支持时间、用户、动作、模块、关键词、IP 筛选
// @Tags 审计日志
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param start_time query string false "开始时间(RFC3339)"
// @Param end_time query string false "结束时间(RFC3339)"
// @Param user_id query string false "用户ID"
// @Param username query string false "用户名"
// @Param action query string false "动作 CREATE/UPDATE/DELETE/LOGIN/LOGOUT"
// @Param module query string false "模块"
// @Param keyword query string false "关键词"
// @Param ip_address query string false "IP"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.AuditLogListResponse}}
// @Router /system/audit-logs [get]
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
// @Tags 审计日志
// @Produce json
// @Security Bearer
// @Param id path string true "审计日志ID"
// @Success 200 {object} response.Response{data=dto.AuditLogResponse}
// @Router /system/audit-logs/{id} [get]
func (h *AuditHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
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
// @Description 导出 CSV 或 Excel，最多 10000 条
// @Tags 审计日志
// @Produce application/octet-stream
// @Security Bearer
// @Param format query string true "导出格式" Enums(csv, excel)
// @Success 200 {file} file "导出文件"
// @Router /system/audit-logs/export [get]
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

// TriggerArchive 手动触发归档
// @Summary 手动触发归档
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.ArchiveRequest false "归档参数"
// @Success 200 {object} response.Response{data=dto.ArchiveResponse}
// @Router /system/audit-logs/archive [post]
func (h *AuditHandler) TriggerArchive(c *gin.Context) {
	var req dto.ArchiveRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "参数错误: "+err.Error())
			return
		}
	}
	if req.BeforeDays <= 0 {
		req.BeforeDays = model.DefaultArchiveDays
	}
	result, err := h.auditService.Archive(c.Request.Context(), req.BeforeDays)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
