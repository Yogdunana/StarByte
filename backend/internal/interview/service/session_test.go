package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func TestCreateSession_InvalidTime(t *testing.T) {
	svc := NewInterviewService(&mockSessionRepo{}, &mockInterviewRepo{}, &mockEvalRepo{}, nil, nil)
	now := time.Now()
	_, err := svc.CreateSession(context.Background(), uuid.New(), &dto.CreateSessionRequest{
		Title: "一面", Round: 1, StartTime: now, EndTime: now, MaxCandidates: 10,
	})
	requireAppError(t, err, response.CodeBadRequest)
}

func TestCreateSession_OK(t *testing.T) {
	sessions := &mockSessionRepo{}
	svc := NewInterviewService(sessions, &mockInterviewRepo{}, &mockEvalRepo{}, nil, nil)
	now := time.Now()
	sessions.On("Create", mock.Anything, mock.AnythingOfType("*model.Session")).Return(nil)
	sessions.On("GetByIDWithNames", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&model.SessionWithNames{
		Session: model.Session{Title: "一面", Status: model.SessionPending, MaxCandidates: 10},
	}, nil)
	out, err := svc.CreateSession(context.Background(), uuid.New(), &dto.CreateSessionRequest{
		Title: "一面", Round: 1, StartTime: now, EndTime: now.Add(time.Hour), MaxCandidates: 10, Location: "A101",
	})
	require.NoError(t, err)
	require.Equal(t, "一面", out.Title)
}

func TestStartSession_InvalidState(t *testing.T) {
	sessions := &mockSessionRepo{}
	svc := NewInterviewService(sessions, &mockInterviewRepo{}, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	sessions.On("GetByID", mock.Anything, id).Return(&model.Session{ID: id, Status: model.SessionEnded}, nil)
	_, err := svc.StartSession(context.Background(), id)
	requireAppError(t, err, response.CodeInterviewInvalidState)
}

func TestGetSession_AssignedInterviewer(t *testing.T) {
	sessions := &mockSessionRepo{}
	records := &mockInterviewRepo{}
	svc := NewInterviewService(sessions, records, &mockEvalRepo{}, nil, nil)
	owner, examiner, sid := uuid.New(), uuid.New(), uuid.New()
	self := &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}
	sessions.On("GetByIDWithNames", mock.Anything, sid).Return(&model.SessionWithNames{
		Session: model.Session{ID: sid, Title: "一面", CreatedBy: &owner},
	}, nil)
	records.On("IsAssignedToSession", mock.Anything, sid, examiner).Return(true, nil)
	out, err := svc.GetSession(context.Background(), Viewer{ID: examiner, Scope: self}, sid)
	require.NoError(t, err)
	require.Equal(t, "一面", out.Title)

	records.On("IsAssignedToSession", mock.Anything, sid, owner).Return(false, nil).Maybe()
	outsider := uuid.New()
	records.On("IsAssignedToSession", mock.Anything, sid, outsider).Return(false, nil)
	_, err = svc.GetSession(context.Background(), Viewer{ID: outsider, Scope: self}, sid)
	requireAppError(t, err, response.CodeForbidden)
}

func TestRewriteSessionScope_SelfIncludesAssigned(t *testing.T) {
	uid := uuid.New()
	got := rewriteSessionScope(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, uid)
	require.Contains(t, got.Query, "created_by")
	require.Contains(t, got.Query, "interview_interviewers")
	require.Equal(t, uid, got.Args[0])
	require.Equal(t, uid, got.Args[1])
}

func TestGetSession_NotFound(t *testing.T) {
	sessions := &mockSessionRepo{}
	svc := NewInterviewService(sessions, &mockInterviewRepo{}, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	sessions.On("GetByIDWithNames", mock.Anything, id).Return(nil, nil)
	_, err := svc.GetSession(context.Background(), Viewer{ID: uuid.New()}, id)
	requireAppError(t, err, response.CodeInterviewNotFound)
}

func TestEndSession_MarksAbsent(t *testing.T) {
	sessions := &mockSessionRepo{}
	records := &mockInterviewRepo{}
	svc := NewInterviewService(sessions, records, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	sess := &model.Session{ID: id, Status: model.SessionOngoing, Title: "一面"}
	sessions.On("GetByID", mock.Anything, id).Return(sess, nil)
	records.On("MarkAbsentBySession", mock.Anything, id).Return(nil)
	sessions.On("Update", mock.Anything, mock.AnythingOfType("*model.Session")).Return(nil)
	sessions.On("GetByIDWithNames", mock.Anything, id).Return(namedSession(&model.Session{ID: id, Status: model.SessionEnded, Title: "一面"}), nil)
	out, err := svc.EndSession(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, model.SessionEnded, out.Status)
}
