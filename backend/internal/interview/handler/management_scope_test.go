package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

func TestCreateSessionRejectsMissingOrWrongDepartmentScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scope := range []*rbacModel.DataScopeCondition{nil, {Query: "1 = 0"}, {Query: "1 = 0", IsSelf: true}, {Query: "department_id = ?", Args: []interface{}{uuid.New()}}} {
		svc := &mockSvc{}
		handler := NewInterviewHandler(svc)
		router := gin.New()
		router.POST("/sessions", withUser(uuid.New()), func(c *gin.Context) { c.Set("data_scope_condition", scope); c.Next() }, handler.CreateSession)
		now := time.Now()
		body, err := json.Marshal(dto.CreateSessionRequest{Title: "一面", DepartmentID: uuid.NewString(), Round: 1, MaxCandidates: 10, StartTime: now, EndTime: now.Add(time.Hour)})
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		result := httptest.NewRecorder()
		router.ServeHTTP(result, request)
		require.Equal(t, http.StatusForbidden, result.Code)
		require.Empty(t, svc.Calls, "authorization must happen before any write")
	}
}
