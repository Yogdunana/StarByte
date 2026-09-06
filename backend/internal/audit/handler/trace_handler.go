package handler

import (
	"net/http"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Trace 按实体查询变更历史
// @Summary 实体变更追踪
// @Description 实体变更追踪
// @Tags 审计日志
// @Produce json
// @Security BearerAuth
// @Param entity_type path string true "实体类型"
// @Param entity_id path string true "实体ID"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.AuditTraceItem}}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/audit-logs/traces/{entity_type}/{entity_id} [get]
func (h *AuditHandler) Trace(c *gin.Context) {
	var req dto.TraceQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.auditService.Trace(c.Request.Context(), c.Param("entity_type"), c.Param("entity_id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// Report 合规报告
// @Summary 审计合规报告
// @Description 审计合规报告
// @Tags 审计日志
// @Produce json
// @Security BearerAuth
// @Param format query string false "json/csv/pdf/excel"
// @Success 200 {object} response.Response{data=dto.ReportResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/audit-logs/reports [get]
func (h *AuditHandler) Report(c *gin.Context) {
	var req dto.ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	report, data, filename, err := h.auditService.Report(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" || format == "json" {
		response.OK(c, report)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, contentTypeOfReport(format), data)
}

// Archives 归档列表；带 id 时按需拉取 MinIO 对象
// @Summary 归档查询
// @Description 归档查询
// @Tags 审计日志
// @Produce json
// @Security BearerAuth
// @Param id query string false "归档ID，传入则拉取对象内容"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/audit-logs/archives [get]
func (h *AuditHandler) Archives(c *gin.Context) {
	var req dto.ArchiveListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.ID) != "" {
		id, err := uuid.Parse(req.ID)
		if err != nil {
			response.BadRequest(c, "无效的归档ID")
			return
		}
		result, err := h.auditService.PullArchive(c.Request.Context(), id, &req)
		if err != nil {
			response.Error(c, err)
			return
		}
		response.OK(c, result)
		return
	}
	list, total, err := h.auditService.ListArchives(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func contentTypeOfReport(format string) string {
	switch format {
	case "csv":
		return "text/csv; charset=utf-8"
	case "excel":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
