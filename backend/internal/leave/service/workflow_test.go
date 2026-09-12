package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type stubFlow struct {
	nodes        []string
	idx          int
	rejected     bool
	failStart    error
	failComplete error
	started      int
}

func (s *stubFlow) Start(context.Context, string, string, string, uuid.UUID, map[string]interface{}) (*wfmodel.FlowInstance, error) {
	if s.failStart != nil {
		return nil, s.failStart
	}
	s.started++
	s.idx = 0
	s.rejected = false
	return &wfmodel.FlowInstance{ID: uuid.New(), Status: 0}, nil
}

func (s *stubFlow) CompleteLeaveApproval(_ context.Context, _ uuid.UUID, _ uuid.UUID, action engine.TaskAction, _ string) error {
	if s.failComplete != nil {
		return s.failComplete
	}
	if action == engine.ActionReject {
		s.rejected = true
		return nil
	}
	s.idx++
	return nil
}

func (s *stubFlow) RunningApprovalNode(context.Context, uuid.UUID) (string, bool, error) {
	if s.rejected || s.idx >= len(s.nodes) {
		return "", true, nil
	}
	return s.nodes[s.idx], false, nil
}

type stubEngine struct{ flow *stubFlow }

func (s *stubEngine) BindTransaction(*gorm.DB) (LeaveFlow, func(context.Context), error) {
	return s.flow, func(context.Context) {}, nil
}

func workflowFixture(t *testing.T) (*leaveService, *memRepo, *stubFlow, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	svc, mem, applicant, approver, annualID := fixture()
	flow := &stubFlow{nodes: []string{model.StageDepartment, model.StageOrg}}
	svc.engine = &stubEngine{flow: flow}
	return svc, mem, flow, applicant, approver, annualID
}

func TestWorkflowSubmitApproveTwoLevels(t *testing.T) {
	svc, _, flow, applicant, approver, annualID := workflowFixture(t)
	ctx := context.Background()
	start := monday()
	app, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, start.Add(8*time.Hour)))
	require.NoError(t, err)
	assert.Equal(t, 1, flow.started)
	assert.Equal(t, model.ApprovalStatusPending, app.Status)
	assert.Equal(t, model.StageDepartment, app.WorkflowStage)
	require.NotEmpty(t, app.WorkflowInstanceID)

	id := uuid.MustParse(app.ID)
	require.NoError(t, svc.Approve(ctx, approverViewer(approver), id, "dept ok"))
	got, err := svc.Get(ctx, applicantViewer(applicant), id)
	require.NoError(t, err)
	assert.Equal(t, model.ApprovalStatusPending, got.Status)
	assert.Equal(t, model.StageOrg, got.WorkflowStage)

	require.NoError(t, svc.Approve(ctx, approverViewer(approver), id, "org ok"))
	got, err = svc.Get(ctx, applicantViewer(applicant), id)
	require.NoError(t, err)
	assert.Equal(t, model.ApprovalStatusApproved, got.Status)
	assert.Empty(t, got.WorkflowStage)
	bals, err := svc.Balances(ctx, applicantViewer(applicant), "", 2026)
	require.NoError(t, err)
	assert.Equal(t, 4.0, findAnnual(bals, annualID).RemainingDays)
}

func TestWorkflowRejectRestoresBalance(t *testing.T) {
	svc, _, _, applicant, approver, annualID := workflowFixture(t)
	ctx := context.Background()
	start := monday()
	app, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, start.Add(8*time.Hour)))
	require.NoError(t, err)
	require.NoError(t, svc.Reject(ctx, approverViewer(approver), uuid.MustParse(app.ID), "no"))
	got, err := svc.Get(ctx, applicantViewer(applicant), uuid.MustParse(app.ID))
	require.NoError(t, err)
	assert.Equal(t, model.ApprovalStatusRejected, got.Status)
	bals, err := svc.Balances(ctx, applicantViewer(applicant), "", 2026)
	require.NoError(t, err)
	assert.Equal(t, 5.0, findAnnual(bals, annualID).RemainingDays)
}

func TestWorkflowStartFailureRollsBack(t *testing.T) {
	svc, mem, flow, applicant, _, annualID := workflowFixture(t)
	flow.failStart = workflowUnavailable()
	_, err := svc.Submit(context.Background(), applicantViewer(applicant), submitReq(annualID, monday(), monday().Add(8*time.Hour)))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveWorkflow, err.(*response.AppError).Code)
	assert.Empty(t, mem.apps)
}

func TestAdaptEngineNil(t *testing.T) {
	require.Nil(t, AdaptEngine(nil))
}

func TestListTypesAndNew(t *testing.T) {
	svc, mem, _, _, _ := fixture()
	rows, err := svc.ListTypes(context.Background())
	require.NoError(t, err)
	require.Len(t, rows, 2)
	got := New(mem, nil)
	require.NotNil(t, got)
}

func TestCalendarInvalidRange(t *testing.T) {
	svc, _, applicant, _, _ := fixture()
	_, err := svc.Calendar(context.Background(), applicantViewer(applicant), &dto.ListLeaveRequest{From: "bad", To: "2026-09-30"})
	require.Error(t, err)
	_, err = svc.Calendar(context.Background(), applicantViewer(applicant), &dto.ListLeaveRequest{From: "2026-09-30", To: "2026-09-01"})
	require.Error(t, err)
}

func TestParseAttachmentsAndTypeCode(t *testing.T) {
	_, err := parseAttachments([]dto.Attachment{{FileID: "x"}, {FileID: "y"}, {FileID: "z"}, {FileID: "a"}, {FileID: "b"}, {FileID: "c"}})
	require.Error(t, err)
	_, err = parseAttachments([]dto.Attachment{{FileID: "not-a-uuid"}})
	require.Error(t, err)
	svc, _, _, approver, _ := fixture()
	_, err = svc.CreateType(context.Background(), approverViewer(approver), &dto.UpsertLeaveTypeRequest{Name: "X", Code: "1bad"})
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveTypeInvalid, err.(*response.AppError).Code)
}

func TestUniqueExcept(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	got := uniqueExcept([]uuid.UUID{a, a, b, uuid.Nil}, a)
	require.Equal(t, []uuid.UUID{b}, got)
}

func TestLeaveApproverRejectsOtherBusiness(t *testing.T) {
	r := NewLeaveApprover(nil)
	_, err := r.Resolve(context.Background(), &wfmodel.FlowInstance{BusinessType: "member_application"}, &engine.FlowNode{Config: map[string]interface{}{}})
	require.Error(t, err)
	require.NotNil(t, r.ForTransaction(nil))
}
