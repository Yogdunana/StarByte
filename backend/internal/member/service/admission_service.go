package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type AdmissionService interface {
	SubmitApplication(context.Context, uuid.UUID, *dto.SubmitApplicationRequest) (*dto.ApplicationResponse, error)
	ResubmitApplication(context.Context, uuid.UUID, uuid.UUID, *dto.ResubmitApplicationRequest) (*dto.ApplicationResponse, error)
	Maintenance(context.Context, string, func(string)) error
	Objection(context.Context, uuid.UUID, uuid.UUID, *dto.AdmissionObjectionRequest) error
	Snapshot(context.Context, uuid.UUID, uuid.UUID) (*dto.AdmissionResponse, error)
	Sign(context.Context, uuid.UUID, uuid.UUID, *dto.SignAdmissionRequest) (*dto.AdmissionResponse, error)
	ReviewMaterials(context.Context, uuid.UUID, uuid.UUID, string, string) error
	RequestSupplement(context.Context, uuid.UUID, uuid.UUID, *dto.SupplementRequest) error
	TransferReview(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) error
	ApplicationProgress(context.Context, uuid.UUID, uuid.UUID) (*dto.ApplicationProgressResponse, error)
	TransferCandidates(context.Context, uuid.UUID, uuid.UUID, string) ([]dto.TransferCandidate, error)
}
type AdmissionPermissionCache interface {
	InvalidateUserPermissions(context.Context, uuid.UUID) error
}
type admissionService struct {
	flow        *engine.FlowEngine
	permissions AdmissionPermissionCache
	db          *gorm.DB
	now         func() time.Time
}

func NewAdmissionService(db *gorm.DB, caches ...AdmissionPermissionCache) AdmissionService {
	s := &admissionService{db: db, now: time.Now}
	if len(caches) > 0 {
		s.permissions = caches[0]
	}
	return s
}
func admissionDenied(message string) error { return response.NewError(response.CodeForbidden, message) }

func (s *admissionService) Snapshot(ctx context.Context, viewer, id uuid.UUID) (*dto.AdmissionResponse, error) {
	var result *dto.AdmissionResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store := repo.NewAdmissionRepo(tx)
		app, actor, parent, err := loadAdmission(ctx, store, viewer, id)
		if err != nil {
			return err
		}
		allowed, _ := admissionAuthority(actor, app, parent, "materials")
		if app.UserID != viewer && !allowed {
			return admissionDenied("无权查看该申请审批记录")
		}
		result, err = s.snapshot(ctx, store, app, actor, parent)
		return err
	})
	return result, err
}
func loadAdmission(ctx context.Context, store repo.AdmissionRepo, viewer, id uuid.UUID) (*model.MemberApplication, *model.AdmissionActor, *uuid.UUID, error) {
	app, err := store.LockApplication(ctx, id)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load admission: %w", err)
	}
	if app == nil {
		return nil, nil, nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
	}
	actor, err := store.Actor(ctx, viewer)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load signer: %w", err)
	}
	if actor == nil {
		return nil, nil, nil, admissionDenied("签字账号不可用")
	}
	var parent *uuid.UUID
	if app.DepartmentID != nil {
		parent, err = store.ParentDepartment(ctx, *app.DepartmentID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("load department: %w", err)
		}
	}
	return app, actor, parent, nil
}
func (s *admissionService) snapshot(ctx context.Context, store repo.AdmissionRepo, app *model.MemberApplication, actor *model.AdmissionActor, parent *uuid.UUID) (*dto.AdmissionResponse, error) {
	signatures, err := store.Signatures(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	complete := true
	if round := admissionRound(app.AdmissionStage); round > 0 {
		complete, err = store.InterviewCompleted(ctx, app.ID, round, app.StageEnteredAt)
		if err != nil {
			return nil, err
		}
	}
	allowed := []string{}
	if !app.HistoricalReviewRequired && complete {
		for _, role := range requiredAdmissionRoles(app.AdmissionStage) {
			permitted, delegated := admissionAuthority(actor, app, parent, role)
			if permitted && (!delegated || !s.now().Before(app.StageEnteredAt.Add(24*time.Hour))) && !signedAdmissionRole(signatures, app, role) {
				allowed = append(allowed, role)
			}
		}
	}
	maintenance := store
	objections, err := maintenance.Objections(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	objectionActions := []string{}
	if app.AdmissionStage == model.AdmissionProbation {
		pending, err := maintenance.OpenObjection(ctx, app.ID)
		if err != nil {
			return nil, err
		}
		if pending == nil && app.ProbationUntil != nil && s.now().Before(*app.ProbationUntil) {
			if ok, delegated := admissionAuthority(actor, app, parent, "minister"); ok && !delegated {
				objectionActions = append(objectionActions, "raise")
			}
		}
		if pending != nil && pending.Status == "center_review" {
			if ok, delegated := admissionAuthority(actor, app, parent, "center"); ok && !delegated {
				objectionActions = append(objectionActions, "center_review")
			}
		}
		if pending != nil && pending.Status == "president_review" {
			if ok, _ := admissionAuthority(actor, app, parent, "president"); ok {
				objectionActions = append(objectionActions, "uphold", "dismiss")
			}
		}
	}
	staff := app.UserID != actor.ID
	return &dto.AdmissionResponse{Objections: objectionViews(objections, staff), AllowedObjectionActions: objectionActions, ApplicationID: app.ID.String(), Stage: app.AdmissionStage, Revision: app.AdmissionRevision, HistoricalReviewRequired: app.HistoricalReviewRequired, Signatures: signatures, AllowedRoles: allowed, InterviewCompleted: complete}, nil
}

// objectionViews keeps probation-objection internals off the applicant snapshot.
// Officers still see who raised it and the review comments.
func objectionViews(items []model.AdmissionObjection, staff bool) []dto.AdmissionObjectionView {
	out := make([]dto.AdmissionObjectionView, 0, len(items))
	for _, item := range items {
		view := dto.AdmissionObjectionView{
			ID:        item.ID.String(),
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
		}
		if staff {
			view.RaisedBy = item.RaisedBy.String()
			view.Reason = item.Reason
			view.CenterComment = item.CenterComment
			view.FinalComment = item.FinalComment
			if item.CenterReviewerID != nil {
				view.CenterReviewerID = item.CenterReviewerID.String()
			}
			if item.FinalReviewerID != nil {
				view.FinalReviewerID = item.FinalReviewerID.String()
			}
		}
		out = append(out, view)
	}
	return out
}
func (s *admissionService) Sign(ctx context.Context, viewer, id uuid.UUID, req *dto.SignAdmissionRequest) (*dto.AdmissionResponse, error) {
	var out *dto.AdmissionResponse
	var deliver func(context.Context)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store := repo.NewAdmissionRepo(tx)
		app, actor, parent, err := loadAdmission(ctx, store, viewer, id)
		if err != nil {
			return err
		}
		if app.HistoricalReviewRequired || app.AdmissionVersion < 2 {
			return response.NewError(response.CodeMemberAppInvalid, "历史申请须先核验，不能自动补签")
		}
		engineChain, err := s.isEngineChain(ctx, store, app)
		if err != nil {
			return err
		}
		if engineChain {
			return response.NewError(response.CodeMemberAppInvalid, "该申请已接入流程引擎，请使用入会审批，不能走章程签字")
		}
		if req.Stage != app.AdmissionStage || req.Revision != app.AdmissionRevision {
			return response.NewError(response.CodeConflict, "申请进度已变化，请刷新后操作")
		}
		permitted, delegated := admissionAuthority(actor, app, parent, req.Role)
		if !permitted {
			return admissionDenied("无权在该环节签字")
		}
		if delegated && (strings.TrimSpace(req.DelegationReason) == "" || s.now().Before(app.StageEnteredAt.Add(24*time.Hour))) {
			return admissionDenied("超时24小时后上级才可代签，并须填写原因")
		}
		snapshot, err := s.snapshot(ctx, store, app, actor, parent)
		if err != nil {
			return err
		}
		roleAllowed := false
		for _, role := range snapshot.AllowedRoles {
			if role == req.Role {
				roleAllowed = true
			}
		}
		if !roleAllowed {
			return response.NewError(response.CodeMemberAppInvalid, "该环节尚未开放或已签字；请先完成相应面试")
		}
		if req.Decision != "approve" && req.Decision != "reject" && req.Decision != "supplement" {
			return response.NewError(response.CodeBadRequest, "无效的审批决定")
		}
		if req.Decision != "approve" && strings.TrimSpace(req.Comment) == "" {
			return response.NewError(response.CodeBadRequest, "拒绝或补充材料须填写原因")
		}
		if req.Decision == "supplement" && app.AdmissionStage != model.AdmissionMaterials {
			return response.NewError(response.CodeMemberAppInvalid, "只有资料审核环节可以要求补充材料")
		}
		for _, field := range req.RequiredFields {
			switch field {
			case "real_name", "student_no", "reason", "skills", "experience", "contact_phone", "contact_email":
			default:
				return response.NewError(response.CodeBadRequest, "不支持的补充字段")
			}
		}
		from, now := app.Status, s.now()
		signature := model.AdmissionSignature{ID: uuid.New(), ApplicationID: id, Revision: app.AdmissionRevision, Stage: app.AdmissionStage, SignerID: viewer, SignerRole: req.Role, Decision: req.Decision, Comment: req.Comment, Delegated: delegated, DelegationReason: req.DelegationReason, CreatedAt: now}
		if err := store.AddSignature(ctx, &signature); err != nil {
			return fmt.Errorf("save signature: %w", err)
		}
		snapshot.Signatures = append(snapshot.Signatures, signature)
		oldStage := app.AdmissionStage
		switch req.Decision {
		case "reject":
			app.Status = model.AppRejected
			app.AdmissionStage = model.AdmissionRejected
		case "supplement":
			app.Status = model.AppSupplement
			app.AdmissionStage = model.AdmissionSupplement
			app.RequiredFields = nonemptyStrings(req.RequiredFields)
		case "approve":
			allSigned := true
			for _, role := range requiredAdmissionRoles(app.AdmissionStage) {
				if !signedAdmissionRole(snapshot.Signatures, app, role) {
					allSigned = false
				}
			}
			if allSigned {
				advanceAdmission(app, now)
			}
		}
		if oldStage != app.AdmissionStage {
			app.StageEnteredAt = now
		}
		app.UpdatedAt = now
		app.ReviewedAt = &now
		app.ReviewerID = &viewer
		app.ReviewComment = req.Comment
		app.CurrentStage = admissionStageLabel(app.AdmissionStage)
		deliver, err = s.advanceWorkflow(ctx, tx, app, &signature)
		if err != nil {
			return err
		}
		if err := store.SaveApplication(ctx, app); err != nil {
			return fmt.Errorf("save admission: %w", err)
		}
		members := &memberService{apps: repo.NewApplicationRepo(tx), profs: repo.NewProfileRepo(tx)}
		if app.AdmissionStage == model.AdmissionApproved {
			if err := repo.NewAdmissionMaintenanceRepo(tx).GrantRole(ctx, app.UserID, "member", app.DepartmentID); err != nil {
				return fmt.Errorf("grant membership: %w", err)
			}
			if err := repo.NewAdmissionJobsRepo(tx).QueuePermissionRefresh(ctx, app.UserID); err != nil {
				return err
			}
		}
		if app.AdmissionStage == model.AdmissionProbation {
			if err := repo.NewAdmissionProfileRepo(tx).Capture(ctx, app.ID, app.UserID); err != nil {
				return err
			}
		}
		if app.AdmissionStage == model.AdmissionApproved || app.AdmissionStage == model.AdmissionProbation {
			if err := members.ensureProfile(ctx, app); err != nil {
				return err
			}
		}
		if err := members.recordAppHistory(ctx, id, from, app.Status, &viewer, req.Comment, map[string]interface{}{"signature_id": signature.ID, "stage": signature.Stage, "role": signature.SignerRole, "decision": signature.Decision}); err != nil {
			return err
		}
		out, err = s.snapshot(ctx, store, app, actor, parent)
		return err
	})
	if err == nil {
		if deliver != nil {
			deliver(ctx)
		}
		_ = s.refreshPermissions(ctx)
	} // Retry failures from the durable scheduler queue.
	return out, err
}
func (s *admissionService) ReviewMaterials(ctx context.Context, viewer, id uuid.UUID, action, comment string) error {
	handled, err := s.tryEngineReview(ctx, viewer, id, action, comment, nil)
	if handled || err != nil {
		return err
	}
	snapshot, err := s.Snapshot(ctx, viewer, id)
	if err != nil {
		return err
	}
	if snapshot.Stage != model.AdmissionMaterials {
		return response.NewError(response.CodeMemberAppInvalid, "请在正式审批面板签字，不可跳过审批环节")
	}
	_, err = s.Sign(ctx, viewer, id, &dto.SignAdmissionRequest{Stage: snapshot.Stage, Revision: snapshot.Revision, Role: "materials", Decision: action, Comment: comment})
	return err
}

func (s *admissionService) RequestSupplement(ctx context.Context, viewer, id uuid.UUID, req *dto.SupplementRequest) error {
	handled, err := s.tryEngineReview(ctx, viewer, id, actionSupplement, req.Comment, req.RequiredFields)
	if handled || err != nil {
		return err
	}
	snapshot, err := s.Snapshot(ctx, viewer, id)
	if err != nil {
		return err
	}
	_, err = s.Sign(ctx, viewer, id, &dto.SignAdmissionRequest{Stage: snapshot.Stage, Revision: snapshot.Revision, Role: "materials", Decision: "supplement", Comment: req.Comment, RequiredFields: req.RequiredFields})
	return err
}
