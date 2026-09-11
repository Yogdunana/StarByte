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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

func TestCanManageSessionRejectsAssignedInterviewerOutsideDepartment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownDept, otherDept := uuid.New(), uuid.New()
	sessionID, uid := uuid.New(), uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{ownDept}}

	svc := &mockSvc{}
	handler := NewInterviewHandler(svc)
	router := gin.New()
	router.DELETE("/sessions/:id", withUser(uid), func(c *gin.Context) { c.Set("data_scope_condition", scope); c.Next() }, handler.requireManagedSession, handler.DeleteSession)
	svc.On("GetSession", mock.Anything, sessionID, mock.Anything).
		Return(&dto.SessionResponse{ID: sessionID.String(), DepartmentID: otherDept.String()}, nil)

	request := httptest.NewRequest(http.MethodDelete, "/sessions/"+sessionID.String(), nil)
	result := httptest.NewRecorder()
	router.ServeHTTP(result, request)
	require.Equal(t, http.StatusForbidden, result.Code)
	svc.AssertNotCalled(t, "DeleteSession")
}

func TestCanManageSessionAllowsDepartmentManager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dept := uuid.New()
	sessionID, uid := uuid.New(), uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}

	svc := &mockSvc{}
	handler := NewInterviewHandler(svc)
	router := gin.New()
	router.DELETE("/sessions/:id", withUser(uid), func(c *gin.Context) { c.Set("data_scope_condition", scope); c.Next() }, handler.requireManagedSession, handler.DeleteSession)
	svc.On("GetSession", mock.Anything, sessionID, mock.Anything).
		Return(&dto.SessionResponse{ID: sessionID.String(), DepartmentID: dept.String()}, nil)
	svc.On("DeleteSession", mock.Anything, sessionID).Return(nil)

	request := httptest.NewRequest(http.MethodDelete, "/sessions/"+sessionID.String(), nil)
	result := httptest.NewRecorder()
	router.ServeHTTP(result, request)
	require.Equal(t, http.StatusOK, result.Code)
	svc.AssertCalled(t, "DeleteSession", mock.Anything, sessionID)
}

func TestCreateInterviewRejectsAssignedInterviewerOutsideDepartment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownDept, otherDept := uuid.New(), uuid.New()
	sessionID, uid := uuid.New(), uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{ownDept}}

	svc := &mockSvc{}
	handler := NewInterviewHandler(svc)
	router := gin.New()
	router.POST("/interviews", withUser(uid), func(c *gin.Context) { c.Set("data_scope_condition", scope); c.Next() }, handler.CreateInterview)
	svc.On("GetSession", mock.Anything, sessionID, mock.Anything).
		Return(&dto.SessionResponse{ID: sessionID.String(), DepartmentID: otherDept.String()}, nil)

	body, err := json.Marshal(dto.CreateInterviewRequest{SessionID: sessionID.String()})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/interviews", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()
	router.ServeHTTP(result, request)
	require.Equal(t, http.StatusForbidden, result.Code)
	svc.AssertNotCalled(t, "CreateInterview")
}

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
