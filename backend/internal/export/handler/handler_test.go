package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	table *dto.ExportTaskResponse
	tpl   *dto.ExportTaskResponse
	task  *dto.ExportTaskResponse
	dl    *dto.DownloadResult
	list  []dto.TemplateInfo
	err   error
}

func (s *stubSvc) ExportTable(context.Context, string, string, *dto.TableExportRequest) (*dto.ExportTaskResponse, error) {
	return s.table, s.err
}
func (s *stubSvc) ExportTemplate(context.Context, string, string, *dto.TemplateExportRequest) (*dto.ExportTaskResponse, error) {
	return s.tpl, s.err
}
func (s *stubSvc) GetTask(context.Context, string) (*dto.ExportTaskResponse, error) {
	return s.task, s.err
}
func (s *stubSvc) Download(context.Context, string) (*dto.DownloadResult, error) {
	return s.dl, s.err
}
func (s *stubSvc) ListTemplates() []dto.TemplateInfo { return s.list }

func doJSON(h func(*gin.Context), method, path string, body any, params gin.Params) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	c.Request = httptest.NewRequest(method, path, &buf)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	c.Set("request_id", "rid")
	h(c)
	return w
}

func TestExportExcel_OK(t *testing.T) {
	h := NewExportHandler(&stubSvc{table: &dto.ExportTaskResponse{TaskID: "t1", Status: "done", FileID: "f1"}})
	w := doJSON(h.ExportExcel, http.MethodPost, "/api/v1/export/excel", dto.TableExportRequest{
		Columns: []string{"a"}, Rows: [][]string{{"1"}},
	}, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)
}

func TestExportCSV_BadJSON(t *testing.T) {
	h := NewExportHandler(&stubSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/export/csv", bytes.NewBufferString("{"))
	c.Set("request_id", "rid")
	h.ExportCSV(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTask_NotFound(t *testing.T) {
	h := NewExportHandler(&stubSvc{err: response.NewError(response.CodeExportTaskNotFound, "导出任务不存在")})
	w := doJSON(h.GetTask, http.MethodGet, "/api/v1/export/tasks/x", nil, gin.Params{{Key: "task_id", Value: "x"}})
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, response.CodeExportTaskNotFound, resp.Code)
}

func TestExportTemplate_MissingID(t *testing.T) {
	h := NewExportHandler(&stubSvc{})
	w := doJSON(h.ExportTemplate, http.MethodPost, "/api/v1/export/template/", nil, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportTemplate_NotFound(t *testing.T) {
	h := NewExportHandler(&stubSvc{err: response.NewError(response.CodeExportTplNotFound, "导出模板不存在")})
	w := doJSON(h.ExportTemplate, http.MethodPost, "/api/v1/export/template/x", dto.TemplateExportRequest{}, gin.Params{{Key: "template_id", Value: "x"}})
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, response.CodeExportTplNotFound, resp.Code)
}

func TestListTemplates(t *testing.T) {
	h := NewExportHandler(&stubSvc{list: []dto.TemplateInfo{{ID: "member_application", Name: "入会申请表"}}})
	w := doJSON(h.ListTemplates, http.MethodGet, "/api/v1/export/templates", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDownload_Stream(t *testing.T) {
	h := NewExportHandler(&stubSvc{dl: &dto.DownloadResult{
		FileID: "f1", Filename: "a.csv", ContentType: "text/csv", Bytes: []byte("hi"),
	}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/download/f1?stream=1", nil)
	c.Params = gin.Params{{Key: "file_id", Value: "f1"}}
	c.Set("request_id", "rid")
	h.Download(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hi", w.Body.String())
}

func TestDownload_JSONWithoutURL(t *testing.T) {
	h := NewExportHandler(&stubSvc{dl: &dto.DownloadResult{
		FileID: "f1", Filename: "a.csv", ContentType: "text/csv", Bytes: []byte("hi"),
	}})
	w := doJSON(h.Download, http.MethodGet, "/api/v1/export/download/f1", nil, gin.Params{{Key: "file_id", Value: "f1"}})
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)
}

func TestDownload_JSONWithURL(t *testing.T) {
	h := NewExportHandler(&stubSvc{dl: &dto.DownloadResult{
		FileID: "f1", Filename: "a.xlsx", ContentType: "application/octet-stream", URL: "http://minio/x", Bytes: []byte("x"),
	}})
	w := doJSON(h.Download, http.MethodGet, "/api/v1/export/download/f1", nil, gin.Params{{Key: "file_id", Value: "f1"}})
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)
}

func TestDownload_Expired(t *testing.T) {
	h := NewExportHandler(&stubSvc{err: response.NewError(response.CodeExportFileExpired, "导出文件已过期")})
	w := doJSON(h.Download, http.MethodGet, "/api/v1/export/download/f1", nil, gin.Params{{Key: "file_id", Value: "f1"}})
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, response.CodeExportFileExpired, resp.Code)
}

func TestExportPDFAndJSON_OK(t *testing.T) {
	h := NewExportHandler(&stubSvc{table: &dto.ExportTaskResponse{TaskID: "t1", Status: "done"}})
	assert.Equal(t, http.StatusOK, doJSON(h.ExportPDF, http.MethodPost, "/api/v1/export/pdf", dto.TableExportRequest{
		Columns: []string{"a"}, Rows: [][]string{{"1"}},
	}, nil).Code)
	assert.Equal(t, http.StatusOK, doJSON(h.ExportJSON, http.MethodPost, "/api/v1/export/json", dto.TableExportRequest{
		Columns: []string{"a"}, Rows: [][]string{{"1"}},
	}, nil).Code)
}

func TestGetTask_OK(t *testing.T) {
	h := NewExportHandler(&stubSvc{task: &dto.ExportTaskResponse{TaskID: "t1", Status: "running", Progress: 10}})
	w := doJSON(h.GetTask, http.MethodGet, "/api/v1/export/tasks/t1", nil, gin.Params{{Key: "task_id", Value: "t1"}})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDownload_Redirect(t *testing.T) {
	h := NewExportHandler(&stubSvc{dl: &dto.DownloadResult{
		FileID: "f1", Filename: "a.xlsx", URL: "http://minio/x",
	}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/download/f1?stream=1", nil)
	c.Params = gin.Params{{Key: "file_id", Value: "f1"}}
	c.Set("request_id", "rid")
	h.Download(c)
	assert.Equal(t, http.StatusFound, w.Code)
}

func TestDownload_EmptyExpired(t *testing.T) {
	h := NewExportHandler(&stubSvc{dl: &dto.DownloadResult{FileID: "f1"}})
	w := doJSON(h.Download, http.MethodGet, "/api/v1/export/download/f1?stream=1", nil, gin.Params{{Key: "file_id", Value: "f1"}})
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, response.CodeExportFileExpired, resp.Code)
}

func TestDownload_EmptyID(t *testing.T) {
	h := NewExportHandler(&stubSvc{})
	w := doJSON(h.Download, http.MethodGet, "/api/v1/export/download/", nil, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTask_EmptyID(t *testing.T) {
	h := NewExportHandler(&stubSvc{})
	w := doJSON(h.GetTask, http.MethodGet, "/api/v1/export/tasks/", nil, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
