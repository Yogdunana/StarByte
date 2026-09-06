package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/internal/stats/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StatsHandler struct {
	svc service.StatsService
}

func NewStatsHandler(svc service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

// Providers 统计提供者列表
// @Summary 列出统计提供者
// @Tags 数据统计
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]dto.ProviderInfo}
// @Router /stats/providers [get]
func (h *StatsHandler) Providers(c *gin.Context) {
	response.OK(c, h.svc.ListProviders())
}

// Overview 首页概览
// @Summary 获取统计概览
// @Tags 数据统计
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=dto.OverviewResponse}
// @Router /stats/overview [get]
func (h *StatsHandler) Overview(c *gin.Context) {
	uid, _ := uuid.Parse(auth.GetUserID(c))
	res, err := h.svc.Overview(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// Get 按提供者查询统计
// @Summary 获取指定提供者的统计数据
// @Tags 数据统计
// @Produce json
// @Security Bearer
// @Param provider path string true "提供者名称"
// @Param start_date query string false "开始日期 YYYY-MM-DD"
// @Param end_date query string false "结束日期 YYYY-MM-DD"
// @Param department_id query string false "部门ID"
// @Param group_by query string false "分组 date/department/type/status/grade"
// @Param granularity query string false "粒度 day/week/month"
// @Success 200 {object} response.Response{data=dto.StatsResult}
// @Router /stats/{provider} [get]
func (h *StatsHandler) Get(c *gin.Context) {
	h.serve(c, c.Param("provider"))
}

func (h *StatsHandler) serveNamed(code string) gin.HandlerFunc {
	return func(c *gin.Context) { h.serve(c, code) }
}

func (h *StatsHandler) serve(c *gin.Context, code string) {
	q, err := bindQuery(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	res, err := h.svc.GetStats(c.Request.Context(), code, applyScope(c, q))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// Export 导出统计数据
// @Summary 导出统计数据 CSV/Excel
// @Tags 数据统计
// @Produce application/octet-stream
// @Security Bearer
// @Param provider path string true "提供者名称"
// @Param format query string false "csv 或 excel" Enums(csv, excel)
// @Success 200 {file} file "导出文件"
// @Router /stats/export/{provider} [get]
func (h *StatsHandler) Export(c *gin.Context) {
	q, err := bindQuery(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	format := c.DefaultQuery("format", "excel")
	data, filename, err := h.svc.Export(c.Request.Context(), c.Param("provider"), format, applyScope(c, q))
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func bindQuery(c *gin.Context) (*dto.StatsQuery, error) {
	q := &dto.StatsQuery{
		GroupBy:     strings.TrimSpace(c.Query("group_by")),
		Granularity: strings.TrimSpace(c.Query("granularity")),
		Format:      strings.TrimSpace(c.Query("format")),
	}
	start, err := parseDate(c.Query("start_date"), false)
	if err != nil {
		return nil, response.NewError(response.CodeStatsInvalidParam, "查询参数无效")
	}
	end, err := parseDate(c.Query("end_date"), true)
	if err != nil {
		return nil, response.NewError(response.CodeStatsInvalidParam, "查询参数无效")
	}
	q.StartDate = start
	q.EndDate = end
	if raw := strings.TrimSpace(c.Query("department_id")); raw != "" {
		id, perr := uuid.Parse(raw)
		if perr != nil {
			return nil, response.NewError(response.CodeStatsInvalidParam, "查询参数无效")
		}
		q.DepartmentID = &id
	}
	return q, nil
}

func applyScope(c *gin.Context, q *dto.StatsQuery) *dto.StatsQuery {
	if q == nil {
		q = &dto.StatsQuery{}
	}
	scope := middleware.GetDataScopeFromContext(c)
	if scope == nil || scope.IsEmpty() {
		q.AllScope = true
		return q
	}
	q.AllScope = false
	if scope.Query == "1 = 0" {
		q.Denied = true
		return q
	}
	q.ScopeDeptIDs = extractScopeDeptIDs(scope.Args)
	return q
}

func extractScopeDeptIDs(args []interface{}) []uuid.UUID {
	out := make([]uuid.UUID, 0)
	for _, arg := range args {
		switch v := arg.(type) {
		case uuid.UUID:
			out = append(out, v)
		case *uuid.UUID:
			if v != nil {
				out = append(out, *v)
			}
		case []uuid.UUID:
			out = append(out, v...)
		case string:
			if id, err := uuid.Parse(v); err == nil {
				out = append(out, id)
			}
		}
	}
	return out
}

func parseDate(raw string, endOfDay bool) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", raw, time.Local); err == nil {
		if endOfDay {
			t = t.Add(24*time.Hour - time.Nanosecond)
		}
		return &t, nil
	}
	return nil, response.NewError(response.CodeStatsInvalidParam, "查询参数无效")
}
