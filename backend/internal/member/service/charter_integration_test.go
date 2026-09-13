package service

import (
	"context"
	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	rb "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	um "github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine/nodes"
	wm "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	wr "github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCharterChainConcurrentOfficesCommitteeAndExplicitProbation(t *testing.T) {
	ctx := context.Background()
	tx := testutil.OpenPostgres(t).Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	// A transaction-local template and electorate make the test independent of seed data.
	require.NoError(t, tx.Exec("DELETE FROM user_roles WHERE role_id IN (SELECT id FROM roles WHERE code IN ('minister','president','vice_president','center_director'))").Error)
	raw, err := os.ReadFile("../../../migrations/000073_charter_appointments.up.sql")
	require.NoError(t, err)
	graph := strings.Split(string(raw), "$graph$")[1]
	var def wm.FlowDefinition
	err = tx.Where("key=?", engine.MemberApplicationDefinitionKey).First(&def).Error
	if err != nil {
		def = wm.FlowDefinition{ID: uuid.New(), Key: engine.MemberApplicationDefinitionKey, Name: "test", Status: 1}
		require.NoError(t, tx.Create(&def).Error)
	}
	require.NoError(t, tx.Model(&def).Update("status", 1).Error)
	require.NoError(t, tx.Model(&wm.FlowDefinitionVersion{}).Where("definition_id=?", def.ID).Update("status", 0).Error)
	ver := wm.FlowDefinitionVersion{ID: uuid.New(), DefinitionID: def.ID, Version: 900001, BpmnData: []byte(graph), Status: 1}
	require.NoError(t, tx.Create(&ver).Error)
	center := rb.Department{ID: uuid.New(), Code: "test_" + uuid.NewString(), Name: "center"}
	require.NoError(t, tx.Create(&center).Error)
	var dept rb.Department
	if tx.Where("code='admin'").First(&dept).Error != nil {
		dept = rb.Department{ID: uuid.New(), Code: "admin", Name: "admin", ParentID: &center.ID}
		require.NoError(t, tx.Create(&dept).Error)
	} else {
		require.NoError(t, tx.Model(&dept).Updates(map[string]interface{}{"parent_id": center.ID, "status": 0}).Error)
	}
	makeUser := func() um.User {
		u := um.User{ID: uuid.New(), Username: "charter_" + uuid.NewString(), PasswordHash: "x"}
		require.NoError(t, tx.Create(&u).Error)
		return u
	}
	applicant, chair, other := makeUser(), makeUser(), makeUser()
	for _, code := range []string{"minister", "center_director", "president", "member", "probationary"} {
		require.NoError(t, tx.Exec("INSERT INTO roles(id,name,code,status) VALUES (?,?,?,0) ON CONFLICT(code) DO UPDATE SET status=0", uuid.New(), code, code).Error)
	}
	assign := func(user uuid.UUID, code string, department *uuid.UUID) {
		var role rb.Role
		require.NoError(t, tx.Where("code=?", code).First(&role).Error)
		ur := rb.UserRole{ID: uuid.New(), UserID: user, RoleID: role.ID}
		require.NoError(t, tx.Create(&ur).Error)
		if department != nil {
			require.NoError(t, tx.Exec("INSERT INTO user_role_departments VALUES (?,?)", ur.ID, *department).Error)
		}
	}
	assign(chair.ID, "minister", &dept.ID)
	assign(chair.ID, "center_director", &center.ID)
	assign(chair.ID, "president", nil)
	assign(other.ID, "minister", &dept.ID)
	tasks := wr.NewTaskRepo(tx)
	registry := nodes.NewNodeRegistry()
	registry.Register(&nodes.StartNode{})
	registry.Register(&nodes.EndNode{})
	registry.Register(&nodes.ApprovalNode{TaskRepo: tasks, Approvers: wr.NewApproverRepo(tx), EventBus: events.NewEventBus()})
	flow := engine.NewFlowEngine(wr.NewDefinitionRepo(tx), wr.NewInstanceRepo(tx), tasks, wr.NewVariableRepo(tx), tx, registry, engine.NewExpressionEngine(), events.NewEventBus(), zap.NewNop())
	now := time.Now()
	s := &admissionService{db: tx, flow: flow, now: func() time.Time { return now }}
	app := model.MemberApplication{ID: uuid.New(), UserID: applicant.ID, Type: model.ApplicantMember, RealName: "member", StudentNo: "test" + uuid.NewString()[:8], AdmissionVersion: 2, AdmissionRevision: 1, StageEnteredAt: now, SubmittedAt: now, Skills: model.JSONStrings{}, RequiredFields: model.JSONStrings{}}
	require.NoError(t, tx.Create(&app).Error)
	_, err = s.startAdmissionWorkflow(ctx, tx, app.ID)
	require.NoError(t, err)
	read := func() { require.NoError(t, tx.First(&app, "id=?", app.ID).Error) }
	read()
	require.True(t, app.CharterPolicy)
	require.Nil(t, app.DepartmentID)
	require.Equal(t, dept.ID, *app.ReviewDepartmentID)
	require.NoError(t, s.ReviewMaterials(ctx, chair.ID, app.ID, "approve", "department"))
	read()
	require.Equal(t, "中心复审", app.CurrentStage)
	require.NoError(t, s.ReviewMaterials(ctx, chair.ID, app.ID, "approve", "center"))
	read()
	require.Equal(t, "常委会会签", app.CurrentStage)
	require.NoError(t, s.ReviewMaterials(ctx, chair.ID, app.ID, "approve", "committee"))
	read()
	require.Equal(t, model.AdmissionEngine, app.AdmissionStage)
	require.NoError(t, s.ReviewMaterials(ctx, other.ID, app.ID, "approve", "committee"))
	read()
	require.Equal(t, model.AdmissionProbation, app.AdmissionStage)
	var count int64
	require.NoError(t, tx.Table("user_roles ur").Joins("JOIN roles r ON r.id=ur.role_id").Where("ur.user_id=? AND r.code='member'", app.UserID).Count(&count).Error)
	require.Zero(t, count)
	now = app.ProbationUntil.Add(time.Minute)
	require.NoError(t, s.finishProbation(ctx, app.ID))
	read()
	require.Equal(t, model.AdmissionProbation, app.AdmissionStage)
	snapshot, err := s.Snapshot(ctx, chair.ID, app.ID)
	require.NoError(t, err)
	require.Contains(t, snapshot.AllowedRoles, "minister")
	_, err = s.Sign(ctx, chair.ID, app.ID, &dto.SignAdmissionRequest{Stage: "probation", Revision: 1, Role: "minister", Decision: "approve", Comment: "confirmed"})
	require.NoError(t, err)
	read()
	require.Equal(t, model.AdmissionApproved, app.AdmissionStage)
	roles, err := repo.NewAdmissionRepo(tx).Actor(ctx, app.UserID)
	require.NoError(t, err)
	require.Contains(t, roles.Roles, "member")
	require.NotContains(t, roles.Roles, "probationary")
}
