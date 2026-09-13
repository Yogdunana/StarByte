package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
)

func (s *admissionService) Maintenance(ctx context.Context, _ string, logf func(string)) error {
	if err := s.refreshPermissions(ctx); err != nil {
		return err
	}
	if err := s.remindOverdue(ctx); err != nil {
		return err
	}
	ids, err := repo.NewAdmissionMaintenanceRepo(s.db).DueApplications(ctx, s.now())
	if err != nil {
		return fmt.Errorf("list due probation: %w", err)
	}
	for _, id := range ids {
		if err := s.finishProbation(ctx, id); err != nil {
			return err
		}
	}
	logf(fmt.Sprintf("检查候补到期申请 %d 项", len(ids)))
	return nil
}
func (s *admissionService) finishProbation(ctx context.Context, id uuid.UUID) error {

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store, maintenance := repo.NewAdmissionRepo(tx), repo.NewAdmissionMaintenanceRepo(tx)
		app, err := store.LockApplication(ctx, id)
		if err != nil {
			return err
		}
		if app == nil {
			return nil
		}
		now := s.now()
		if app.CharterPolicy || app.HistoricalReviewRequired || app.AdmissionStage != model.AdmissionProbation || app.ProbationUntil == nil || now.Before(*app.ProbationUntil) {
			return nil
		}
		objection, err := maintenance.OpenObjection(ctx, id)
		if err != nil {
			return err
		}
		if objection != nil {
			return nil
		}
		if err := maintenance.ActivateProfile(ctx, app.UserID, now); err != nil {
			return fmt.Errorf("activate probation profile: %w", err)
		}
		if err := maintenance.GrantRole(ctx, app.UserID, "officer", app.DepartmentID); err != nil {
			return err
		}
		if err := repo.NewAdmissionJobsRepo(tx).QueuePermissionRefresh(ctx, app.UserID); err != nil {
			return err
		}
		app.AdmissionStage = model.AdmissionApproved
		app.CurrentStage = "正式成员"
		app.StageEnteredAt = now
		app.UpdatedAt = now
		if err := store.SaveApplication(ctx, app); err != nil {
			return err
		}
		members := &memberService{apps: repo.NewApplicationRepo(tx)}
		return members.recordAppHistory(ctx, id, app.Status, app.Status, nil, "候补一个自然月届满，无待处理异议，转为正式成员", map[string]interface{}{"event": "probation_completed"})
	})
	if err == nil {
		_ = s.refreshPermissions(ctx)
	}
	return err
}
