package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *admissionService) SubmitApplication(ctx context.Context, user uuid.UUID, req *dto.SubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	var result *dto.ApplicationResponse
	var deliver func(context.Context)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		jobs := repo.NewAdmissionJobsRepo(tx)
		// Serialize duplicate submissions, including when no application row exists yet.
		if err := jobs.LockApplicant(ctx, user); err != nil {
			return err
		}
		if req.ApplicantType == 2 && req.DepartmentID == "" {
			return response.NewError(response.CodeBadRequest, "干事申请必须选择意向部门")
		}
		if err := validateApplicationDepartment(ctx, jobs, req.DepartmentID); err != nil {
			return err
		}
		if strings.TrimSpace(req.RealName) == "" || strings.TrimSpace(req.StudentNo) == "" || strings.TrimSpace(req.Reason) == "" {
			return response.NewError(response.CodeBadRequest, "姓名、学号和申请理由不能为空白")
		}
		profiles := repo.NewProfileRepo(tx)
		profile, err := profiles.GetByUserID(ctx, user)
		if err != nil {
			return err
		}
		if profile != nil {
			if profile.Status == model.ProfileDisabled {
				return admissionDenied("档案已停用，请联系管理人员处理")
			}
			if profile.Status == model.ProfileProbation {
				return response.NewError(response.CodeMemberAppDuplicate, "已有候补期申请，请等待处理")
			}
			if profile.Status == model.ProfileActive && profile.MemberType >= int16(req.ApplicantType) {
				return response.NewError(response.CodeMemberAppDuplicate, "已具有该成员身份，无需重复申请")
			}
		}
		members := &memberService{apps: repo.NewApplicationRepo(tx), profs: profiles}
		result, err = members.Submit(ctx, user, req)
		if err != nil {
			return err
		}
		id := uuid.MustParse(result.ID)
		deliver, err = s.startAdmissionWorkflow(ctx, tx, id)
		if err != nil {
			return err
		}
		result, err = members.GetApplication(ctx, user, id, nil)
		return err
	})
	if err == nil && deliver != nil {
		deliver(ctx)
	}
	return result, err
}
func (s *admissionService) ResubmitApplication(ctx context.Context, user, id uuid.UUID, req *dto.ResubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	var result *dto.ApplicationResponse
	var deliver func(context.Context)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		app, err := repo.NewAdmissionRepo(tx).LockApplication(ctx, id)
		if err != nil {
			return err
		}
		if app == nil {
			return response.NewError(response.CodeMemberAppNotFound, "申请不存在")
		}
		if app.UserID != user {
			return admissionDenied("无权操作该申请")
		}
		if app.HistoricalReviewRequired {
			return response.NewError(response.CodeMemberAppInvalid, "历史申请须由负责人核验后再处理")
		}
		if req.DepartmentID != "" {
			if err = validateApplicationDepartment(ctx, repo.NewAdmissionJobsRepo(tx), req.DepartmentID); err != nil {
				return err
			}
		}
		members := &memberService{apps: repo.NewApplicationRepo(tx), profs: repo.NewProfileRepo(tx)}
		result, err = members.Resubmit(ctx, user, id, req)
		if err != nil {
			return err
		}
		deliver, err = s.startAdmissionWorkflow(ctx, tx, id)
		if err != nil {
			return err
		}
		result, err = members.GetApplication(ctx, user, id, nil)
		return err
	})
	if err == nil && deliver != nil {
		deliver(ctx)
	}
	return result, err
}
func validateApplicationDepartment(ctx context.Context, jobs repo.AdmissionJobsRepo, value string) error {
	if value == "" {
		return nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return response.NewError(response.CodeBadRequest, "请选择有效的意向部门")
	}
	valid, err := jobs.ValidDepartment(ctx, id)
	if err != nil {
		return err
	}
	if !valid {
		return response.NewError(response.CodeBadRequest, "意向部门不存在、已停用或不是职能部门")
	}
	return nil
}
