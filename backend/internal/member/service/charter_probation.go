package service

import (
	"context"
	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *admissionService) confirmCharterProbation(ctx context.Context, tx *gorm.DB, store repo.AdmissionRepo, app *model.MemberApplication, actor *model.AdmissionActor, parent *uuid.UUID, req *dto.SignAdmissionRequest) error {
	now := s.now()
	ok, delegated := admissionAuthority(actor, app, parent, "minister")
	if !ok || delegated || req.Role != "minister" || req.Stage != app.AdmissionStage || req.Revision != app.AdmissionRevision || req.Decision != "approve" || app.ProbationUntil == nil || now.Before(*app.ProbationUntil) {
		return admissionDenied("候补期届满后须由直属部长确认转正")
	}
	pending, err := store.OpenObjection(ctx, app.ID)
	if err != nil {
		return err
	}
	if pending != nil {
		return admissionDenied("请先处理候补异议")
	}
	if err := ensureActiveApplicant(ctx, tx, app.UserID); err != nil {
		return err
	}
	signature := model.AdmissionSignature{ID: uuid.New(), ApplicationID: app.ID, Revision: app.AdmissionRevision, Stage: app.AdmissionStage, SignerID: actor.ID, SignerRole: "minister", Decision: "approve", Comment: req.Comment, CreatedAt: now}
	if err := store.AddSignature(ctx, &signature); err != nil {
		return err
	}
	maintenance := repo.NewAdmissionMaintenanceRepo(tx)
	if err := maintenance.ActivateProfile(ctx, app.UserID, now); err != nil {
		return err
	}
	if err := maintenance.GrantRole(ctx, app.UserID, membershipRole(app), app.DepartmentID); err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM user_roles WHERE user_id=? AND role_id IN (SELECT id FROM roles WHERE code='probationary')", app.UserID).Error; err != nil {
		return err
	}
	app.AdmissionStage = model.AdmissionApproved
	app.CurrentStage = "正式成员"
	app.StageEnteredAt = now
	app.UpdatedAt = now
	if err := store.SaveApplication(ctx, app); err != nil {
		return err
	}
	if err := repo.NewAdmissionJobsRepo(tx).QueuePermissionRefresh(ctx, app.UserID); err != nil {
		return err
	}
	members := &memberService{apps: repo.NewApplicationRepo(tx)}
	return members.recordAppHistory(ctx, app.ID, app.Status, app.Status, &actor.ID, req.Comment, map[string]interface{}{"event": "probation_confirmed", "signature_id": signature.ID})
}
