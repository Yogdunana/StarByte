package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/discipline/dto"
	"github.com/Yogdunana/StarByte/backend/internal/discipline/model"
	"github.com/Yogdunana/StarByte/backend/internal/discipline/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	Create(ctx context.Context, operator uuid.UUID, req *dto.CreateRecordRequest, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error)
	Update(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateRecordRequest, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error)
	Get(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error)
	List(ctx context.Context, viewer uuid.UUID, req *dto.ListRecordRequest, scope *rbacModel.DataScopeCondition) ([]*dto.RecordResponse, int64, int, int, error)
	Approve(ctx context.Context, operator, id uuid.UUID, comment string, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error)
	Revoke(ctx context.Context, operator, id uuid.UUID, reason string, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error)
	Appeal(ctx context.Context, operator, id uuid.UUID, reason string, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error)
}

type disciplineService struct {
	rows   repo.Repository
	notify Notifier
	flow   FlowStarter
}

func New(rows repo.Repository, notify Notifier, flow FlowStarter) Service {
	return &disciplineService{rows: rows, notify: notify, flow: flow}
}

func (s *disciplineService) Create(ctx context.Context, operator uuid.UUID, req *dto.CreateRecordRequest, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
	if !model.ValidLevel(req.Level) {
		return nil, response.NewError(response.CodeDisciplineInvalidLevel, "处分等级不合法")
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "用户 ID 无效")
	}
	target, err := s.rows.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if target == nil {
		return nil, response.NewError(response.CodeBadRequest, "处分对象不存在")
	}
	if !canAccess(scope, uid, target.DepartmentID, operator) {
		return nil, response.NewError(response.CodeDisciplineNoAccess, "无权为该成员登记处分")
	}
	now := time.Now()
	row := &model.Record{
		ID: uuid.New(), UserID: uid, Title: strings.TrimSpace(req.Title), Description: req.Description,
		Level: req.Level, Status: model.StatusPending, IssuedBy: &operator, IssuedAt: now, CreatedAt: now,
	}
	if err := s.rows.Create(ctx, row); err != nil {
		return nil, fmt.Errorf("create discipline: %w", err)
	}
	if s.flow != nil {
		if fid, ferr := s.flow.Start(ctx, row.ID, operator, map[string]interface{}{"title": row.Title, "user_id": uid.String()}); ferr == nil && fid != nil {
			row.FlowInstanceID = fid
			_ = s.rows.Update(ctx, row)
		}
	}
	s.notifyUser(ctx, uid, "discipline_notice", row.Title)
	return s.Get(ctx, operator, row.ID, nil)
}

func (s *disciplineService) Update(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateRecordRequest, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
	row, err := s.load(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	if row.Status != model.StatusPending {
		return nil, response.NewError(response.CodeDisciplineInvalidState, "仅待审批处分可修改")
	}
	if req.Title != nil {
		row.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		row.Description = *req.Description
	}
	if req.Level != nil {
		if !model.ValidLevel(*req.Level) {
			return nil, response.NewError(response.CodeDisciplineInvalidLevel, "处分等级不合法")
		}
		row.Level = *req.Level
	}
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("update discipline: %w", err)
	}
	return s.Get(ctx, operator, id, nil)
}

func (s *disciplineService) Get(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
	named, err := s.rows.GetNamed(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get discipline: %w", err)
	}
	if named == nil {
		return nil, response.NewError(response.CodeDisciplineNotFound, "处分记录不存在")
	}
	if !canAccess(scope, named.UserID, named.DepartmentID, viewer) {
		return nil, response.NewError(response.CodeDisciplineNoAccess, "无权查看该处分")
	}
	appeals, err := s.rows.ListAppeals(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list appeals: %w", err)
	}
	return mapRecord(named, appeals), nil
}

func (s *disciplineService) List(ctx context.Context, viewer uuid.UUID, req *dto.ListRecordRequest, scope *rbacModel.DataScopeCondition) ([]*dto.RecordResponse, int64, int, int, error) {
	rows, total, err := s.rows.List(ctx, req, rewriteScope(scope, viewer))
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("list discipline: %w", err)
	}
	page, size := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	out := make([]*dto.RecordResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapRecord(&rows[i], nil))
	}
	return out, total, page, size, nil
}

func (s *disciplineService) Approve(ctx context.Context, operator, id uuid.UUID, _ string, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
	row, err := s.load(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	if row.Status != model.StatusPending && row.Status != model.StatusAppealing {
		return nil, response.NewError(response.CodeDisciplineInvalidState, "当前状态不可审批")
	}
	now := time.Now()
	row.Status = model.StatusActive
	row.ApprovedBy = &operator
	row.ApprovedAt = &now
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("approve discipline: %w", err)
	}
	s.notifyUser(ctx, row.UserID, "discipline_notice", row.Title)
	return s.Get(ctx, operator, id, nil)
}

func (s *disciplineService) Revoke(ctx context.Context, operator, id uuid.UUID, reason string, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
	row, err := s.load(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	if row.Status == model.StatusRevoked {
		return nil, response.NewError(response.CodeDisciplineInvalidState, "处分已撤销")
	}
	now := time.Now()
	row.Status = model.StatusRevoked
	row.RevokeReason = strings.TrimSpace(reason)
	row.RevokedBy = &operator
	row.RevokedAt = &now
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("revoke discipline: %w", err)
	}
	s.notifyUser(ctx, row.UserID, "discipline_notice", row.Title)
	return s.Get(ctx, operator, id, nil)
}

func (s *disciplineService) Appeal(ctx context.Context, operator, id uuid.UUID, reason string, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
	row, err := s.load(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	if row.UserID != operator {
		return nil, response.NewError(response.CodeDisciplineNoAccess, "只能申诉本人的处分")
	}
	if row.Status == model.StatusAppealing {
		return nil, response.NewError(response.CodeDisciplineDupAppeal, "已有进行中的申诉")
	}
	if row.Status != model.StatusActive {
		return nil, response.NewError(response.CodeDisciplineInvalidState, "仅生效中的处分可申诉")
	}
	open, err := s.rows.HasOpenAppeal(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("check appeal: %w", err)
	}
	if open {
		return nil, response.NewError(response.CodeDisciplineDupAppeal, "已有进行中的申诉")
	}
	appeal := &model.Appeal{
		ID: uuid.New(), RecordID: id, ApplicantID: operator,
		Reason: strings.TrimSpace(reason), Status: model.AppealPending, CreatedAt: time.Now(),
	}
	if err := s.rows.CreateAppeal(ctx, appeal); err != nil {
		return nil, fmt.Errorf("create appeal: %w", err)
	}
	row.Status = model.StatusAppealing
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("mark appealing: %w", err)
	}
	return s.Get(ctx, operator, id, nil)
}

func (s *disciplineService) load(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*model.Record, error) {
	row, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get discipline: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeDisciplineNotFound, "处分记录不存在")
	}
	user, err := s.rows.GetUser(ctx, row.UserID)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	var dept *uuid.UUID
	if user != nil {
		dept = user.DepartmentID
	}
	if !canAccess(scope, row.UserID, dept, viewer) {
		return nil, response.NewError(response.CodeDisciplineNoAccess, "无权操作该处分")
	}
	return row, nil
}

func (s *disciplineService) notifyUser(ctx context.Context, userID uuid.UUID, tpl, title string) {
	if s.notify == nil {
		return
	}
	name := ""
	if u, err := s.rows.GetUser(ctx, userID); err == nil {
		name = displayName(u)
	}
	if err := s.notify.Send(ctx, []uuid.UUID{userID}, tpl, map[string]interface{}{"real_name": name, "title": title}); err != nil {
		logger.Warn("discipline notify failed", zap.Error(err))
	}
}

func mapRecord(row *model.RecordNamed, appeals []model.Appeal) *dto.RecordResponse {
	out := &dto.RecordResponse{
		ID: row.ID.String(), User: dto.Person{ID: row.UserID.String(), Name: row.UserName},
		Title: row.Title, Description: row.Description, Level: row.Level, Status: row.Status,
		IssuedAt: row.IssuedAt, ApprovedAt: row.ApprovedAt, RevokeReason: row.RevokeReason,
		RevokedAt: row.RevokedAt, CreatedAt: row.CreatedAt,
	}
	if row.IssuedBy != nil {
		out.IssuedBy = &dto.Person{ID: row.IssuedBy.String(), Name: row.IssuerName}
	}
	if row.FlowInstanceID != nil {
		out.FlowInstanceID = row.FlowInstanceID.String()
	}
	for _, a := range appeals {
		out.Appeals = append(out.Appeals, dto.AppealResponse{
			ID: a.ID.String(), Reason: a.Reason, Status: a.Status, CreatedAt: a.CreatedAt, ReviewedAt: a.ReviewedAt,
		})
	}
	return out
}
