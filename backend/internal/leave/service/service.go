package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	"github.com/google/uuid"
)

type Viewer struct {
	UserID     uuid.UUID
	CanRead    bool
	CanApprove bool
}

type Service interface {
	ListTypes(ctx context.Context) ([]dto.LeaveTypeResponse, error)
	Submit(ctx context.Context, viewer Viewer, req *dto.SubmitLeaveRequest) (*dto.LeaveApplicationResponse, error)
	Get(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.LeaveApplicationResponse, error)
	ListMine(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error)
	ListAll(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error)
	Approve(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error
	Reject(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error
	Balances(ctx context.Context, viewer Viewer, userID string, year int) ([]*dto.LeaveBalanceResponse, error)
	Stats(ctx context.Context, viewer Viewer) (*dto.LeaveStatsResponse, error)
}

type leaveService struct {
	rows repo.Repository
	now  func() time.Time
}

func New(rows repo.Repository) Service {
	return &leaveService{rows: rows, now: time.Now}
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
	if row.ApplicantID != viewer.UserID && !viewer.CanRead {
		return nil, noAccess("无权查看该请假申请")
	}
	return mapApplication(row), nil
}

func (s *leaveService) ListMine(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error) {
	page, size := defaultPage(req.Page, req.PageSize)
	rows, total, err := s.rows.GetLeaveApplicationsByUser(ctx, viewer.UserID, req.Status, page, size)
	if err != nil {
		return nil, 0, err
	}
	return mapApps(rows), total, nil
}

func (s *leaveService) ListAll(ctx context.Context, viewer Viewer, req *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error) {
	if !viewer.CanRead {
		return nil, 0, noAccess("无权查看全部请假记录")
	}
	page, size := defaultPage(req.Page, req.PageSize)
	if req.UserID != "" {
		uid, err := uuid.Parse(req.UserID)
		if err != nil {
			return nil, 0, invalidTime("用户ID格式错误")
		}
		rows, total, err := s.rows.GetLeaveApplicationsByUser(ctx, uid, req.Status, page, size)
		if err != nil {
			return nil, 0, err
		}
		return mapApps(rows), total, nil
	}
	rows, total, err := s.rows.GetLeaveApplicationsByStatus(ctx, req.Status, page, size)
	if err != nil {
		return nil, 0, err
	}
	return mapApps(rows), total, nil
}

func (s *leaveService) Stats(ctx context.Context, viewer Viewer) (*dto.LeaveStatsResponse, error) {
	if !viewer.CanRead {
		return nil, noAccess("无权查看请假统计")
	}
	total, byStatus, byType, err := s.rows.CountStats(ctx)
	if err != nil {
		return nil, err
	}
	types := make([]dto.TypeStat, 0, len(byType))
	for _, row := range byType {
		types = append(types, dto.TypeStat{
			LeaveTypeID: row.LeaveTypeID.String(),
			Code:        row.Code,
			Name:        row.Name,
			Count:       row.Count,
			Days:        row.Days,
		})
	}
	return &dto.LeaveStatsResponse{Total: total, ByStatus: byStatus, ByType: types}, nil
}

func mapApps(rows []model.ApplicationNamed) []*dto.LeaveApplicationResponse {
	out := make([]*dto.LeaveApplicationResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapApplication(&rows[i]))
	}
	return out
}
