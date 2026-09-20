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
	normalized, err := bindContactPhone(req.ContactPhone)
	if err != nil {
		return nil, err
	}
	req.ContactPhone = normalized
	var result *dto.ApplicationResponse
	var deliver func(context.Context)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		jobs := repo.NewAdmissionJobsRepo(tx)
		// Serialize duplicate submissions, including when no application row exists yet.
		if err := jobs.LockApplicant(ctx, user); err != nil {
			return err
		}
		if err := bindApplicantDepartment(req); err != nil {
			return err
		}
		if err := validateApplicationDepartment(ctx, jobs, req.DepartmentID); err != nil {
			return err
		}
		if strings.TrimSpace(req.RealName) == "" || strings.TrimSpace(req.StudentNo) == "" {
			return response.NewError(response.CodeBadRequest, "姓名和学号不能为空白")
		}
		if err := applyApplicantGender(ctx, tx, user, req.Gender); err != nil {
			return err
		}
		profiles := repo.NewProfileRepo(tx)
		profile, err := profiles.GetByUserID(ctx, user)
		if err != nil {
			return err
		}
		actor, err := repo.NewAdmissionRepo(tx).Actor(ctx, user)
		if err != nil {
			return err
		}
		var roles []string
		if actor != nil {
			roles = actor.Roles
		}
		if err := rejectDuplicateMembership(req.ApplicantType, profile, roles); err != nil {
			return err
		}
		if err := requireOfficerIsMember(req.ApplicantType, profile); err != nil {
			return err
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
	if err == nil && result != nil && result.Status == model.AppApproved {
		_ = s.refreshPermissions(ctx)
	}
	return result, err
}
func (s *admissionService) ResubmitApplication(ctx context.Context, user, id uuid.UUID, req *dto.ResubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	if err := normalizeResubmitPhone(req); err != nil {
		return nil, err
	}
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
	if err == nil && result != nil && result.Status == model.AppApproved {
		_ = s.refreshPermissions(ctx)
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

func requireOfficerIsMember(applicantType int, profile *model.MemberProfile) error {
	if applicantType != int(model.ApplicantOfficer) {
		return nil
	}
	if profile == nil || profile.Status != model.ProfileActive || profile.MemberType < model.MemberTypeMember {
		return response.NewError(response.CodeBadRequest, "须先成为会员后再申请干事或干部职务")
	}
	return nil
}

func applyApplicantGender(ctx context.Context, tx *gorm.DB, user uuid.UUID, gender *int) error {
	if gender != nil && (*gender == 1 || *gender == 2) {
		if err := tx.WithContext(ctx).Table("users").Where("id = ?", user).Update("gender", *gender).Error; err != nil {
			return err
		}
	}
	var stored int
	if err := tx.WithContext(ctx).Table("users").Where("id = ?", user).Select("gender").Take(&stored).Error; err != nil {
		return err
	}
	if stored != 1 && stored != 2 {
		return response.NewError(response.CodeBadRequest, "请选择性别")
	}
	return nil
}
