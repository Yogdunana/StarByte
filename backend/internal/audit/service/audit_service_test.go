package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/audit"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuditRepo struct {
	mock.Mock
}

func (m *mockAuditRepo) Create(ctx context.Context, entry *model.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}
func (m *mockAuditRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}
func (m *mockAuditRepo) List(ctx context.Context, req *repo.ListParams) ([]model.AuditLog, int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]model.AuditLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockAuditRepo) Count(ctx context.Context, req *repo.ListParams) (int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockAuditRepo) Iterate(ctx context.Context, req *repo.ListParams, batchSize int, fn func([]model.AuditLog) error) error {
	args := m.Called(ctx, req, batchSize, fn)
	if logs, ok := args.Get(0).([]model.AuditLog); ok && fn != nil && args.Error(1) == nil {
		return fn(logs)
	}
	return args.Error(1)
}
func (m *mockAuditRepo) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockAuditRepo) CreateArchive(ctx context.Context, archive *model.AuditLogArchive) error {
	return m.Called(ctx, archive).Error(0)
}

func sampleLog() model.AuditLog {
	uid := uuid.New()
	return model.AuditLog{
		ID:             uuid.New(),
		UserID:         &uid,
		Username:       "admin",
		RealName:       "管理员",
		Method:         "POST",
		Path:           "/api/v1/system/roles",
		Module:         "system",
		Action:         model.ActionCreate,
		IP:             "192.168.1.1",
		RequestParams:  `{"password":"secret123","name":"部长"}`,
		ResponseStatus: 200,
		DurationMs:     45,
		CreatedAt:      time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC),
	}
}

func TestGetByID_NotFound(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	id := uuid.New()
	r.On("GetByID", mock.Anything, id).Return(nil, nil)
	_, err := s.GetByID(context.Background(), id)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditNotFound, appErr.Code)
}

func TestGetByID_Desensitizes(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	log := sampleLog()
	r.On("GetByID", mock.Anything, log.ID).Return(&log, nil)
	resp, err := s.GetByID(context.Background(), log.ID)
	assert.NoError(t, err)
	assert.Equal(t, "admin", resp.User.Username)
	assert.Equal(t, "管理员", resp.User.RealName)
	assert.NotContains(t, resp.RequestBody, "secret123")
	assert.Contains(t, resp.RequestBody, "***")
}

func TestQuery_MapsFields(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	log := sampleLog()
	r.On("List", mock.Anything, mock.Anything).Return([]model.AuditLog{log}, int64(1), nil)
	list, total, err := s.Query(context.Background(), &dto.ListAuditLogRequest{Page: 1, PageSize: 20, Action: "CREATE"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "CREATE", list[0].Action)
	assert.Equal(t, "system", list[0].Module)
	assert.Equal(t, "192.168.1.1", list[0].IPAddress)
}

func TestExport_LimitExceeded(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	r.On("Count", mock.Anything, mock.Anything).Return(int64(model.MaxExportRows+1), nil)
	_, _, err := s.Export(context.Background(), &dto.ExportAuditLogRequest{Format: "csv"})
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditExportLimit, appErr.Code)
}

func TestExport_UnsupportedFormat(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	r.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, _, err := s.Export(context.Background(), &dto.ExportAuditLogRequest{Format: "pdf"})
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditExportErr, appErr.Code)
}

func TestExport_CSV(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	log := sampleLog()
	r.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	r.On("Iterate", mock.Anything, mock.Anything, model.DefaultIterateBatch, mock.Anything).Return([]model.AuditLog{log}, nil)
	data, filename, err := s.Export(context.Background(), &dto.ExportAuditLogRequest{Format: "csv"})
	assert.NoError(t, err)
	assert.Equal(t, "audit_logs.csv", filename)
	assert.Contains(t, string(data), "admin")
	assert.Contains(t, string(data), "CREATE")
	assert.Contains(t, string(data), "时间")
}

func TestExport_Excel(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	log := sampleLog()
	r.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	r.On("Iterate", mock.Anything, mock.Anything, model.DefaultIterateBatch, mock.Anything).Return([]model.AuditLog{log}, nil)
	data, filename, err := s.Export(context.Background(), &dto.ExportAuditLogRequest{Format: "excel"})
	assert.NoError(t, err)
	assert.Equal(t, "audit_logs.xlsx", filename)
	assert.Greater(t, len(data), 100)
}

func TestArchive_MinIOFailDoesNotDelete(t *testing.T) {
	r := &mockAuditRepo{}
	svc := &auditService{
		auditRepo: r,
		minioCfg:  &config.MinIOConfig{Endpoint: "localhost:9000", Bucket: "b"},
		uploadFn: func(_ *config.MinIOConfig, _ string, _ []byte, _ string) error {
			return errors.New("boom")
		},
	}
	log := sampleLog()
	r.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	r.On("Iterate", mock.Anything, mock.Anything, model.DefaultIterateBatch, mock.Anything).Return([]model.AuditLog{log}, nil)

	_, err := svc.Archive(context.Background(), 90)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditArchiveErr, appErr.Code)
	r.AssertNotCalled(t, "DeleteBefore", mock.Anything, mock.Anything)
}

func TestArchive_SuccessDeletesAfterUpload(t *testing.T) {
	r := &mockAuditRepo{}
	var uploadedName string
	svc := &auditService{
		auditRepo: r,
		minioCfg:  &config.MinIOConfig{Endpoint: "localhost:9000", Bucket: "b"},
		uploadFn: func(_ *config.MinIOConfig, object string, data []byte, _ string) error {
			uploadedName = object
			assert.NotEmpty(t, data)
			return nil
		},
	}
	log := sampleLog()
	r.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	r.On("Iterate", mock.Anything, mock.Anything, model.DefaultIterateBatch, mock.Anything).Return([]model.AuditLog{log}, nil)
	r.On("DeleteBefore", mock.Anything, mock.Anything).Return(int64(1), nil)
	r.On("CreateArchive", mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.Archive(context.Background(), 90)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.RecordCount)
	assert.Contains(t, uploadedName, "audit-logs/")
	assert.Contains(t, uploadedName, ".json.gz")
	r.AssertCalled(t, "DeleteBefore", mock.Anything, mock.Anything)
}

func TestArchive_UnconfiguredMinIO(t *testing.T) {
	r := &mockAuditRepo{}
	svc := &auditService{auditRepo: r, minioCfg: nil}
	log := sampleLog()
	r.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	r.On("Iterate", mock.Anything, mock.Anything, model.DefaultIterateBatch, mock.Anything).Return([]model.AuditLog{log}, nil)
	_, err := svc.Archive(context.Background(), 90)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeAuditArchiveErr, appErr.Code)
}

func TestLogAndEvents(t *testing.T) {
	r := &mockAuditRepo{}
	s := NewAuditService(r, nil)
	r.On("Create", mock.Anything, mock.Anything).Return(nil)
	err := s.Log(context.Background(), &model.AuditEntry{
		UserID: uuid.New(), Username: "u", Action: model.ActionLogin, Path: "/api/v1/auth/login",
	})
	assert.NoError(t, err)

	bus := events.NewEventBus()
	RegisterAuthEvents(bus, s)
	bus.Publish(context.Background(), events.UserLoginEvent{Username: "u", IP: "1.1.1.1"})
	time.Sleep(20 * time.Millisecond)
}

func TestDesensitizeHelpers(t *testing.T) {
	assert.Equal(t, "138****5678", audit.MaskPhone("13812345678"))
	assert.Equal(t, "z***@example.com", audit.MaskEmail("zhang@example.com"))
}

func TestSha256Hex(t *testing.T) {
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", sha256Hex([]byte("")))
}

func TestModelTableName(t *testing.T) {
	assert.Equal(t, "audit_logs", model.AuditLog{}.TableName())
	assert.Equal(t, "audit_log_archives", model.AuditLogArchive{}.TableName())
}

func TestArchiveObjectPath(t *testing.T) {
	p := archiveObjectPath(time.Date(2026, 9, 4, 2, 0, 0, 0, time.UTC))
	assert.Equal(t, "audit-logs/2026/09/audit_logs_20260904.json.gz", p)
}
