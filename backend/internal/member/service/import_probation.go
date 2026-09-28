package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
)

// promoteImportedProbations 把批量录入、预备期已经届满的人转成正式干事。
//
// 走申请流程的人由 finishProbation 按 member_applications.probation_until 转正，
// 批量录入的人没有申请记录，所以这里按 member_profiles.probation_until（000079）再扫一遍。
func (s *admissionService) promoteImportedProbations(ctx context.Context) (int, error) {
	rows, err := repo.NewMemberImportRepo(s.db).DueImportedProbations(ctx, s.now())
	if err != nil {
		return 0, fmt.Errorf("list due imported probation: %w", err)
	}
	done := 0
	for _, row := range rows {
		ok, err := s.promoteImportedProbation(ctx, row)
		if err != nil {
			return done, err
		}
		if ok {
			done++
		}
	}
	return done, nil
}

func (s *admissionService) promoteImportedProbation(ctx context.Context, row repo.ImportedProbation) (bool, error) {
	applied := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := s.now()
		jobs := repo.NewAdmissionJobsRepo(tx)
		if err := jobs.LockApplicant(ctx, row.UserID); err != nil {
			return err
		}
		// ActivateProfile 只认「当前恰好是预备期」这一种情况，
		// 期间被别的流程改过就返回 ErrRecordNotFound，这里当跳过处理。
		if err := repo.NewAdmissionMaintenanceRepo(tx).ActivateProfile(ctx, row.UserID, now); err != nil {
			return err
		}
		if err := repo.NewAdmissionMaintenanceRepo(tx).GrantRole(ctx, row.UserID, "officer", row.DepartmentID); err != nil {
			return err
		}
		if err := repo.NewMemberImportRepo(tx).ClearProbationUntil(ctx, row.UserID); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Exec(
			"DELETE FROM user_roles WHERE user_id=? AND role_id IN (SELECT id FROM roles WHERE code='probationary')",
			row.UserID).Error; err != nil {
			return err
		}
		if err := jobs.QueuePermissionRefresh(ctx, row.UserID); err != nil {
			return err
		}
		applied = true
		return nil
	})
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return applied, nil
}
