package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *admissionService) Objection(ctx context.Context, viewer, id uuid.UUID, req *dto.AdmissionObjectionRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store, maintenance := repo.NewAdmissionRepo(tx), repo.NewAdmissionMaintenanceRepo(tx)
		app, actor, parent, err := loadAdmission(ctx, store, viewer, id)
		if err != nil {
			return err
		}
		if app.AdmissionStage != model.AdmissionProbation || app.HistoricalReviewRequired {
			return response.NewError(response.CodeMemberAppInvalid, "仅候补期可以处理异议")
		}
		if strings.TrimSpace(req.Comment) == "" {
			return response.NewError(response.CodeBadRequest, "请填写具体理由")
		}
		objection, err := maintenance.OpenObjection(ctx, id)
		if err != nil {
			return err
		}
		now := s.now()
		if req.Action == "raise" {
			allowed, delegated := admissionAuthority(actor, app, parent, "minister")
			if !allowed || delegated {
				return admissionDenied("须由候补成员所属部门的部长提出异议")
			}
			if app.ProbationUntil == nil || !now.Before(*app.ProbationUntil) {
				return response.NewError(response.CodeMemberAppInvalid, "候补期已到期")
			}
			if objection != nil {
				return response.NewError(response.CodeConflict, "已有异议正在处理")
			}
			objection = &model.AdmissionObjection{ID: uuid.New(), ApplicationID: id, RaisedBy: viewer, Reason: req.Comment, Status: "center_review", CreatedAt: now, UpdatedAt: now}
		} else {
			if objection == nil {
				return response.NewError(response.CodeNotFound, "没有待处理异议")
			}
			if req.Action == "center_review" && objection.Status == "center_review" {
				allowed, delegated := admissionAuthority(actor, app, parent, "center")
				if !allowed || delegated {
					return admissionDenied("须由相关中心负责人复核")
				}
				objection.CenterReviewerID = &viewer
				objection.CenterComment = req.Comment
				objection.Status = "president_review"
			} else if (req.Action == "uphold" || req.Action == "dismiss") && objection.Status == "president_review" {
				allowed, _ := admissionAuthority(actor, app, parent, "president")
				if !allowed {
					return admissionDenied("须由会长最终决定")
				}
				objection.FinalReviewerID = &viewer
				objection.FinalComment = req.Comment
				objection.Status = req.Action
				if req.Action == "uphold" {
					app.AdmissionStage = model.AdmissionRejected
					app.Status = model.AppRejected
					app.CurrentStage = "候补异议通过，终止录用"
					app.UpdatedAt = now
					if err := store.SaveApplication(ctx, app); err != nil {
						return err
					}
					if err := repo.NewAdmissionProfileRepo(tx).Restore(ctx, app.ID, app.UserID, now); err != nil {
						return err
					}
				}
			} else {
				return response.NewError(response.CodeConflict, "异议进度已变化，或尚未完成前置复核")
			}
			objection.UpdatedAt = now
		}
		if err := maintenance.SaveObjection(ctx, objection); err != nil {
			return fmt.Errorf("save objection: %w", err)
		}
		members := &memberService{apps: repo.NewApplicationRepo(tx)}
		return members.recordAppHistory(ctx, id, model.AppApproved, app.Status, &viewer, req.Comment, map[string]interface{}{"objection_id": objection.ID, "action": req.Action})
	})
}
