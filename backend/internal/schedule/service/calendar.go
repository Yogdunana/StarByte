package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *scheduleService) ListCalendars(ctx context.Context, viewer uuid.UUID, req *dto.ListCalendarRequest, scope *rbacModel.DataScopeCondition) ([]*dto.CalendarResponse, int64, int, int, error) {
	if req == nil {
		req = &dto.ListCalendarRequest{}
	}
	if err := s.ensurePersonal(ctx, viewer); err != nil {
		return nil, 0, 0, 0, err
	}
	virtuals := s.virtualCalendars(viewer, req)
	page, size := normalizePage(req.Page, req.PageSize)
	start := (page - 1) * size
	nVirt := len(virtuals)
	storedReq := *req
	storedOff, storedLim := storedWindow(start, size, nVirt)
	storedReq.Offset = &storedOff
	storedReq.Limit = &storedLim
	rows, storedTotal, err := s.rows.ListCalendars(ctx, viewer, &storedReq, scope)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("list calendars: %w", err)
	}
	out := make([]*dto.CalendarResponse, 0, size)
	if start < nVirt {
		endV := nVirt
		if endV > start+size {
			endV = start + size
		}
		out = append(out, virtuals[start:endV]...)
	}
	for i := range rows {
		out = append(out, mapCalendar(&rows[i], viewer, scope))
	}
	return out, storedTotal + int64(nVirt), page, size, nil
}

// storedWindow 把虚拟层当作列表前缀后，计算存储日历的 offset/limit。
func storedWindow(start, size, nVirt int) (int, int) {
	if start >= nVirt {
		return start - nVirt, size
	}
	taken := nVirt - start
	if taken >= size {
		return 0, 0
	}
	return 0, size - taken
}

func (s *scheduleService) CreateCalendar(ctx context.Context, operator uuid.UUID, req *dto.CreateCalendarRequest, scope *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error) {
	if !model.ValidCalendarType(req.CalendarType) {
		return nil, response.NewError(response.CodeBadRequest, "日历类型不合法")
	}
	owner := operator
	dept, err := s.resolveDepartment(ctx, operator, req.CalendarType, req.DepartmentID, scope)
	if err != nil {
		return nil, err
	}
	if req.CalendarType == model.CalendarPersonal {
		if existing, err := s.rows.PersonalCalendar(ctx, owner); err != nil {
			return nil, fmt.Errorf("lookup personal calendar: %w", err)
		} else if existing != nil {
			return nil, response.NewError(response.CodeConflict, "个人日历已存在")
		}
		dept = nil
	}
	now := time.Now()
	row := &model.Calendar{
		ID: uuid.New(), Name: strings.TrimSpace(req.Name), Description: strings.TrimSpace(req.Description),
		CalendarType: req.CalendarType, Source: model.SourcePersonal, Color: defaultColor(req.Color, model.SourcePersonal), OwnerID: owner,
		DepartmentID: dept, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.rows.CreateCalendar(ctx, row); err != nil {
		return nil, fmt.Errorf("create calendar: %w", err)
	}
	return s.GetCalendar(ctx, operator, row.ID, scope)
}

func (s *scheduleService) GetCalendar(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error) {
	if feed := s.feedByCalendar(id); feed != nil {
		return feed.Calendar(viewer), nil
	}
	row, err := s.mustCalendar(ctx, id, viewer)
	if err != nil {
		return nil, err
	}
	if !canViewCalendar(scope, &row.Calendar, viewer, row.MemberRole) {
		return nil, response.NewError(response.CodeCalendarNoAccess, "无权查看该日历")
	}
	return mapCalendar(row, viewer, scope), nil
}

func (s *scheduleService) UpdateCalendar(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateCalendarRequest, scope *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error) {
	row, err := s.requireCalendarEdit(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		row.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		row.Description = strings.TrimSpace(*req.Description)
	}
	if req.Color != nil {
		row.Color = defaultColor(*req.Color, row.Source)
	}
	if req.DepartmentID != nil && row.CalendarType != model.CalendarPersonal {
		dept, err := s.resolveDepartment(ctx, operator, row.CalendarType, *req.DepartmentID, scope)
		if err != nil {
			return nil, err
		}
		row.DepartmentID = dept
	}
	row.UpdatedAt = time.Now()
	if err := s.rows.UpdateCalendar(ctx, &row.Calendar); err != nil {
		return nil, fmt.Errorf("update calendar: %w", err)
	}
	return s.GetCalendar(ctx, operator, id, scope)
}

func (s *scheduleService) DeleteCalendar(ctx context.Context, operator, id uuid.UUID, scope *rbacModel.DataScopeCondition) error {
	row, err := s.requireCalendarEdit(ctx, operator, id, scope)
	if err != nil {
		return err
	}
	if model.IsPersonalLayer(row.CalendarType, row.Source) && row.OwnerID == operator && !isAllScope(scope) {
		return response.NewError(response.CodeScheduleInvalidState, "个人日历不可删除")
	}
	if err := s.rows.DeleteCalendar(ctx, id); err != nil {
		return fmt.Errorf("delete calendar: %w", err)
	}
	return nil
}

func (s *scheduleService) ListMembers(ctx context.Context, viewer, calendarID uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.MemberResponse, error) {
	if _, err := s.requireCalendarView(ctx, viewer, calendarID, scope); err != nil {
		return nil, err
	}
	rows, err := s.rows.ListMembers(ctx, calendarID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	out := make([]dto.MemberResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapMember(&rows[i]))
	}
	return out, nil
}

func (s *scheduleService) AddMember(ctx context.Context, operator, calendarID uuid.UUID, req *dto.AddMemberRequest, scope *rbacModel.DataScopeCondition) (*dto.MemberResponse, error) {
	cal, err := s.requireCalendarEdit(ctx, operator, calendarID, scope)
	if err != nil {
		return nil, err
	}
	if cal.CalendarType == model.CalendarPersonal {
		return nil, response.NewError(response.CodeScheduleInvalidState, "个人日历不可添加成员")
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "无效的用户ID")
	}
	if uid == cal.OwnerID {
		return nil, response.NewError(response.CodeBadRequest, "所有者无需加入成员列表")
	}
	if !model.ValidMemberRole(req.Role) {
		return nil, response.NewError(response.CodeBadRequest, "成员角色不合法")
	}
	user, err := s.rows.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeBadRequest, "用户不存在")
	}
	now := time.Now()
	row := &model.CalendarMember{ID: uuid.New(), CalendarID: calendarID, UserID: uid, Role: req.Role, CreatedAt: now}
	if err := s.rows.AddMember(ctx, row); err != nil {
		return nil, fmt.Errorf("add member: %w", err)
	}
	return &dto.MemberResponse{
		ID: row.ID.String(), User: dto.Person{ID: user.ID.String(), Name: displayName(user.RealName, user.Username)},
		Role: req.Role, CreatedAt: now,
	}, nil
}

func (s *scheduleService) RemoveMember(ctx context.Context, operator, calendarID, userID uuid.UUID, scope *rbacModel.DataScopeCondition) error {
	if _, err := s.requireCalendarEdit(ctx, operator, calendarID, scope); err != nil {
		return err
	}
	if err := s.rows.RemoveMember(ctx, calendarID, userID); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

func (s *scheduleService) ensurePersonal(ctx context.Context, owner uuid.UUID) error {
	existing, err := s.rows.PersonalCalendar(ctx, owner)
	if err != nil {
		return fmt.Errorf("lookup personal calendar: %w", err)
	}
	if existing != nil {
		return nil
	}
	user, err := s.rows.GetUser(ctx, owner)
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}
	name := "我的日历"
	if user != nil && strings.TrimSpace(displayName(user.RealName, user.Username)) != "" {
		name = displayName(user.RealName, user.Username) + "的日历"
	}
	now := time.Now()
	return s.rows.CreateCalendar(ctx, &model.Calendar{
		ID: uuid.New(), Name: name, CalendarType: model.CalendarPersonal,
		Source: model.SourcePersonal, Color: model.DefaultLayerColor(model.SourcePersonal),
		OwnerID: owner, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *scheduleService) resolveDepartment(ctx context.Context, operator uuid.UUID, calType int16, raw string, scope *rbacModel.DataScopeCondition) (*uuid.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		if calType != model.CalendarDepartment {
			return nil, nil
		}
		user, err := s.rows.GetUser(ctx, operator)
		if err != nil {
			return nil, fmt.Errorf("lookup user: %w", err)
		}
		if user == nil || user.DepartmentID == nil {
			return nil, response.NewError(response.CodeBadRequest, "部门日历需要部门")
		}
		if !isAllScope(scope) && !scope.IsSelf && !deptInScope(scope, user.DepartmentID) {
			return nil, response.NewError(response.CodeCalendarNoAccess, "无权在该部门创建日历")
		}
		return user.DepartmentID, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "部门 ID 无效")
	}
	if scope != nil && scope.IsSelf {
		user, err := s.rows.GetUser(ctx, operator)
		if err != nil {
			return nil, fmt.Errorf("lookup user: %w", err)
		}
		if user == nil || user.DepartmentID == nil || *user.DepartmentID != id {
			return nil, response.NewError(response.CodeCalendarNoAccess, "无权为该部门创建日历")
		}
		return &id, nil
	}
	if !isAllScope(scope) && !deptInScope(scope, &id) {
		user, err := s.rows.GetUser(ctx, operator)
		if err != nil {
			return nil, fmt.Errorf("lookup user: %w", err)
		}
		if user == nil || user.DepartmentID == nil || *user.DepartmentID != id {
			return nil, response.NewError(response.CodeCalendarNoAccess, "无权为该部门创建日历")
		}
	}
	return &id, nil
}

func (s *scheduleService) mustCalendar(ctx context.Context, id, viewer uuid.UUID) (*model.CalendarNamed, error) {
	row, err := s.rows.GetCalendarNamed(ctx, id, viewer)
	if err != nil {
		return nil, fmt.Errorf("get calendar: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeCalendarNotFound, "日历不存在")
	}
	return row, nil
}

func (s *scheduleService) requireCalendarView(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*model.CalendarNamed, error) {
	if feed := s.feedByCalendar(id); feed != nil {
		cal := feed.Calendar(viewer)
		return &model.CalendarNamed{
			Calendar: model.Calendar{
				ID: id, Name: cal.Name, CalendarType: cal.CalendarType,
				Source: cal.Source, Color: cal.Color, OwnerID: viewer,
			},
			MemberRole: model.MemberViewer,
		}, nil
	}
	row, err := s.mustCalendar(ctx, id, viewer)
	if err != nil {
		return nil, err
	}
	if !canViewCalendar(scope, &row.Calendar, viewer, row.MemberRole) {
		return nil, response.NewError(response.CodeCalendarNoAccess, "无权查看该日历")
	}
	return row, nil
}

func (s *scheduleService) requireCalendarEdit(ctx context.Context, operator, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*model.CalendarNamed, error) {
	if s.feedByCalendar(id) != nil {
		return nil, response.NewError(response.CodeScheduleInvalidState, "系统图层只读，不可管理")
	}
	row, err := s.mustCalendar(ctx, id, operator)
	if err != nil {
		return nil, err
	}
	if !canEditCalendar(scope, &row.Calendar, operator, row.MemberRole) {
		return nil, response.NewError(response.CodeCalendarNoAccess, "无权管理该日历")
	}
	return row, nil
}
