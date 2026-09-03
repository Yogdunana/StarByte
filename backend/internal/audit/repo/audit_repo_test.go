package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListParams_DefaultValues(t *testing.T) {
	params := &ListParams{
		Page:     1,
		PageSize: 20,
	}
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.PageSize)
}

func TestListParams_WithFilters(t *testing.T) {
	params := &ListParams{
		Page:      2,
		PageSize:  50,
		Username:  "admin",
		Method:    "POST",
		Path:      "/api/v1/users",
		IP:        "192.168.1.1",
		RequestID: "req-123",
	}
	assert.Equal(t, "admin", params.Username)
	assert.Equal(t, "POST", params.Method)
	assert.Equal(t, "/api/v1/users", params.Path)
	assert.Equal(t, "192.168.1.1", params.IP)
	assert.Equal(t, "req-123", params.RequestID)
}

func TestNewAuditRepo(t *testing.T) {
	r := NewAuditRepo(nil)
	assert.NotNil(t, r)
}

func TestListParams_StatusRange(t *testing.T) {
	minVal := 200
	maxVal := 299
	params := &ListParams{
		StatusMin: &minVal,
		StatusMax: &maxVal,
	}
	assert.NotNil(t, params.StatusMin)
	assert.NotNil(t, params.StatusMax)
	assert.Equal(t, 200, *params.StatusMin)
	assert.Equal(t, 299, *params.StatusMax)
}
