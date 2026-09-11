package service

import (
	"errors"
	"leave-backend/internal/leave/model"
	"leave-backend/internal/leave/repo"
	"math"
	"time"

	"gorm.io/gorm"
)

// LeaveService 请假模块业务逻辑层
type LeaveService struct {
	repo *repo.LeaveRepository
}

// NewLeaveService 创建服务实例
func NewLeaveService(repo *repo.LeaveRepository) *LeaveService {
	return &LeaveService{repo: repo}
}

// SubmitLeaveRequest 提交请假申请
// 业务流程：校验请假类型 → 校验时间 → 校验余额 → 创建申请 → 扣减余额
func (s *LeaveService) SubmitLeaveRequest(applicantID uint, leaveTypeID uint, startTime, endTime time.Time, reason string) (*model.LeaveApplication, error) {
	// 1. 校验请假类型是否存在
	leaveType, err := s.repo.GetLeaveTypeByID(leaveTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("请假类型不存在")
		}
		return nil, err
	}

	// 2. 校验时间合法性
	if startTime.After(endTime) {
		return nil, errors.New("开始时间不能晚于结束时间")
	}
	if startTime.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, errors.New("开始时间不能早于今天")
	}

	// 3. 计算请假时长（天），按工作日计算（排除周末）
	durationDays := s.calculateDurationDays(startTime, endTime)
	if durationDays <= 0 {
		return nil, errors.New("请假时长必须大于0")
	}

	// 4. 如果该请假类型需要扣减余额，校验并扣减
	year := startTime.Year()
	if leaveType.Deductible {
		balance, err := s.repo.GetLeaveBalance(applicantID, year, leaveTypeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("未找到该类型的假期余额")
			}
			return nil, err
		}
		if balance.RemainingDays < durationDays {
			return nil, errors.New("假期余额不足")
		}
		// 扣减余额
		if err := s.repo.DeductLeaveBalance(applicantID, year, leaveTypeID, durationDays); err != nil {
			return nil, err
		}
	}

	// 5. 创建请假申请
	application := &model.LeaveApplication{
		ApplicantID:  applicantID,
		LeaveTypeID:  leaveTypeID,
		StartTime:    startTime,
		EndTime:      endTime,
		DurationDays: durationDays,
		Reason:       reason,
		Status:       model.ApprovalStatusPending,
	}

	if err := s.repo.CreateLeaveApplication(application); err != nil {
		// 创建失败，返还已扣减的余额
		if leaveType.Deductible {
			_ = s.repo.DeductLeaveBalance(applicantID, year, leaveTypeID, -durationDays)
		}
		return nil, err
	}

	// 重新查询以获取关联的请假类型信息
	return s.repo.GetLeaveApplicationByID(application.ID)
}

// ApproveLeave 审批请假申请（通过）
func (s *LeaveService) ApproveLeave(applicationID uint, approverID uint, remark string) error {
	application, err := s.repo.GetLeaveApplicationByID(applicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("请假申请不存在")
		}
		return err
	}

	if application.Status != model.ApprovalStatusPending {
		return errors.New("该请假申请已审批，不可重复审批")
	}

	return s.repo.UpdateApprovalStatus(applicationID, approverID, model.ApprovalStatusApproved, remark)
}

// RejectLeave 驳回请假申请
// 驳回时需要返还已扣减的假期余额
func (s *LeaveService) RejectLeave(applicationID uint, approverID uint, remark string) error {
	application, err := s.repo.GetLeaveApplicationByID(applicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("请假申请不存在")
		}
		return err
	}

	if application.Status != model.ApprovalStatusPending {
		return errors.New("该请假申请已审批，不可重复审批")
	}

	// 驳回申请
	if err := s.repo.UpdateApprovalStatus(applicationID, approverID, model.ApprovalStatusRejected, remark); err != nil {
		return err
	}

	// 如果该请假类型扣减了余额，需要返还
	leaveType, err := s.repo.GetLeaveTypeByID(application.LeaveTypeID)
	if err != nil {
		return nil
	}
	if leaveType.Deductible {
		year := application.StartTime.Year()
		_ = s.repo.DeductLeaveBalance(application.ApplicantID, year, application.LeaveTypeID, -application.DurationDays)
	}

	return nil
}

// GetLeaveApplicationByID 根据ID获取请假申请详情
func (s *LeaveService) GetLeaveApplicationByID(id uint) (*model.LeaveApplication, error) {
	application, err := s.repo.GetLeaveApplicationByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("请假申请不存在")
		}
		return nil, err
	}
	return application, nil
}

// GetMyLeaveRecords 查询我的请假记录（分页）
func (s *LeaveService) GetMyLeaveRecords(userID uint, page, pageSize int) ([]model.LeaveApplication, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.repo.GetLeaveApplicationsByUser(userID, page, pageSize)
}

// GetLeaveApplicationsByStatus 按状态查询请假申请（分页）
func (s *LeaveService) GetLeaveApplicationsByStatus(status string, page, pageSize int) ([]model.LeaveApplication, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.repo.GetLeaveApplicationsByStatus(status, page, pageSize)
}

// GetAvailableLeaveTypes 获取所有可用请假类型
func (s *LeaveService) GetAvailableLeaveTypes() ([]model.LeaveType, error) {
	return s.repo.GetAllLeaveTypes()
}

// GetMyLeaveBalances 获取用户某年所有假期余额
func (s *LeaveService) GetMyLeaveBalances(userID uint, year int) ([]model.LeaveBalance, error) {
	if year == 0 {
		year = time.Now().Year()
	}
	return s.repo.GetLeaveBalancesByUser(userID, year)
}

// calculateDurationDays 计算请假时长（天）
// 按工作日计算（排除周六周日），不足一天按半天/一天计算
func (s *LeaveService) calculateDurationDays(startTime, endTime time.Time) float64 {
	// 如果同一天，按一天计算
	if startTime.Truncate(24*time.Hour).Equal(endTime.Truncate(24*time.Hour)) {
		return 1
	}

	totalDays := 0.0
	current := startTime.Truncate(24 * time.Hour)
	end := endTime.Truncate(24 * time.Hour)

	for current.Before(end) || current.Equal(end) {
		weekday := current.Weekday()
		// 排除周末
		if weekday != time.Saturday && weekday != time.Sunday {
			totalDays++
		}
		current = current.Add(24 * time.Hour)
	}

	// 如果结束时间不是整天结束，检查是否需要按半天计算
	if endTime.Sub(current.Add(-24*time.Hour)) < 12*time.Hour && totalDays > 0 {
		// 不足12小时，按半天处理
		totalDays = totalDays - 0.5
	}

	// 确保至少返回1天
	return math.Max(totalDays, 1)
}
