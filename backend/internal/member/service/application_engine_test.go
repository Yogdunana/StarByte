package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

func TestApplicationVariables(t *testing.T) {
	dept := uuid.New()
	app := &model.MemberApplication{
		ID: uuid.New(), UserID: uuid.New(), Type: model.ApplicantOfficer,
		RealName: "王五", StudentNo: "2024001", DepartmentID: &dept, AdmissionRevision: 2,
	}
	vars := applicationVariables(app)
	require.Equal(t, app.ID.String(), vars["application_id"])
	require.Equal(t, app.UserID.String(), vars["applicant"])
	require.Equal(t, app.Type, vars["apply_type"])
	require.Equal(t, dept.String(), vars["department_id"])
	require.Equal(t, dept.String(), vars["department"])
	require.Nil(t, vars[engine.SkipMinisterVariable])
	require.Nil(t, vars[engine.SkipOfficerVariable])

	member := applicationVariables(&model.MemberApplication{
		ID: uuid.New(), UserID: uuid.New(), Type: model.ApplicantMember, RealName: "会员", StudentNo: "2024002",
	})
	require.Equal(t, true, member[engine.SkipMinisterVariable])
	require.Equal(t, true, member[engine.SkipOfficerVariable])
	_, hasDept := member["department_id"]
	require.False(t, hasDept)
}

func TestApplyEngineOutcome(t *testing.T) {
	now := time.Now()
	t.Run("minister approve waits for president", func(t *testing.T) {
		app := &model.MemberApplication{Status: model.AppPending, AdmissionStage: model.AdmissionMaterials}
		applyEngineOutcome(app, actionApprove, "president", false, now)
		require.Equal(t, model.AppReviewing, app.Status)
		require.Equal(t, engineStagePresident, app.CurrentStage)
		require.Equal(t, model.AdmissionEngine, app.AdmissionStage)
	})
	t.Run("president approve admits", func(t *testing.T) {
		app := &model.MemberApplication{Status: model.AppReviewing}
		applyEngineOutcome(app, actionApprove, "", true, now)
		require.Equal(t, model.AppApproved, app.Status)
		require.Equal(t, model.AdmissionApproved, app.AdmissionStage)
		require.Equal(t, stageLabel(model.AppApproved), app.CurrentStage)
	})
	t.Run("reject any node", func(t *testing.T) {
		app := &model.MemberApplication{Status: model.AppReviewing}
		applyEngineOutcome(app, actionReject, "president", false, now)
		require.Equal(t, model.AppRejected, app.Status)
		require.Equal(t, model.AdmissionRejected, app.AdmissionStage)
	})
	t.Run("supplement terminates chain", func(t *testing.T) {
		app := &model.MemberApplication{Status: model.AppPending}
		applyEngineOutcome(app, actionSupplement, "minister", false, now)
		require.Equal(t, model.AppSupplement, app.Status)
	})
}

func TestEngineNodeRole(t *testing.T) {
	require.Equal(t, "officer", engineNodeRole("officer"))
	require.Equal(t, "minister", engineNodeRole("minister"))
	require.Equal(t, "president", engineNodeRole("president"))
	require.Empty(t, engineNodeRole("start"))
}

func TestEngineStageLabel(t *testing.T) {
	require.Equal(t, engineStageOfficer, engineStageLabel("officer"))
	require.Equal(t, engineStageMinister, engineStageLabel("minister"))
	require.Equal(t, engineStagePresident, engineStageLabel("president"))
	require.Equal(t, "custom", engineStageLabel("custom"))
}

func TestApplyEngineOutcomeOfficerApprove(t *testing.T) {
	now := time.Now()
	app := &model.MemberApplication{Status: model.AppPending}
	applyEngineOutcome(app, actionApprove, "officer", false, now)
	require.Equal(t, model.AppPending, app.Status)
	require.Equal(t, engineStageOfficer, app.CurrentStage)
	require.Equal(t, model.AdmissionEngine, app.AdmissionStage)
}

func TestEngineReviewClosed(t *testing.T) {
	require.False(t, engineReviewClosed(model.AppPending))
	require.True(t, engineReviewClosed(model.AppApproved))
	require.True(t, engineReviewClosed(model.AppRejected))
	require.True(t, engineReviewClosed(model.AppSupplement))
}

func TestMapTransferCandidate(t *testing.T) {
	id := uuid.New()
	got := mapTransferCandidate(wfmodel.ApproverOption{ID: id, Name: "李四", DepartmentName: "技术部"})
	require.Equal(t, id.String(), got.ID)
	require.Equal(t, "李四", got.Name)
	require.Equal(t, "技术部", got.DepartmentName)
}

func TestAdmissionEngineMethodsRequireFlow(t *testing.T) {
	s := &admissionService{}
	ctx := context.Background()
	require.Error(t, s.TransferReview(ctx, uuid.New(), uuid.New(), uuid.New(), ""))
	_, err := s.ApplicationProgress(ctx, uuid.New(), uuid.New())
	require.Error(t, err)
	_, err = s.TransferCandidates(ctx, uuid.New(), uuid.New(), "")
	require.Error(t, err)
}

func TestCanViewEngineProgress(t *testing.T) {
	s := &admissionService{}
	applicant := uuid.New()
	app := &model.MemberApplication{UserID: applicant}
	require.NoError(t, s.canViewEngineProgress(context.Background(), stubAdmissionView{actor: nil}, app, applicant))
	require.Error(t, s.canViewEngineProgress(context.Background(), stubAdmissionView{}, app, uuid.New()))
	require.NoError(t, s.canViewEngineProgress(context.Background(), stubAdmissionView{actor: &model.AdmissionActor{Roles: []string{"minister"}}}, app, uuid.New()))
}

func TestMemberServiceEngineWrappersRequireAdmission(t *testing.T) {
	svc := NewMemberService(&mockAppRepo{}, &mockProfRepo{}, nil)
	ctx := context.Background()
	_, err := svc.Transfer(ctx, uuid.New(), uuid.New(), uuid.New(), "")
	require.Error(t, err)
	_, err = svc.ApplicationProgress(ctx, uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	_, err = svc.TransferCandidates(ctx, uuid.New(), uuid.New(), "")
	require.Error(t, err)
}

func TestMembershipRole(t *testing.T) {
	require.Equal(t, "member", membershipRole(&model.MemberApplication{Type: model.ApplicantMember}))
	require.Equal(t, "officer", membershipRole(&model.MemberApplication{Type: model.ApplicantOfficer}))
}

func TestSkipMinisterNode(t *testing.T) {
	require.True(t, skipMinisterNode(&model.MemberApplication{Type: model.ApplicantMember}))
	dept := uuid.New()
	require.False(t, skipMinisterNode(&model.MemberApplication{Type: model.ApplicantOfficer, DepartmentID: &dept}))
}

func TestEngineReviewPermissionKeepsDelegation(t *testing.T) {
	dept, center := uuid.New(), uuid.New()
	now := time.Now()
	app := &model.MemberApplication{UserID: uuid.New(), DepartmentID: &dept, StageEnteredAt: now}
	officer := &model.AdmissionActor{ID: uuid.New(), DepartmentID: &dept, Roles: []string{"officer"}}
	minister := &model.AdmissionActor{ID: uuid.New(), DepartmentID: &dept, Roles: []string{"minister"}}
	president := &model.AdmissionActor{ID: uuid.New(), Roles: []string{"president"}}
	other := &model.AdmissionActor{ID: uuid.New(), DepartmentID: &center, Roles: []string{"minister"}}

	require.NoError(t, engineReviewPermission(officer, app, &center, "officer", "", now))
	require.Error(t, engineReviewPermission(officer, app, &center, "minister", "", now))
	require.NoError(t, engineReviewPermission(minister, app, &center, "minister", "", now))
	require.Error(t, engineReviewPermission(president, app, &center, "minister", "代签", now))
	require.Error(t, engineReviewPermission(other, app, &center, "minister", "", now))
	require.NoError(t, engineReviewPermission(president, app, &center, "president", "", now))

	later := now.Add(25 * time.Hour)
	require.NoError(t, engineReviewPermission(president, app, &center, "minister", "超时代签", later))
	require.Error(t, engineReviewPermission(president, app, &center, "minister", "", later))
}

func TestEngineChainUsesDefaultDefinition(t *testing.T) {
	require.Equal(t, "member_application", engine.MemberApplicationDefinitionKey)
	graph, err := engine.ParseGraph(engine.MemberApplicationBPMN())
	require.NoError(t, err)
	require.NoError(t, engine.ValidateBusinessDefinition(engine.MemberApplicationDefinitionKey, graph))
}

type stubAdmissionView struct {
	actor *model.AdmissionActor
}

func (s stubAdmissionView) Objections(context.Context, uuid.UUID) ([]model.AdmissionObjection, error) {
	return nil, nil
}
func (s stubAdmissionView) OpenObjection(context.Context, uuid.UUID) (*model.AdmissionObjection, error) {
	return nil, nil
}
func (s stubAdmissionView) LockApplication(context.Context, uuid.UUID) (*model.MemberApplication, error) {
	return nil, nil
}
func (s stubAdmissionView) Actor(context.Context, uuid.UUID) (*model.AdmissionActor, error) {
	return s.actor, nil
}
func (s stubAdmissionView) ParentDepartment(context.Context, uuid.UUID) (*uuid.UUID, error) {
	return nil, nil
}
func (s stubAdmissionView) Signatures(context.Context, uuid.UUID) ([]model.AdmissionSignature, error) {
	return nil, nil
}
func (s stubAdmissionView) AddSignature(context.Context, *model.AdmissionSignature) error { return nil }
func (s stubAdmissionView) InterviewCompleted(context.Context, uuid.UUID, int16, time.Time) (bool, error) {
	return false, nil
}
func (s stubAdmissionView) SaveApplication(context.Context, *model.MemberApplication) error {
	return nil
}
func (s stubAdmissionView) WorkflowDefinitionKey(context.Context, uuid.UUID) (string, error) {
	return "", nil
}
