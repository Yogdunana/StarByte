package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateInterview_SessionFull(t *testing.T) {
	sessions := &mockSessionRepo{}
	svc := NewInterviewService(sessions, &mockInterviewRepo{}, &mockEvalRepo{}, nil, nil)
	sid := uuid.New()
	sessions.On("GetByID", mock.Anything, sid).Return(&model.Session{ID: sid, Status: model.SessionPending, MaxCandidates: 1}, nil)
	sessions.On("CountCandidates", mock.Anything, sid).Return(int64(1), nil)
	_, err := svc.CreateInterview(context.Background(), uuid.New(), &dto.CreateInterviewRequest{
		SessionID: sid.String(), ApplicantID: uuid.New().String(),
	})
	requireAppError(t, err, response.CodeInterviewSessionFull)
}

func TestCreateInterview_FromApplication(t *testing.T) {
	sessions := &mockSessionRepo{}
	records := &mockInterviewRepo{}
	notify := &mockNotify{}
	svc := NewInterviewService(sessions, records, &mockEvalRepo{}, notify, nil)
	sid := uuid.New()
	appID := uuid.New()
	uid := uuid.New()
	sess := &model.Session{ID: sid, Status: model.SessionPending, MaxCandidates: 20, Round: 1, Location: "A", Title: "一面"}
	sessions.On("GetByID", mock.Anything, sid).Return(sess, nil)
	sessions.On("CountCandidates", mock.Anything, sid).Return(int64(0), nil)
	records.On("GetApplication", mock.Anything, appID).Return(&model.ApplicationBrief{ID: appID, UserID: uid, RealName: "张三"}, nil)
	records.On("FindBySessionApplicant", mock.Anything, sid, uid).Return(nil, nil)
	records.On("Create", mock.Anything, mock.AnythingOfType("*model.Interview")).Return(nil)
	records.On("GetUser", mock.Anything, uid).Return(&model.NamedUser{ID: uid, RealName: "张三"}, nil)
	notify.On("Send", mock.Anything, []uuid.UUID{uid}, tplInvite, mock.Anything).Return(nil)
	iv := &model.Interview{ID: uuid.New(), SessionID: &sid, ApplicantID: uid, Status: model.InterviewPending}
	records.On("GetByIDWithNames", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(namedInterview(iv, "张三"), nil)
	records.On("ListInterviewers", mock.Anything, mock.Anything).Return([]model.InterviewerNamed{}, nil)
	out, err := svc.CreateInterview(context.Background(), uuid.New(), &dto.CreateInterviewRequest{
		SessionID: sid.String(), ApplicationID: appID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, "张三", out.Applicant.Name)
}

func TestCheckin_WrongUser(t *testing.T) {
	records := &mockInterviewRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, ApplicantID: uuid.New(), Status: model.InterviewPending}, nil)
	_, err := svc.Checkin(context.Background(), uuid.New(), id, "")
	requireAppError(t, err, response.CodeForbidden)
}

func TestCheckin_OK(t *testing.T) {
	records := &mockInterviewRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	uid := uuid.New()
	iv := &model.Interview{ID: id, ApplicantID: uid, Status: model.InterviewPending}
	records.On("GetByID", mock.Anything, id).Return(iv, nil)
	records.On("Update", mock.Anything, mock.AnythingOfType("*model.Interview")).Return(nil)
	records.On("GetByIDWithNames", mock.Anything, id).Return(namedInterview(iv, "张三"), nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{}, nil)
	out, err := svc.Checkin(context.Background(), uid, id, "")
	require.NoError(t, err)
	require.Equal(t, model.InterviewCheckedIn, iv.Status)
	require.Equal(t, "张三", out.Applicant.Name)
}

func TestAssign_TimeConflict(t *testing.T) {
	records := &mockInterviewRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	evalID := uuid.New()
	now := time.Now()
	iv := &model.Interview{ID: id, Status: model.InterviewPending, ScheduledAt: &now, Duration: 30}
	records.On("GetByID", mock.Anything, id).Return(iv, nil)
	records.On("HasInterviewerConflict", mock.Anything, evalID, mock.Anything, mock.Anything, id).Return(true, nil)
	_, err := svc.AssignEvaluators(context.Background(), id, &dto.AssignEvaluatorsRequest{EvaluatorIDs: []string{evalID.String()}})
	requireAppError(t, err, response.CodeInterviewConflict)
}

func TestAssign_SendsNotification(t *testing.T) {
	records := &mockInterviewRepo{}
	notify := &mockNotify{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, notify, nil)
	id := uuid.New()
	evalID := uuid.New()
	iv := &model.Interview{ID: id, Status: model.InterviewPending, ApplicantID: uuid.New()}
	records.On("GetByID", mock.Anything, id).Return(iv, nil)
	records.On("ReplaceInterviewers", mock.Anything, id, mock.Anything).Return(nil)
	records.On("GetUser", mock.Anything, iv.ApplicantID).Return(&model.NamedUser{RealName: "张三"}, nil)
	notify.On("Send", mock.Anything, []uuid.UUID{evalID}, tplAssigned, mock.Anything).Return(nil)
	records.On("GetByIDWithNames", mock.Anything, id).Return(namedInterview(iv, "张三"), nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{{InterviewID: id, InterviewerID: evalID, Name: "官"}}, nil)
	out, err := svc.AssignEvaluators(context.Background(), id, &dto.AssignEvaluatorsRequest{EvaluatorIDs: []string{evalID.String()}})
	require.NoError(t, err)
	require.Len(t, out.Evaluators, 1)
}

func TestSubmitResult_SyncsApplication(t *testing.T) {
	records := &mockInterviewRepo{}
	evals := &mockEvalRepo{}
	notify := &mockNotify{}
	syncer := &mockSyncer{}
	svc := NewInterviewService(&mockSessionRepo{}, records, evals, notify, syncer)
	id := uuid.New()
	appID := uuid.New()
	op := uuid.New()
	iv := &model.Interview{ID: id, Status: model.InterviewDone, ApplicantID: uuid.New(), ApplicationID: &appID, ResultCode: 1}
	records.On("GetByID", mock.Anything, id).Return(iv, nil)
	records.On("GetByIDWithNames", mock.Anything, id).Return(namedInterview(iv, "张三"), nil)
	evals.On("ListByInterview", mock.Anything, id).Return([]model.EvaluationNamed{}, nil)
	evals.On("ListDimensions", mock.Anything).Return([]model.Dimension{}, nil)
	records.On("Update", mock.Anything, mock.AnythingOfType("*model.Interview")).Return(nil)
	records.On("GetUser", mock.Anything, iv.ApplicantID).Return(&model.NamedUser{RealName: "张三"}, nil)
	notify.On("Send", mock.Anything, mock.Anything, tplResult, mock.Anything).Return(nil)
	syncer.On("SyncFromInterview", mock.Anything, op, appID, int16(1), "过").Return(nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{}, nil)
	out, err := svc.SubmitResult(context.Background(), op, id, &dto.SubmitResultRequest{Result: 1, Comment: "过"})
	require.NoError(t, err)
	require.Equal(t, int16(1), out.Result)
}

func TestStartInterview_NoEvaluator(t *testing.T) {
	records := &mockInterviewRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, Status: model.InterviewCheckedIn}, nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{}, nil)
	_, err := svc.StartInterview(context.Background(), uuid.New(), id)
	requireAppError(t, err, response.CodeInterviewNoEvaluator)
}

func TestGetInterview_NotFound(t *testing.T) {
	records := &mockInterviewRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	records.On("GetByIDWithNames", mock.Anything, id).Return(nil, nil)
	_, err := svc.GetInterview(context.Background(), id, nil)
	requireAppError(t, err, response.CodeInterviewRecordGone)
}

func TestStats_PassRate(t *testing.T) {
	evals := &mockEvalRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, &mockInterviewRepo{}, evals, nil, nil)
	evals.On("Stats", mock.Anything, mock.Anything).Return(
		model.StatsRow{Total: 4, PassCount: 2, FailCount: 1, PendingCount: 1},
		[]model.ScoreBucket{{Range: "80-89", Count: 2}},
		[]model.DeptStat{{Department: "技术部", Count: 4, PassCount: 2}},
		nil,
	)
	out, err := svc.Stats(context.Background(), &dto.StatsQuery{})
	require.NoError(t, err)
	require.Equal(t, 50.0, out.PassRate)
}

func TestDeleteDimension_NotFound(t *testing.T) {
	evals := &mockEvalRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, &mockInterviewRepo{}, evals, nil, nil)
	id := uuid.New()
	evals.On("GetDimensionByID", mock.Anything, id).Return(nil, nil)
	err := svc.DeleteDimension(context.Background(), id)
	requireAppError(t, err, response.CodeInterviewDimGone)
}
