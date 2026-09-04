package service

import (
	"errors"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/file/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireAppError(t *testing.T, err error, code int, msgPart string) {
	t.Helper()
	require.Error(t, err)
	var appErr *response.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, code, appErr.Code)
	assert.Contains(t, appErr.Message, msgPart)
}

func TestValidateUpload_TypeRejected(t *testing.T) {
	_, err := validateUpload("malware.exe", 1024, "")
	requireAppError(t, err, response.CodeBadRequest, "文件类型不允许: .exe")
}

func TestValidateUpload_SizeRejected(t *testing.T) {
	_, err := validateUpload("doc.pdf", 60<<20, "")
	requireAppError(t, err, response.CodeBadRequest, "文件大小超过限制: 60MB > 50MB")
}

func TestValidateUpload_CategoryMismatch(t *testing.T) {
	_, err := validateUpload("photo.jpg", 1024, model.CategoryDocument)
	requireAppError(t, err, response.CodeBadRequest, "文件类型不允许: .jpg")
}

func TestValidateUpload_OK(t *testing.T) {
	got, err := validateUpload("photo.PNG", 1024, "")
	require.NoError(t, err)
	assert.Equal(t, model.CategoryImage, got.Category)
	assert.Equal(t, ".png", got.Ext)
	assert.Equal(t, "image/png", got.MimeType)
}
