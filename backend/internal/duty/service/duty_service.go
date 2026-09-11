package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/duty/dto"
	"github.com/Yogdunana/StarByte/backend/internal/duty/model"
	"github.com/Yogdunana/StarByte/backend/internal/duty/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notifier 通知接口（由外部注入，避免直接依赖 notification 模块）
type Notifier interface {
	NotifyDutyReminder(ctx context.Context, userID uuid.UUID, dutyDate time.Time, timeSlot, location string) error
	NotifySwapResult(ctx context.Context, userID uuid.UUID, approved bool, remark string) error
}

// DutyService 值班服务接口
type DutyService interface {
	ListSchedules(ctx context.Context, req *dto.ListScheduleRequest) ([]dto.ScheduleResponse, int64, error)
	CreateSchedule(ctx context.Context, req *dto.CreateScheduleRequest, creatorID string) (*dto.ScheduleResponse, error)
	BatchCreateSchedules(ctx context.Context, req *dto.BatchCreateScheduleRequest, creatorID string) (int, error)
	UpdateSchedule(ctx context.Context, id uuid.UUID, req *dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error)
	CreateSwapRequest(ctx context.Context, req *dto.SwapRequestDTO, requesterID string) (*dto.SwapResponse, error)
	ActionSwap(ctx context.Context, id uuid.UUID, req *dto.SwapActionRequest, approverID string) (*dto.SwapResponse, error)
	ListSwapRequests(ctx context.Context, page, pageSize int, requesterID string, status *int) ([]dto.SwapResponse, int64, error)
	GetDutyStats(ctx context.Context, req *dto.DutyStatsRequest) (*dto.DutyStatsResponse, error)
}

type dutyService struct {
	db          *gorm.DB
	scheduleRepo repo.ScheduleRepo
	swapRepo    repo.SwapRepo
	notifier    Notifier
}

func NewDutyService(db *gorm.DB, scheduleRepo repo.ScheduleRepo, swapRepo repo.SwapRepo, notifier Notifier) DutyService {
	return &dutyService{
		db:          db,
		scheduleRepo: scheduleRepo,
		swapRepo:    swapRepo,
		notifier:    notifier,
	}
}

func (s *dutyService) ListSchedules(ctx context.Context, req *dto.ListScheduleRequest) ([]dto.ScheduleResponse, int64, error) {
	var deptID uuid.UUID
	if req.DepartmentID != "" {
		parsed, err := parseRequiredUUID(req.DepartmentID, "部门ID")
		if err != nil {
			return nil, 0, err
		}
		deptID = parsed
	}

	var uid uuid.UUID
	if req.UserID != "" {
		parsed, err := parseRequiredUUID(req.UserID, "用户ID")
		if err != nil {
			return nil, 0, err
		}
		uid = parsed
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, 0, response.NewError(response.CodeBadRequest, "start_date 格式无效")
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, 0, response.NewError(response.CodeBadRequest, "end_date 格式无效")
		}
		endDate = &t
	}

	// 根据 view 类型自动计算日期范围
	if req.View != "" {
		now := time.Now()
		switch req.View {
		case "week":
			if startDate == nil {
				start := now.AddDate(0, 0, -int(now.Weekday())+1)
				startDate = &start
				end := start.AddDate(0, 0, 6)
				endDate = &end
			}
		case "month":
			if startDate == nil {
				start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
				startDate = &start
				end := start.AddDate(0, 1, -1)
				endDate = &end
			}
		case "day":
			if startDate == nil {
				start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
				startDate = &start
				endDate = &start
			}
		}
	}

	schedules, total, err := s.scheduleRepo.List(ctx, req.Page, req.PageSize, deptID, uid, startDate, endDate)
	if err != nil {
		return nil, 0, fmt.Errorf("list schedules: %w", err)
	}

	result := make([]dto.ScheduleResponse, 0, len(schedules))
	for _, sc := range schedules {
		item := dto.ScheduleResponse{
			ID:            sc.ID.String(),
			UserID:        sc.UserID.String(),
			DutyDate:      sc.DutyDate.Format("2006-01-02"),
			TimeSlot:      sc.TimeSlot,
			Location:      sc.Location,
			Remark:        sc.Remark,
			Status:        sc.Status,
			StatusText:    dto.GetScheduleStatusText(sc.Status),
			CreatedAt:     sc.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     sc.UpdatedAt.Format(time.RFC3339),
		}
		if sc.DepartmentID != nil {
			item.DepartmentID = sc.DepartmentID.String()
		}
		result = append(result, item)
	}

	return result, total, nil
}

func (s *dutyService) CreateSchedule(ctx context.Context, req *dto.CreateScheduleRequest, creatorID string) (*dto.ScheduleResponse, error) {
	uid, err := parseRequiredUUID(req.UserID, "用户ID")
	if err != nil {
		return nil, err
	}

	creatorUUID, _ := uuid.Parse(creatorID)

	dutyDate, err := time.Parse("2006-01-02", req.DutyDate)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "duty_date 格式无效，应为 YYYY-MM-DD")
	}

	// 检查冲突
	count, err := s.scheduleRepo.CountByUserAndDate(ctx, uid, dutyDate, req.TimeSlot)
	if err != nil {
		return nil, fmt.Errorf("check conflict: %w", err)
	}
	if count > 0 {
		return nil, response.NewError(26001, "该用户在此时段已有排班")
	}

	var deptID *uuid.UUID
	if req.DepartmentID != "" {
		parsed, err := parseOptionalUUID(req.DepartmentID, "部门ID")
		if err != nil {
			return nil, err
		}
		deptID = parsed
	}

	schedule := &model.Schedule{
		ID:           uuid.New(),
		UserID:       uid,
		DepartmentID: deptID,
		DutyDate:     dutyDate,
		TimeSlot:     req.TimeSlot,
		Location:     req.Location,
		Remark:       req.Remark,
		Status:       model.ScheduleStatusPending,
		CreatedBy:    &creatorUUID,
	}

	err = s.scheduleRepo.Create(ctx, nil, schedule)
	if err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}

	return s.scheduleToResponse(schedule), nil
}

func (s *dutyService) BatchCreateSchedules(ctx context.Context, req *dto.BatchCreateScheduleRequest, creatorID string) (int, error) {
	creatorUUID, _ := uuid.Parse(creatorID)
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return 0, response.NewError(response.CodeBadRequest, "start_date 格式无效")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return 0, response.NewError(response.CodeBadRequest, "end_date 格式无效")
	}
	if endDate.Before(startDate) {
		return 0, response.NewError(response.CodeBadRequest, "结束日期不能早于开始日期")
	}

	var deptID *uuid.UUID
	if req.DepartmentID != "" {
		parsed, err := parseOptionalUUID(req.DepartmentID, "部门ID")
		if err != nil {
			return 0, err
		}
		deptID = parsed
	}

	// 解析用户IDs
	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, uidStr := range req.UserIDs {
		uid, err := parseRequiredUUID(uidStr, "用户ID")
		if err != nil {
			return 0, err
		}
		userIDs = append(userIDs, uid)
	}

	created := 0
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 遍历日期范围
		current := startDate
		userIdx := 0
		for !current.After(endDate) {
			var assignIdx int
			if req.Rotation {
				assignIdx = userIdx % len(userIDs)
				userIdx++
			} else {
				assignIdx = 0 // 默认都排第一个？不，不轮换时每人每天都排
			}

			if req.Rotation {
				schedule := &model.Schedule{
					ID:           uuid.New(),
					UserID:       userIDs[assignIdx],
					DepartmentID: deptID,
					DutyDate:     current,
					TimeSlot:     req.TimeSlot,
					Location:     req.Location,
					Status:       model.ScheduleStatusPending,
					CreatedBy:    &creatorUUID,
				}
				if err := s.scheduleRepo.Create(ctx, tx, schedule); err != nil {
					return err
				}
				created++
			} else {
				for _, uid := range userIDs {
					schedule := &model.Schedule{
						ID:           uuid.New(),
						UserID:       uid,
						DepartmentID: deptID,
						DutyDate:     current,
						TimeSlot:     req.TimeSlot,
						Location:     req.Location,
						Status:       model.ScheduleStatusPending,
						CreatedBy:    &creatorUUID,
					}
					if err := s.scheduleRepo.Create(ctx, tx, schedule); err != nil {
						return err
					}
					created++
				}
			}
			current = current.AddDate(0, 0, 1)
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("batch create: %w", err)
	}
	return created, nil
}

func (s *dutyService) UpdateSchedule(ctx context.Context, id uuid.UUID, req *dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error) {
	schedule, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	if schedule == nil {
		return nil, response.NewError(26002, "排班记录不存在")
	}

	if req.UserID != "" {
		uid, err := parseRequiredUUID(req.UserID, "用户ID")
		if err != nil {
			return nil, err
		}
		schedule.UserID = uid
	}
	if req.DutyDate != "" {
		dutyDate, err := time.Parse("2006-01-02", req.DutyDate)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "duty_date 格式无效")
		}
		schedule.DutyDate = dutyDate
	}
	if req.TimeSlot != "" {
		schedule.TimeSlot = req.TimeSlot
	}
	if req.Location != "" {
		schedule.Location = req.Location
	}
	if req.Remark != "" {
		schedule.Remark = req.Remark
	}
	if req.Status != nil {
		schedule.Status = *req.Status
	}

	err = s.scheduleRepo.Update(ctx, nil, schedule)
	if err != nil {
		return nil, fmt.Errorf("update schedule: %w", err)
	}

	return s.scheduleToResponse(schedule), nil
}

func (s *dutyService) CreateSwapRequest(ctx context.Context, req *dto.SwapRequestDTO, requesterID string) (*dto.SwapResponse, error) {
	rID, err := parseRequiredUUID(requesterID, "申请人ID")
	if err != nil {
		return nil, err
	}

	targetID, err := parseRequiredUUID(req.TargetUserID, "目标用户ID")
	if err != nil {
		return nil, err
	}

	scheduleID, err := parseRequiredUUID(req.RequesterScheduleID, "排班ID")
	if err != nil {
		return nil, err
	}

	// 验证排班存在
	schedule, err := s.scheduleRepo.GetByID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	if schedule == nil {
		return nil, response.NewError(26002, "排班记录不存在")
	}
	if schedule.UserID != rID {
		return nil, response.NewError(26003, "无权申请调班：排班不属于申请人")
	}

	var targetScheduleID *uuid.UUID
	if req.TargetScheduleID != "" {
		parsed, err := parseOptionalUUID(req.TargetScheduleID, "目标排班ID")
		if err != nil {
			return nil, err
		}
		targetScheduleID = parsed
	}

	swap := &model.SwapRequest{
		ID:                  uuid.New(),
		RequesterID:         rID,
		TargetUserID:        targetID,
		RequesterScheduleID: scheduleID,
		TargetScheduleID:    targetScheduleID,
		Reason:              req.Reason,
		Status:              model.SwapStatusPending,
	}

	err = s.swapRepo.Create(ctx, nil, swap)
	if err != nil {
		return nil, fmt.Errorf("create swap: %w", err)
	}

	return s.swapToResponse(swap), nil
}

func (s *dutyService) ActionSwap(ctx context.Context, id uuid.UUID, req *dto.SwapActionRequest, approverID string) (*dto.SwapResponse, error) {
	swap, err := s.swapRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get swap: %w", err)
	}
	if swap == nil {
		return nil, response.NewError(26004, "调班申请不存在")
	}
	if swap.Status != model.SwapStatusPending {
		return nil, response.NewError(26005, "调班申请已处理，不可重复操作")
	}

	approverUUID, _ := uuid.Parse(approverID)
	now := time.Now()
	swap.Status = req.Status
	swap.ApproverID = &approverUUID
	swap.ApprovedAt = &now
	swap.ApprovedRemark = req.Remark

	// 如果批准，交换排班的用户
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.swapRepo.Update(ctx, tx, swap); err != nil {
			return err
		}

		if req.Status == model.SwapStatusApproved {
			// 获取申请人和目标用户的排班
			reqSchedule, err := s.scheduleRepo.GetByID(ctx, swap.RequesterScheduleID)
			if err != nil {
				return err
			}
			if reqSchedule == nil {
				return response.NewError(26002, "排班记录不存在")
			}

			// 交换排班的用户ID
			originalUser := reqSchedule.UserID
			reqSchedule.UserID = swap.TargetUserID
			reqSchedule.Status = model.ScheduleStatusSwapped
			if err := s.scheduleRepo.Update(ctx, tx, reqSchedule); err != nil {
				return err
			}

			// 如果有目标排班也交换
			if swap.TargetScheduleID != nil {
				targetSchedule, err := s.scheduleRepo.GetByID(ctx, *swap.TargetScheduleID)
				if err != nil {
					return err
				}
				if targetSchedule != nil {
					targetSchedule.UserID = originalUser
					targetSchedule.Status = model.ScheduleStatusSwapped
					if err := s.scheduleRepo.Update(ctx, tx, targetSchedule); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("action swap: %w", err)
	}

	// 异步通知
	if s.notifier != nil {
		go func() {
			_ = s.notifier.NotifySwapResult(context.Background(), swap.RequesterID, req.Status == model.SwapStatusApproved, req.Remark)
			_ = s.notifier.NotifySwapResult(context.Background(), swap.TargetUserID, req.Status == model.SwapStatusApproved, req.Remark)
		}()
	}

	return s.swapToResponse(swap), nil
}

func (s *dutyService) ListSwapRequests(ctx context.Context, page, pageSize int, requesterID string, status *int) ([]dto.SwapResponse, int64, error) {
	var rID uuid.UUID
	if requesterID != "" {
		parsed, err := parseRequiredUUID(requesterID, "申请人ID")
		if err != nil {
			return nil, 0, err
		}
		rID = parsed
	}

	swaps, total, err := s.swapRepo.List(ctx, page, pageSize, rID, status)
	if err != nil {
		return nil, 0, fmt.Errorf("list swaps: %w", err)
	}

	result := make([]dto.SwapResponse, 0, len(swaps))
	for _, sr := range swaps {
		result = append(result, *s.swapToResponse(&sr))
	}
	return result, total, nil
}

func (s *dutyService) GetDutyStats(ctx context.Context, req *dto.DutyStatsRequest) (*dto.DutyStatsResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "start_date 格式无效")
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "end_date 格式无效")
		}
		endDate = &t
	}

	var stats []repo.UserDutyStat
	if req.DepartmentID != "" {
		deptID, err := parseRequiredUUID(req.DepartmentID, "部门ID")
		if err != nil {
			return nil, err
		}
		stats, err = s.scheduleRepo.StatsByDepartment(ctx, deptID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("stats by department: %w", err)
		}
	} else if req.UserID != "" {
		uid, err := parseRequiredUUID(req.UserID, "用户ID")
		if err != nil {
			return nil, err
		}
		stats, err = s.scheduleRepo.StatsByUser(ctx, uid, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("stats by user: %w", err)
		}
	} else {
		// 全局统计
		deptID := uuid.Nil
		stats, _ = s.scheduleRepo.StatsByDepartment(ctx, deptID, startDate, endDate)
	}

	resp := &dto.DutyStatsResponse{}
	userStats := make([]dto.UserDutyStatItem, 0, len(stats))
	for _, st := range stats {
		var rate float64
		if st.TotalCount > 0 {
			rate = float64(st.Completed) / float64(st.TotalCount) * 100
		}
		resp.TotalDuties += st.TotalCount
		resp.CompletedDuties += st.Completed
		resp.AbsentDuties += st.Absent
		resp.PendingDuties += st.TotalCount - st.Completed - st.Absent
		userStats = append(userStats, dto.UserDutyStatItem{
			UserID:     st.UserID.String(),
			TotalCount: st.TotalCount,
			Completed:  st.Completed,
			Absent:     st.Absent,
			Rate:       rate,
		})
	}
	if resp.TotalDuties > 0 {
		resp.CompletionRate = float64(resp.CompletedDuties) / float64(resp.TotalDuties) * 100
	}
	resp.UserStats = userStats
	return resp, nil
}

// ========== 工具函数 ==========

func (s *dutyService) scheduleToResponse(sc *model.Schedule) *dto.ScheduleResponse {
	resp := &dto.ScheduleResponse{
		ID:         sc.ID.String(),
		UserID:     sc.UserID.String(),
		DutyDate:   sc.DutyDate.Format("2006-01-02"),
		TimeSlot:   sc.TimeSlot,
		Location:   sc.Location,
		Remark:     sc.Remark,
		Status:     sc.Status,
		StatusText: dto.GetScheduleStatusText(sc.Status),
		CreatedAt:  sc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  sc.UpdatedAt.Format(time.RFC3339),
	}
	if sc.DepartmentID != nil {
		resp.DepartmentID = sc.DepartmentID.String()
	}
	return resp
}

func (s *dutyService) swapToResponse(sr *model.SwapRequest) *dto.SwapResponse {
	resp := &dto.SwapResponse{
		ID:                  sr.ID.String(),
		RequesterID:         sr.RequesterID.String(),
		TargetUserID:       sr.TargetUserID.String(),
		RequesterScheduleID: sr.RequesterScheduleID.String(),
		Reason:             sr.Reason,
		Status:             sr.Status,
		StatusText:         dto.GetSwapStatusText(sr.Status),
		CreatedAt:          sr.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          sr.UpdatedAt.Format(time.RFC3339),
	}
	if sr.TargetScheduleID != nil {
		resp.TargetScheduleID = sr.TargetScheduleID.String()
	}
	if sr.ApproverID != nil {
		resp.ApproverID = sr.ApproverID.String()
	}
	if sr.ApprovedAt != nil {
		resp.ApprovedAt = sr.ApprovedAt.Format(time.RFC3339)
	}
	resp.ApprovedRemark = sr.ApprovedRemark
	return resp
}

func parseRequiredUUID(value, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "参数错误: 无效的"+field)
	}
	return id, nil
}

func parseOptionalUUID(value, field string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := parseRequiredUUID(value, field)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
