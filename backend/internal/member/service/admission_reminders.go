package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
)

// Durable invalidations survive Redis failures and process restarts. A revision
// prevents an older worker from deleting a newly queued permission change.
func (s *admissionService) refreshPermissions(ctx context.Context) error {
	if s.permissions == nil {
		return nil
	}
	jobs := repo.NewAdmissionJobsRepo(s.db)
	rows, err := jobs.PermissionRefreshes(ctx)
	if err != nil {
		return err
	}
	for _, item := range rows {
		if err = s.permissions.InvalidateUserPermissions(ctx, item.UserID); err != nil {
			return err
		}
		if err = jobs.CompletePermissionRefresh(ctx, item); err != nil {
			return err
		}
	}
	return nil
}
func (s *admissionService) remindOverdue(ctx context.Context) error {
	ids, err := repo.NewAdmissionJobsRepo(s.db).OverdueApplications(ctx, s.now().Add(-24*time.Hour))
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err = s.remindApplication(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
func (s *admissionService) remindApplication(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store, jobs := repo.NewAdmissionRepo(tx), repo.NewAdmissionJobsRepo(tx)
		app, err := store.LockApplication(ctx, id)
		if err != nil {
			return err
		}
		if app == nil || app.HistoricalReviewRequired || s.now().Before(app.StageEnteredAt.Add(24*time.Hour)) {
			return nil
		}
		roles := requiredAdmissionRoles(app.AdmissionStage)
		if app.CharterPolicy && app.AdmissionStage == model.AdmissionEngine && app.FlowInstanceID != nil && s.flow != nil {
			flow, _, err := s.flow.BindTransaction(tx)
			if err != nil {
				return err
			}
			node, done, err := flow.RunningApprovalNode(ctx, *app.FlowInstanceID)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
			role, _, err := flow.ApprovalPolicy(ctx, *app.FlowInstanceID, node)
			if err != nil {
				return err
			}
			roles = []string{role}
		}
		if app.CharterPolicy && app.AdmissionStage == model.AdmissionProbation && app.ProbationUntil != nil && !s.now().Before(*app.ProbationUntil) {
			roles = []string{"minister"}
		}
		if len(roles) == 0 {
			return nil
		}
		var parent *uuid.UUID
		if dept := approvalDepartment(app); dept != nil {
			parent, err = store.ParentDepartment(ctx, *dept)
			if err != nil {
				return err
			}
		}
		actors, err := jobs.Reviewers(ctx)
		if err != nil {
			return err
		}
		signatures, err := store.Signatures(ctx, id)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("overdue:%s:%d:%s", id, app.AdmissionRevision, repo.ReminderStage(app))
		recipients := 0
		for _, actor := range actors {
			allowed := false
			for _, role := range roles {
				if signedAdmissionRole(signatures, app, role) {
					continue
				}
				if ok, _ := admissionAuthority(&actor, app, parent, role); ok {
					allowed = true
				}
			}
			if !allowed {
				continue
			}
			if err = jobs.Notify(ctx, actor.ID, key, "入会审批已等待超过24小时", fmt.Sprintf("申请 %s 的%s环节仍待处理。请核对进度；上级代签须记录原因，不能跳过面试或自动通过。", app.RealName, admissionStageLabel(app.AdmissionStage)), s.now()); err != nil {
				return err
			}
			recipients++
		}
		// Keep retrying if the department has no eligible reviewer configured yet.
		if recipients == 0 {
			return nil
		}
		return jobs.MarkReminded(ctx, app)
	})
}
