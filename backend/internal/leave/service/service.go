package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

type Viewer struct {
	UserID     uuid.UUID
	CanRead    bool
	CanApprove bool
	Scope      *rbacModel.DataScopeCondition
}

type Service interface {
	ListTypes(ctx context.Context) ([]dto.LeaveTypeResponse, error)
	CreateType(ctx context.Context, viewer Viewer, req *dto.UpsertLeaveTypeRequest) (*dto.LeaveTypeResponse, error)
	UpdateType(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpsertLeaveTypeRequest) (*dto.LeaveTypeResponse, error)
	Submit(ctx context.Context, viewer Viewer, req *dto.SubmitLeaveRequest) (*dto.LeaveApplicationResponse, error)
	Get(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.LeaveApplicationResponse, error)
	ListMine(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error)
	ListAll(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error)
	ListTodos(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error)
	Calendar(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, error)
	Approve(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error
	Reject(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error
	Balances(ctx context.Context, viewer Viewer, userID string, year int) ([]*dto.LeaveBalanceResponse, error)
	Stats(ctx context.Context, viewer Viewer, year int) (*dto.LeaveStatsResponse, error)
}

type leaveService struct {
	rows   repo.Repository
	engine LeaveEngine
	now    func() time.Time
}

func New(rows repo.Repository, engine LeaveEngine) Service {
	return &leaveService{rows: rows, engine: engine, now: time.Now}
}

func (s *leaveService) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func defaultPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (s *leaveService) ListTypes(ctx context.Context) ([]dto.LeaveTypeResponse, error) {
	rows, err := s.rows.GetAllLeaveTypes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.LeaveTypeResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapType(row))
	}
	return out, nil
}

func (s *leaveService) Get(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.LeaveApplicationResponse, error) {
	row, err := s.rows.GetLeaveApplicationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, notFound()
	}
	if row.ApplicantID == viewer.UserID {
		return mapApplication(row), nil
	}
	if !viewer.CanRead || !canAccessApplicant(viewer.Scope, row.ApplicantID, row.ApplicantDepartmentID, viewer.UserID) {
		return nil, noAccess("无权查看该请假申请")
	}
	return mapApplication(row), nil
}

func (s *leaveService) ListMine(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error) {
	page, size := defaultPage(req.Page, req.PageSize)
	rows, total, err := s.rows.GetLeaveApplicationsByUser(ctx, viewer.UserID, req.Status, page, size, nil)
	if err != nil {
		return nil, 0, err
	}
	return mapApps(rows), total, nil
}

func (s *leaveService) ListAll(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error) {
	if !viewer.CanRead {
		return nil, 0, noAccess("无权查看全部请假记录")
	}
	if viewer.Scope == nil {
		return nil, 0, noAccess("无权查看全部请假记录")
	}
	page, size := defaultPage(req.Page, req.PageSize)
	sqlScope := rewriteApplicantScope(viewer.Scope, viewer.UserID)
	if req.UserID != "" {
		uid, err := uuid.Parse(req.UserID)
		if err != nil {
			return nil, 0, invalidTime("用户ID格式错误")
		}
		dept, err := s.rows.GetUserDepartmentID(ctx, uid)
		if err != nil {
			return nil, 0, err
		}
		if !canAccessApplicant(viewer.Scope, uid, dept, viewer.UserID) {
			return nil, 0, noAccess("无权查看该用户的请假记录")
		}
		rows, total, err := s.rows.GetLeaveApplicationsByUser(ctx, uid, req.Status, page, size, sqlScope)
		if err != nil {
			return nil, 0, err
		}
		return mapApps(rows), total, nil
	}
	rows, total, err := s.rows.GetLeaveApplicationsByStatus(ctx, req.Status, page, size, sqlScope)
	if err != nil {
		return nil, 0, err
	}
	return mapApps(rows), total, nil
}

func (s *leaveService) ListTodos(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error) {
	if !viewer.CanApprove {
		return nil, 0, noAccess("无权查看请假待办")
	}
	if req == nil {
		req = &dto.ListLeaveRequest{}
	}
	req.Status = model.ApprovalStatusPending
	return s.ListAll(ctx, viewer, req)
}

func (s *leaveService) Calendar(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, error) {
	from, to, err := parseCalendarRange(req.From, req.To, s.clock())
	if err != nil {
		return nil, err
	}
	var userID *uuid.UUID
	var scope *rbacModel.DataScopeCondition
	if viewer.CanRead && viewer.Scope != nil {
		scope = rewriteApplicantScope(viewer.Scope, viewer.UserID)
		if req.UserID != "" {
			uid, err := uuid.Parse(req.UserID)
			if err != nil {
				return nil, invalidTime("用户ID格式错误")
			}
			dept, err := s.rows.GetUserDepartmentID(ctx, uid)
			if err != nil {
				return nil, err
			}
			if !canAccessApplicant(viewer.Scope, uid, dept, viewer.UserID) {
				return nil, noAccess("无权查看该用户的请假日历")
			}
			userID = &uid
		}
	} else {
		id := viewer.UserID
		userID = &id
	}
	rows, err := s.rows.ListCalendar(ctx, from, to, userID, scope)
	if err != nil {
		return nil, err
	}
	return mapApps(rows), nil
}

func (s *leaveService) Stats(ctx context.Context, viewer Viewer, year int) (*dto.LeaveStatsResponse, error) {
	if !viewer.CanRead {
		return nil, noAccess("无权查看请假统计")
	}
	if viewer.Scope == nil {
		return nil, noAccess("无权查看请假统计")
	}
	sqlScope := rewriteApplicantScope(viewer.Scope, viewer.UserID)
	// year==0：待办页不传年份，合计须覆盖全部历史，与列表分页一致。
	// year>0：统计页按所选年份过滤。
	total, byStatus, byType, err := s.rows.CountStats(ctx, year, sqlScope)
	if err != nil {
		return nil, err
	}
	personalTotal, personalDays, personalTypes, err := s.rows.CountPersonalStats(ctx, viewer.UserID, year)
	if err != nil {
		return nil, err
	}
	depts, err := s.rows.CountDepartmentStats(ctx, year, sqlScope)
	if err != nil {
		return nil, err
	}
	months, err := s.rows.CountMonthlyStats(ctx, year, sqlScope)
	if err != nil {
		return nil, err
	}
	deptOut := make([]dto.DepartmentStat, 0, len(depts))
	for _, row := range depts {
		item := dto.DepartmentStat{DepartmentName: row.DepartmentName, Count: row.Count, Days: row.Days}
		if row.DepartmentID != nil {
			item.DepartmentID = row.DepartmentID.String()
		}
		deptOut = append(deptOut, item)
	}
	monthOut := make([]dto.MonthStat, 0, len(months))
	for _, row := range months {
		monthOut = append(monthOut, dto.MonthStat{Month: row.Month, Count: row.Count, Days: row.Days})
	}
	return &dto.LeaveStatsResponse{
		Total: total, ByStatus: byStatus, ByType: mapTypeStats(byType),
		Personal:    dto.PersonalStat{Total: personalTotal, Days: personalDays, ByType: mapTypeStats(personalTypes)},
		Departments: deptOut,
		ByMonth:     monthOut,
	}, nil
}

func mapTypeStats(rows []repo.TypeCount) []dto.TypeStat {
	types := make([]dto.TypeStat, 0, len(rows))
	for _, row := range rows {
		types = append(types, dto.TypeStat{
			LeaveTypeID: row.LeaveTypeID.String(),
			Code:        row.Code,
			Name:        row.Name,
			Count:       row.Count,
			Days:        row.Days,
		})
	}
	return types
}

func mapApps(rows []model.ApplicationNamed) []*dto.LeaveApplicationResponse {
	out := make([]*dto.LeaveApplicationResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapApplication(&rows[i]))
	}
	return out
}
