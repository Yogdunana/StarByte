package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/internal/export/service"
	"github.com/Yogdunana/StarByte/backend/pkg/locale"
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

// ExportExcel 导出 Excel
// @Summary 导出 Excel
// @Description 将表格数据导出为 xlsx
// @Tags 导出
// @Accept json
// @Produce json
// @Param request body dto.TableExportRequest true "表格数据"
// @Success 200 {object} response.Response{data=dto.ExportTaskResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /export/excel [post]
// @Security BearerAuth
func (h *ExportHandler) ExportExcel(c *gin.Context) { h.exportTable(c, "excel") }

// ExportCSV 导出 CSV
// @Summary 导出 CSV
// @Description 将表格数据导出为 csv
// @Tags 导出
// @Accept json
// @Produce json
// @Param request body dto.TableExportRequest true "表格数据"
// @Success 200 {object} response.Response{data=dto.ExportTaskResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /export/csv [post]
// @Security BearerAuth
func (h *ExportHandler) ExportCSV(c *gin.Context) { h.exportTable(c, "csv") }

// ExportPDF 导出 PDF
// @Summary 导出 PDF
// @Description 将表格数据导出为 pdf
// @Tags 导出
// @Accept json
// @Produce json
// @Param request body dto.TableExportRequest true "表格数据"
// @Success 200 {object} response.Response{data=dto.ExportTaskResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /export/pdf [post]
// @Security BearerAuth
func (h *ExportHandler) ExportPDF(c *gin.Context) { h.exportTable(c, "pdf") }

// ExportJSON 导出 JSON
// @Summary 导出 JSON
// @Description 将表格数据导出为 json
// @Tags 导出
// @Accept json
// @Produce json
// @Param request body dto.TableExportRequest true "表格数据"
// @Success 200 {object} response.Response{data=dto.ExportTaskResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /export/json [post]
// @Security BearerAuth
func (h *ExportHandler) ExportJSON(c *gin.Context) { h.exportTable(c, "json") }

func (h *ExportHandler) exportTable(c *gin.Context, format string) {
	var req dto.TableExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Locale = locale.FromHeader(c.GetHeader("Accept-Language"))
	out, err := h.svc.ExportTable(c.Request.Context(), format, auth.GetUserID(c), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ExportTemplate 按模板导出
// @Summary 按模板导出
// @Description 使用内置 HTML 模板生成导出文件
// @Tags 导出
// @Accept json
// @Produce json
// @Param template_id path string true "模板 ID"
// @Param request body dto.TemplateExportRequest false "模板变量"
// @Success 200 {object} response.Response{data=dto.ExportTaskResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /export/template/{template_id} [post]
// @Security BearerAuth
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
	req.Locale = locale.FromHeader(c.GetHeader("Accept-Language"))
	out, err := h.svc.ExportTemplate(c.Request.Context(), templateID, auth.GetUserID(c), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetTask 导出任务状态
// @Summary 导出任务状态
// @Description 查询异步导出任务进度与结果文件
// @Tags 导出
// @Produce json
// @Param task_id path string true "任务 ID"
// @Success 200 {object} response.Response{data=dto.ExportTaskResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /export/tasks/{task_id} [get]
// @Security BearerAuth
func (h *ExportHandler) GetTask(c *gin.Context) {
	id := strings.TrimSpace(c.Param("task_id"))
	if id == "" {
		response.BadRequest(c, "任务 ID 不能为空")
		return
	}
	out, err := h.svc.GetTask(c.Request.Context(), id, auth.GetUserID(c), isSuperAdmin(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if out != nil {
		copy := *out
		copy.Error = locale.Message(locale.FromHeader(c.GetHeader("Accept-Language")), copy.Error)
		out = &copy
	}
	response.OK(c, out)
}

// ListTemplates 导出模板列表
// @Summary 导出模板列表
// @Description 列出内置打印/导出模板
// @Tags 导出
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /export/templates [get]
// @Security BearerAuth
func (h *ExportHandler) ListTemplates(c *gin.Context) {
	items := h.svc.ListTemplates()
	lang := locale.FromHeader(c.GetHeader("Accept-Language"))
	for i := range items {
		items[i].Name = locale.Text(lang, items[i].Name)
		items[i].Description = locale.Text(lang, items[i].Description)
	}
	response.OK(c, items)
}

// Download 下载导出文件
// @Summary 下载导出文件
// @Description 默认返回预签名信息；stream=1 时直接输出字节流
// @Tags 导出
// @Produce json
// @Param file_id path string true "文件 ID"
// @Param stream query int false "1 表示直接下载字节流"
// @Success 200 {object} response.Response{data=dto.DownloadInfo}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /export/download/{file_id} [get]
// @Security BearerAuth
func (h *ExportHandler) Download(c *gin.Context) {
	id := strings.TrimSpace(c.Param("file_id"))
	if id == "" {
		response.BadRequest(c, "文件 ID 不能为空")
		return
	}
	stream := c.Query("stream") == "1"
	out, err := h.svc.Download(c.Request.Context(), id, auth.GetUserID(c), isSuperAdmin(c), stream)
	if err != nil {
		response.Error(c, err)
		return
	}
	if !stream {
		response.OK(c, dto.DownloadInfo{
			FileID:      out.FileID,
			Filename:    out.Filename,
			ContentType: out.ContentType,
			ExpiresIn:   900,
		})
		return
	}
	if len(out.Bytes) == 0 {
		response.Error(c, response.NewError(response.CodeExportFileExpired, "导出文件已过期"))
		return
	}
	ct := out.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	escaped := url.PathEscape(out.Filename)
	c.Header("Content-Disposition", `attachment; filename="`+escaped+`"; filename*=UTF-8''`+escaped)
	c.Data(http.StatusOK, ct, out.Bytes)
}

func isSuperAdmin(c *gin.Context) bool {
	v, ok := c.Get("is_super_admin")
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}
