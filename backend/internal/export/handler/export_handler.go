package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/internal/export/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	svc service.ExportService
}

func NewExportHandler(svc service.ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) ExportExcel(c *gin.Context) { h.exportTable(c, "excel") }
func (h *ExportHandler) ExportCSV(c *gin.Context)   { h.exportTable(c, "csv") }
func (h *ExportHandler) ExportPDF(c *gin.Context)   { h.exportTable(c, "pdf") }
func (h *ExportHandler) ExportJSON(c *gin.Context)  { h.exportTable(c, "json") }

func (h *ExportHandler) exportTable(c *gin.Context, format string) {
	var req dto.TableExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.ExportTable(c.Request.Context(), format, auth.GetUserID(c), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *ExportHandler) ExportTemplate(c *gin.Context) {
	templateID := strings.TrimSpace(c.Param("template_id"))
	if templateID == "" {
		response.BadRequest(c, "模板 ID 不能为空")
		return
	}
	var req dto.TemplateExportRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.ExportTemplate(c.Request.Context(), templateID, auth.GetUserID(c), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *ExportHandler) GetTask(c *gin.Context) {
	id := strings.TrimSpace(c.Param("task_id"))
	if id == "" {
		response.BadRequest(c, "任务 ID 不能为空")
		return
	}
	out, err := h.svc.GetTask(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *ExportHandler) ListTemplates(c *gin.Context) {
	response.OK(c, h.svc.ListTemplates())
}

func (h *ExportHandler) Download(c *gin.Context) {
	id := strings.TrimSpace(c.Param("file_id"))
	if id == "" {
		response.BadRequest(c, "文件 ID 不能为空")
		return
	}
	out, err := h.svc.Download(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	stream := c.Query("stream") == "1"
	if !stream && out.URL != "" {
		response.OK(c, dto.DownloadInfo{
			FileID:      out.FileID,
			Filename:    out.Filename,
			ContentType: out.ContentType,
			URL:         out.URL,
			ExpiresIn:   900,
		})
		return
	}
	if len(out.Bytes) == 0 {
		if out.URL != "" {
			c.Redirect(http.StatusFound, out.URL)
			return
		}
		response.Error(c, response.NewError(response.CodeExportFileExpired, "导出文件已过期"))
		return
	}
	ct := out.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Header("Content-Disposition", `attachment; filename="`+url.PathEscape(out.Filename)+`"`)
	c.Data(http.StatusOK, ct, out.Bytes)
}
