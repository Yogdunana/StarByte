package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
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
	require.Equal(t, "minister", engineNodeRole("minister"))
	require.Equal(t, "president", engineNodeRole("president"))
	require.Empty(t, engineNodeRole("start"))
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
	minister := &model.AdmissionActor{ID: uuid.New(), DepartmentID: &dept, Roles: []string{"minister"}}
	president := &model.AdmissionActor{ID: uuid.New(), Roles: []string{"president"}}
	other := &model.AdmissionActor{ID: uuid.New(), DepartmentID: &center, Roles: []string{"minister"}}

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
