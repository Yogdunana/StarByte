package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTrace_InvalidEntity(t *testing.T) {
	s := NewAuditService(&mockAuditRepo{}, nil)
	_, _, err := s.Trace(context.Background(), "nope", "abc", nil)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditEntityInvalid, appErr.Code)
}

func TestTrace_MapsDiff(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	id := uuid.New().String()
	log := sampleLog()
	log.EntityType = "user"
	log.EntityID = id
	log.BeforeJSON = `{"name":"旧","password":"secret123"}`
	log.AfterJSON = `{"name":"新","password":"secret123"}`
	r.On("ListByEntity", mock.Anything, "user", id, 1, 20).Return([]model.AuditLog{log}, int64(1), nil)

	list, total, err := s.Trace(context.Background(), "user", id, &dto.TraceQueryRequest{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	assert.NotContains(t, list[0].BeforeJSON, "secret123")
	assert.Equal(t, "user", list[0].EntityType)
	require.NotEmpty(t, list[0].Diff)
}

func TestReport_JSON(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	r.On("Count", mock.Anything, mock.Anything).Return(int64(4), nil)
	r.On("GroupCount", mock.Anything, mock.Anything, "action").Return([]repo.CountRow{{Key: "DELETE", Count: 2}}, nil)
	r.On("GroupCount", mock.Anything, mock.Anything, "module").Return([]repo.CountRow{{Key: "system", Count: 4}}, nil)
	r.On("GroupCount", mock.Anything, mock.Anything, "username").Return([]repo.CountRow{{Key: "admin", Count: 4}}, nil)
	r.On("GroupCompliance", mock.Anything, mock.Anything).Return([]repo.CountRow{{Key: "delete", Count: 2}}, nil)

	rep, data, filename, err := s.Report(context.Background(), &dto.ReportRequest{Format: "json"})
	require.NoError(t, err)
	assert.Empty(t, data)
	assert.Empty(t, filename)
	assert.Equal(t, int64(4), rep.Total)
	assert.Equal(t, "delete", rep.ByCompliance[0].Key)
	assert.Contains(t, rep.Note, "频繁删除")
}

func TestReport_CSVUsesExportEngine(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	r.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	r.On("GroupCount", mock.Anything, mock.Anything, mock.Anything).Return([]repo.CountRow{{Key: "UPDATE", Count: 1}}, nil)
	r.On("GroupCompliance", mock.Anything, mock.Anything).Return([]repo.CountRow{}, nil)

	_, data, filename, err := s.Report(context.Background(), &dto.ReportRequest{Format: "csv"})
	require.NoError(t, err)
	assert.Equal(t, "audit_compliance_report.csv", filename)
	assert.Contains(t, string(data), "类别")
	assert.Contains(t, string(data), "UPDATE")
}

func TestReport_BadFormat(t *testing.T) {
	s := NewAuditService(&mockAuditRepo{}, nil)
	_, _, _, err := s.Report(context.Background(), &dto.ReportRequest{Format: "xml"})
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditExportErr, appErr.Code)
}

func TestListArchives(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	row := model.AuditLogArchive{ID: uuid.New(), ArchiveDate: "2026-09-06", RecordCount: 3, MinIOObject: "audit-logs/x.json.gz"}
	r.On("ListArchives", mock.Anything, 1, 20).Return([]model.AuditLogArchive{row}, int64(1), nil)
	list, total, err := s.ListArchives(context.Background(), &dto.ArchiveListRequest{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, row.MinIOObject, list[0].MinIOObject)
}

func TestPullArchive_FromStore(t *testing.T) {
	r := &mockAuditRepo{}
	mem := storage.NewMemory("b")
	svc := NewAuditServiceWithStore(r, nil, mem)
	id := uuid.New()
	log := sampleLog()
	log.EntityType = "role"
	payload, err := json.Marshal([]model.AuditLog{log})
	require.NoError(t, err)
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	require.NoError(t, mem.Upload(context.Background(), "audit-logs/a.json.gz", bytes.NewReader(buf.Bytes()), int64(buf.Len()), "application/gzip"))

	r.On("GetArchiveByID", mock.Anything, id).Return(&model.AuditLogArchive{
		ID: id, ArchiveDate: "2026-09-06", MinIOObject: "audit-logs/a.json.gz", RecordCount: 1,
		CreatedAt: time.Now(),
	}, nil)

	got, err := svc.PullArchive(context.Background(), id, &dto.ArchiveListRequest{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, got.List, 1)
	assert.Equal(t, "role", got.List[0].EntityType)
}

func TestPullArchive_NotFound(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	id := uuid.New()
	r.On("GetArchiveByID", mock.Anything, id).Return(nil, nil)
	_, err := s.PullArchive(context.Background(), id, nil)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditArchiveNotFound, appErr.Code)
}

func TestPullArchive_KeywordAndPlainJSON(t *testing.T) {
	r := &mockAuditRepo{}
	mem := storage.NewMemory("b")
	svc := NewAuditServiceWithStore(r, nil, mem)
	id := uuid.New()
	keep := sampleLog()
	keep.Path = "/api/v1/system/roles/x"
	drop := sampleLog()
	drop.Path = "/api/v1/users/y"
	payload, err := json.Marshal([]model.AuditLog{keep, drop})
	require.NoError(t, err)
	require.NoError(t, mem.Upload(context.Background(), "audit-logs/plain.json", bytes.NewReader(payload), int64(len(payload)), "application/json"))
	r.On("GetArchiveByID", mock.Anything, id).Return(&model.AuditLogArchive{
		ID: id, MinIOObject: "audit-logs/plain.json", CreatedAt: time.Now(),
	}, nil)
	got, err := svc.PullArchive(context.Background(), id, &dto.ArchiveListRequest{Page: 1, PageSize: 20, Keyword: "roles"})
	require.NoError(t, err)
	require.Len(t, got.List, 1)
	assert.Contains(t, got.List[0].Path, "roles")
}

func TestPullArchive_UnconfiguredStore(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	id := uuid.New()
	r.On("GetArchiveByID", mock.Anything, id).Return(&model.AuditLogArchive{
		ID: id, MinIOObject: "audit-logs/a.json.gz",
	}, nil)
	_, err := s.PullArchive(context.Background(), id, nil)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditArchiveFetch, appErr.Code)
}

func TestReport_PDFAndLimit(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	r.On("Count", mock.Anything, mock.Anything).Return(int64(12), nil)
	many := make([]repo.CountRow, 0, 12)
	for i := 0; i < 12; i++ {
		many = append(many, repo.CountRow{Key: "u" + string(rune('a'+i)), Count: int64(12 - i)})
	}
	r.On("GroupCount", mock.Anything, mock.Anything, "username").Return(many, nil)
	r.On("GroupCount", mock.Anything, mock.Anything, mock.Anything).Return([]repo.CountRow{{Key: "CREATE", Count: 1}}, nil)
	r.On("GroupCompliance", mock.Anything, mock.Anything).Return([]repo.CountRow{{Key: "", Count: 1}, {Key: "export", Count: 2}}, nil)
	rep, data, filename, err := s.Report(context.Background(), &dto.ReportRequest{Format: "pdf"})
	require.NoError(t, err)
	assert.Equal(t, "audit_compliance_report.pdf", filename)
	require.GreaterOrEqual(t, len(data), 4)
	assert.Equal(t, "%PDF", string(data[:4]))
	require.LessOrEqual(t, len(rep.TopOperators), 10)
}

func TestGetByID_IncludesDiff(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	log := sampleLog()
	log.BeforeJSON = `{"name":"a"}`
	log.AfterJSON = `{"name":"b"}`
	log.ComplianceFlags = "permission"
	r.On("GetByID", mock.Anything, log.ID).Return(&log, nil)
	resp, err := s.GetByID(context.Background(), log.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"permission"}, resp.ComplianceFlags)
	require.NotEmpty(t, resp.Diff)
}
