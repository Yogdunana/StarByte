package repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAuditRepo(t *testing.T) {
	assert.NotNil(t, NewAuditRepo(nil))
}

func TestListParams_Filters(t *testing.T) {
	uid := uuid.New()
	now := time.Now()
	params := &ListParams{
		Page:      2,
		PageSize:  50,
		Username:  "admin",
		UserID:    &uid,
		Action:    "CREATE",
		Module:    "system",
		Keyword:   "roles",
		IP:        "192.168.1.1",
		StartTime: &now,
	}
	assert.Equal(t, "CREATE", params.Action)
	assert.Equal(t, "system", params.Module)
	assert.Equal(t, "roles", params.Keyword)
	assert.Equal(t, 2, params.Page)
	assert.Equal(t, 50, params.PageSize)
}
