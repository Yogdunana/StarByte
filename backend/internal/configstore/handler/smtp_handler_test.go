package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSMTPSvc struct {
	smtp *dto.SMTPSettingsResponse
	err  error
}

func (s *stubSMTPSvc) List(context.Context, dto.ListQuery) ([]dto.ConfigResponse, error) {
	return nil, nil
}
func (s *stubSMTPSvc) GetByKey(context.Context, string) (*dto.ConfigResponse, error) { return nil, nil }
func (s *stubSMTPSvc) Create(context.Context, uuid.UUID, *dto.CreateConfigRequest) (*dto.ConfigResponse, error) {
	return nil, nil
}
func (s *stubSMTPSvc) Update(context.Context, uuid.UUID, uuid.UUID, *dto.UpdateConfigRequest) (*dto.ConfigResponse, error) {
	return nil, nil
}
func (s *stubSMTPSvc) Delete(context.Context, uuid.UUID) error { return nil }
func (s *stubSMTPSvc) GetSMTP(context.Context) (*dto.SMTPSettingsResponse, error) {
	return s.smtp, s.err
}
func (s *stubSMTPSvc) UpdateSMTP(context.Context, uuid.UUID, *dto.UpdateSMTPRequest) (*dto.SMTPSettingsResponse, error) {
	return s.smtp, s.err
}
func (s *stubSMTPSvc) TestSMTP(context.Context, *dto.TestSMTPRequest) (*dto.TestSMTPResponse, error) {
	return &dto.TestSMTPResponse{Sent: true, To: "a@b.c"}, s.err
}

func TestGetSMTPNeverIncludesPasswordField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewConfigHandler(&stubSMTPSvc{smtp: &dto.SMTPSettingsResponse{
		SMTPSettings:       dto.SMTPSettings{Host: "smtp.exmail.qq.com", Port: 465, From: "a@b.c"},
		PasswordConfigured: true,
		PasswordSource:     "STARBYTE_SMTP_PASSWORD",
	}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/smtp", nil)
	h.GetSMTP(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), `"password"`)
	assert.Contains(t, w.Body.String(), `"password_configured":true`)
}

func TestUpdateSMTPBadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewConfigHandler(&stubSMTPSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/system/smtp", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uuid.New().String())
	h.UpdateSMTP(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTestSMTPOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewConfigHandler(&stubSMTPSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(dto.TestSMTPRequest{To: "ops@example.test"})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/system/smtp/test", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.TestSMTP(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
