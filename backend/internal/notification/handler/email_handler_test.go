package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubEmailSvc struct {
	send  *dto.SendEmailResponse
	batch *dto.BatchSendEmailResponse
	logs  []*dto.EmailLogResponse
	err   error
}

func (s *stubEmailSvc) Send(context.Context, *dto.SendEmailRequest, uuid.UUID) (*dto.SendEmailResponse, error) {
	return s.send, s.err
}
func (s *stubEmailSvc) SendBatch(context.Context, *dto.BatchSendEmailRequest, uuid.UUID) (*dto.BatchSendEmailResponse, error) {
	return s.batch, s.err
}
func (s *stubEmailSvc) ListLogs(context.Context, *dto.ListEmailLogsRequest) ([]*dto.EmailLogResponse, int64, error) {
	return s.logs, int64(len(s.logs)), s.err
}

func TestSendEmailOK(t *testing.T) {
	h := NewEmailHandler(&stubEmailSvc{send: &dto.SendEmailResponse{MessageID: "m1", Status: "queued", RecipientCount: 1}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(dto.SendEmailRequest{To: []string{"a@b.c"}, Subject: "s", Body: "b"})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/notifications/email/send", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uuid.New().String())
	h.SendEmail(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSendEmailInvalidJSON(t *testing.T) {
	h := NewEmailHandler(&stubEmailSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/notifications/email/send", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	h.SendEmail(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterRoutesWithEmail(t *testing.T) {
	assert.NotPanics(t, func() {
		RegisterRoutes(gin.New().Group("/api/v1"), gin.New().Group("/api/v1"),
			NewNotificationHandler(nil, nil), NewTemplateHandler(nil), nil,
			NewEmailHandler(&stubEmailSvc{}), nil)
	})
}

func TestListEmailLogsOK(t *testing.T) {
	h := NewEmailHandler(&stubEmailSvc{logs: []*dto.EmailLogResponse{{ID: "1", Status: "sent"}}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/notifications/email/logs", nil)
	h.ListEmailLogs(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSendBatchOK(t *testing.T) {
	h := NewEmailHandler(&stubEmailSvc{batch: &dto.BatchSendEmailResponse{Total: 1, Queued: 1}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(dto.BatchSendEmailRequest{
		Recipients: []dto.BatchRecipient{{Email: "a@b.c"}}, Subject: "s", Body: "b",
	})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/notifications/email/batch", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uuid.New().String())
	h.SendEmailBatch(c)
	require.Equal(t, http.StatusOK, w.Code)
}
