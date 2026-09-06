package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/internal/export/model"
	"github.com/Yogdunana/StarByte/backend/internal/export/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserID = "u1"

func newTestSvc(t *testing.T) (ExportService, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewExportService(repo.NewRedisRepo(rdb), nil, nil), mr
}

func sampleReq(rows int) *dto.TableExportRequest {
	data := make([][]string, rows)
	for i := 0; i < rows; i++ {
		data[i] = []string{"张三", "技术部"}
	}
	return &dto.TableExportRequest{
		Filename: "members",
		Title:    "会员名单",
		Columns:  []string{"姓名", "部门"},
		Rows:     data,
		Sheet:    "Sheet1",
	}
}

func TestCSV_BOMAndDelimiter(t *testing.T) {
	req := sampleReq(1)
	req.Delimiter = ";"
	raw, err := buildCSV(req)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(raw), 4)
	assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, raw[:3])
	body := string(raw[3:])
	assert.Contains(t, body, "姓名;部门")
	assert.Contains(t, body, "张三;技术部")
}

func TestExcel_ZipMagic(t *testing.T) {
	raw, err := buildExcel(sampleReq(2))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(raw), 4)
	assert.Equal(t, []byte{'P', 'K', 0x03, 0x04}, raw[:4])
}

func TestJSON_Roundtrip(t *testing.T) {
	req := sampleReq(1)
	raw, err := buildJSON(req)
	require.NoError(t, err)
	var got struct {
		Columns []string   `json:"columns"`
		Rows    [][]string `json:"rows"`
	}
	require.NoError(t, json.Unmarshal(raw, &got))
	assert.Equal(t, req.Columns, got.Columns)
	assert.Equal(t, req.Rows, got.Rows)
}

func TestPDF_StartsWithPercentPDF(t *testing.T) {
	raw, err := buildTablePDF(sampleReq(1))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(raw), 4)
	assert.Equal(t, "%PDF", string(raw[:4]))
}

func TestExport_EmptyRows(t *testing.T) {
	svc, _ := newTestSvc(t)
	_, err := svc.ExportTable(context.Background(), "csv", "", &dto.TableExportRequest{
		Columns: []string{"a"}, Rows: nil,
	})
	require.Error(t, err)
	assert.Equal(t, response.CodeExportEmptyData, err.(*response.AppError).Code)
}

func TestExport_InvalidFormat(t *testing.T) {
	svc, _ := newTestSvc(t)
	_, err := svc.ExportTable(context.Background(), "xml", "", sampleReq(1))
	require.Error(t, err)
	assert.Equal(t, response.CodeExportInvalidFormat, err.(*response.AppError).Code)
}

func TestExport_TemplateMissing(t *testing.T) {
	svc, _ := newTestSvc(t)
	_, err := svc.ExportTemplate(context.Background(), "no_such_tpl", "", &dto.TemplateExportRequest{})
	require.Error(t, err)
	assert.Equal(t, response.CodeExportTplNotFound, err.(*response.AppError).Code)
}

func TestExport_TaskNotFound(t *testing.T) {
	svc, _ := newTestSvc(t)
	_, err := svc.GetTask(context.Background(), "missing", testUserID, false)
	require.Error(t, err)
	assert.Equal(t, response.CodeExportTaskNotFound, err.(*response.AppError).Code)
}

func TestExport_SyncCSVAndDownload(t *testing.T) {
	svc, _ := newTestSvc(t)
	out, err := svc.ExportTable(context.Background(), "csv", testUserID, sampleReq(1))
	require.NoError(t, err)
	assert.Equal(t, model.StatusDone, out.Status)
	assert.NotEmpty(t, out.TaskID)
	assert.NotEmpty(t, out.FileID)
	dl, err := svc.Download(context.Background(), out.FileID, testUserID, false, true)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(dl.Bytes), 3)
	assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, dl.Bytes[:3])
}

func TestExport_AllFormats(t *testing.T) {
	svc, _ := newTestSvc(t)
	ctx := context.Background()
	for _, format := range []string{"excel", "csv", "pdf", "json"} {
		out, err := svc.ExportTable(ctx, format, testUserID, sampleReq(1))
		require.NoError(t, err, format)
		assert.Equal(t, model.StatusDone, out.Status, format)
		dl, err := svc.Download(ctx, out.FileID, testUserID, false, true)
		require.NoError(t, err, format)
		require.NotEmpty(t, dl.Bytes, format)
		switch format {
		case "excel":
			assert.Equal(t, []byte{'P', 'K', 0x03, 0x04}, dl.Bytes[:4])
		case "pdf":
			assert.Equal(t, "%PDF", string(dl.Bytes[:4]))
		case "json":
			assert.True(t, json.Valid(dl.Bytes))
		}
	}
}

func TestExport_BuiltinTemplates(t *testing.T) {
	svc, _ := newTestSvc(t)
	list := svc.ListTemplates()
	require.Len(t, list, 3)
	ctx := context.Background()
	for _, tpl := range list {
		out, err := svc.ExportTemplate(ctx, tpl.ID, testUserID, &dto.TemplateExportRequest{
			Filename:  tpl.ID,
			Watermark: "StarByte",
			Vars: map[string]string{
				"Title":    tpl.Name,
				"Date":     "2026-09-06",
				"RealName": "张三",
			},
		})
		require.NoError(t, err, tpl.ID)
		dl, err := svc.Download(ctx, out.FileID, testUserID, false, true)
		require.NoError(t, err, tpl.ID)
		assert.Equal(t, "%PDF", string(dl.Bytes[:4]), tpl.ID)
	}
}

func TestExport_ExpiredDownload(t *testing.T) {
	svc, mr := newTestSvc(t)
	out, err := svc.ExportTable(context.Background(), "json", testUserID, sampleReq(1))
	require.NoError(t, err)
	key := model.FileKey(out.FileID)
	raw, err := mr.Get(key)
	require.NoError(t, err)
	var meta model.FileMeta
	require.NoError(t, json.Unmarshal([]byte(raw), &meta))
	meta.ExpiresAt = time.Now().Add(-time.Minute)
	updated, err := json.Marshal(meta)
	require.NoError(t, err)
	require.NoError(t, mr.Set(key, string(updated)))
	_, err = svc.Download(context.Background(), out.FileID, testUserID, false, true)
	require.Error(t, err)
	assert.Equal(t, response.CodeExportFileExpired, err.(*response.AppError).Code)
}

func TestExport_ExpiredMissingFile(t *testing.T) {
	svc, _ := newTestSvc(t)
	_, err := svc.Download(context.Background(), "gone", testUserID, false, true)
	require.Error(t, err)
	assert.Equal(t, response.CodeExportFileExpired, err.(*response.AppError).Code)
}

func TestExport_AsyncPath(t *testing.T) {
	old := asyncRowThreshold
	asyncRowThreshold = 2
	defer func() { asyncRowThreshold = old }()

	svc, _ := newTestSvc(t)
	out, err := svc.ExportTable(context.Background(), "csv", testUserID, sampleReq(3))
	require.NoError(t, err)
	assert.Equal(t, model.StatusPending, out.Status)
	assert.NotEmpty(t, out.TaskID)
	assert.Empty(t, out.FileID)

	require.Eventually(t, func() bool {
		got, err := svc.GetTask(context.Background(), out.TaskID, testUserID, false)
		return err == nil && got.Status == model.StatusDone && got.FileID != ""
	}, 5*time.Second, 50*time.Millisecond)
}

func TestExport_TooManyRows(t *testing.T) {
	old := maxExportRows
	maxExportRows = 1
	defer func() { maxExportRows = old }()
	svc, _ := newTestSvc(t)
	_, err := svc.ExportTable(context.Background(), "csv", "", sampleReq(2))
	require.Error(t, err)
	assert.Equal(t, response.CodeExportTooManyRows, err.(*response.AppError).Code)
}

func TestFilenameAndSheetHelpers(t *testing.T) {
	assert.Equal(t, "export.csv", filenameOf("", "csv"))
	assert.Equal(t, "a_b.xlsx", filenameOf("a/b", "excel"))
	assert.Equal(t, "Sheet1", sanitizeSheet(""))
	assert.Equal(t, "ok", sanitizeSheet("ok"))
}

func TestHtmlToText(t *testing.T) {
	got := stripTags("<h1>标题</h1><p>hello<br/>world</p>")
	assert.Contains(t, got, "标题")
	assert.Contains(t, got, "hello")
}
