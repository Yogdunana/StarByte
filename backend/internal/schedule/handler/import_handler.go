package handler

import (
	"io"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const maxImportUpload = 4 << 20

// ImportTimetable 导入 SMBU 课表 XLSX 到当前用户的 timetable 图层。
// semester_start：第 1 周星期一的 ISO 日期（YYYY-MM-DD）。标题中的学年学期仅作元数据。
// @Summary 导入课表
// @Tags 日程
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "xlsx"
// @Param semester_start formData string true "第1周星期一 YYYY-MM-DD"
// @Router /schedules/imports/timetable [post]
// @Security BearerAuth
func (h *Handler) ImportTimetable(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	raw, filename, err := readFormFile(c, "file")
	if err != nil {
		response.Error(c, err)
		return
	}
	start, err := parseSemesterStart(c.PostForm("semester_start"))
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.ImportTimetable(c.Request.Context(), userID, filename, raw, start, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ImportICS 导入 ICS 到指定或新建的 import 图层。
// @Summary 导入 ICS
// @Tags 日程
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "ics"
// @Param calendar_id formData string false "目标日历，空则新建 import 图层"
// @Router /schedules/imports/ics [post]
// @Security BearerAuth
func (h *Handler) ImportICS(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	raw, filename, err := readFormFile(c, "file")
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.ImportICS(c.Request.Context(), userID, filename, raw, c.PostForm("calendar_id"), dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) GoogleStatus(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GoogleStatus(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) GoogleConnect(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GoogleConnectURL(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) GoogleCallbackPOST(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	out, err := h.svc.GoogleCallback(c.Request.Context(), userID, req.Code, req.State, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) GoogleCallbackGET(c *gin.Context) {
	// 公开 GET 只回跳前端，由已登录的 POST /google/callback 绑定，避免把他人日历绑到 state 用户。
	dest := h.svc.FrontendCallbackRedirect(c.Query("code"), c.Query("state"))
	if dest == "" {
		response.Error(c, response.NewError(response.CodeUnauthorized, "请登录后通过 POST /schedules/google/callback 完成 Google 绑定"))
		return
	}
	c.Redirect(302, dest)
}

func (h *Handler) GoogleDisconnect(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.GoogleDisconnect(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handler) GoogleSync(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GoogleSync(c.Request.Context(), userID, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func readFormFile(c *gin.Context, field string) ([]byte, string, error) {
	header, err := c.FormFile(field)
	if err != nil {
		return nil, "", response.NewError(response.CodeScheduleImportInvalid, "请上传文件")
	}
	if header.Size > maxImportUpload {
		return nil, "", response.NewError(response.CodeScheduleImportInvalid, "文件超过 4MB")
	}
	f, err := header.Open()
	if err != nil {
		return nil, "", response.NewError(response.CodeScheduleImportInvalid, "无法读取上传文件")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, maxImportUpload+1))
	if err != nil {
		return nil, "", response.NewError(response.CodeScheduleImportInvalid, "读取文件失败")
	}
	if len(raw) > maxImportUpload {
		return nil, "", response.NewError(response.CodeScheduleImportInvalid, "文件超过 4MB")
	}
	return raw, header.Filename, nil
}

func parseSemesterStart(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, response.NewError(response.CodeScheduleImportInvalid, "semester_start 必填，格式 YYYY-MM-DD，表示第 1 周星期一")
	}
	loc := time.FixedZone("CST", 8*3600)
	if tz, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		loc = tz
	}
	if t, err := time.ParseInLocation("2006-01-02", raw, loc); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Time{}, response.NewError(response.CodeScheduleImportInvalid, "semester_start 须为 YYYY-MM-DD（第 1 周星期一）")
}
