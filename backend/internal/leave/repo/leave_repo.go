package repo

import (
	"leave-backend/internal/leave/model"
	"time"

	"gorm.io/gorm"
)

// LeaveRepository 请假模块数据访问层
// 封装所有请假相关的数据库操作
type LeaveRepository struct {
	db *gorm.DB
}

// NewLeaveRepository 创建仓库实例
func NewLeaveRepository(db *gorm.DB) *LeaveRepository {
	return &LeaveRepository{db: db}
}

// ===================== 请假类型操作 =====================

// GetAllLeaveTypes 获取所有请假类型
func (r *LeaveRepository) GetAllLeaveTypes() ([]model.LeaveType, error) {
	var leaveTypes []model.LeaveType
	if err := r.db.Find(&leaveTypes).Error; err != nil {
		return nil, err
	}
	return leaveTypes, nil
}

// GetLeaveTypeByID 根据ID获取请假类型
func (r *LeaveRepository) GetLeaveTypeByID(id uint) (*model.LeaveType, error) {
	var leaveType model.LeaveType
	if err := r.db.First(&leaveType, id).Error; err != nil {
		return nil, err
	}
	return &leaveType, nil
}

// GetLeaveTypeByCode 根据编码获取请假类型
func (r *LeaveRepository) GetLeaveTypeByCode(code string) (*model.LeaveType, error) {
	var leaveType model.LeaveType
	if err := r.db.Where("code = ?", code).First(&leaveType).Error; err != nil {
		return nil, err
	}
	return &leaveType, nil
}

// CreateLeaveType 创建请假类型
func (r *LeaveRepository) CreateLeaveType(leaveType *model.LeaveType) error {
	return r.db.Create(leaveType).Error
}

// ===================== 假期余额操作 =====================

// GetLeaveBalance 查询用户假期余额
// 按用户ID、年份、请假类型ID查询
func (r *LeaveRepository) GetLeaveBalance(userID uint, year int, leaveTypeID uint) (*model.LeaveBalance, error) {
	var balance model.LeaveBalance
	if err := r.db.Where("user_id = ? AND year = ? AND leave_type_id = ?", userID, year, leaveTypeID).
		First(&balance).Error; err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetLeaveBalancesByUser 查询用户某年所有假期余额
func (r *LeaveRepository) GetLeaveBalancesByUser(userID uint, year int) ([]model.LeaveBalance, error) {
	var balances []model.LeaveBalance
	if err := r.db.Preload("LeaveType").
		Where("user_id = ? AND year = ?", userID, year).
		Find(&balances).Error; err != nil {
		return nil, err
	}
	return balances, nil
}

// CreateLeaveBalance 创建假期余额记录
func (r *LeaveRepository) CreateLeaveBalance(balance *model.LeaveBalance) error {
	return r.db.Create(balance).Error
}

// UpdateLeaveBalance 更新假期余额（扣除已用天数）
func (r *LeaveRepository) UpdateLeaveBalance(balance *model.LeaveBalance) error {
	return r.db.Save(balance).Error
}

// DeductLeaveBalance 扣减假期余额
// usedDays 为正数表示扣减，负数表示返还
func (r *LeaveRepository) DeductLeaveBalance(userID uint, year int, leaveTypeID uint, usedDays float64) error {
	return r.db.Model(&model.LeaveBalance{}).
		Where("user_id = ? AND year = ? AND leave_type_id = ?", userID, year, leaveTypeID).
		Updates(map[string]interface{}{
			"used_days":      gorm.Expr("used_days + ?", usedDays),
			"remaining_days": gorm.Expr("remaining_days - ?", usedDays),
		}).Error
}

// ===================== 请假申请操作 =====================

// CreateLeaveApplication 创建请假申请
func (r *LeaveRepository) CreateLeaveApplication(application *model.LeaveApplication) error {
	return r.db.Create(application).Error
}

// GetLeaveApplicationByID 根据ID查询请假申请
func (r *LeaveRepository) GetLeaveApplicationByID(id uint) (*model.LeaveApplication, error) {
	var application model.LeaveApplication
	if err := r.db.Preload("LeaveType").First(&application, id).Error; err != nil {
		return nil, err
	}
	return &application, nil
}

// GetLeaveApplicationsByUser 查询用户请假申请列表
// 支持分页，按创建时间倒序
func (r *LeaveRepository) GetLeaveApplicationsByUser(userID uint, page, pageSize int) ([]model.LeaveApplication, int64, error) {
	var applications []model.LeaveApplication
	var total int64

	db := r.db.Model(&model.LeaveApplication{}).Where("applicant_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Preload("LeaveType").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&applications).Error; err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}

// GetLeaveApplicationsByStatus 按审批状态查询请假申请列表
func (r *LeaveRepository) GetLeaveApplicationsByStatus(status string, page, pageSize int) ([]model.LeaveApplication, int64, error) {
	var applications []model.LeaveApplication
	var total int64

	db := r.db.Model(&model.LeaveApplication{}).Where("status = ?", status)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Preload("LeaveType").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&applications).Error; err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}

// UpdateApprovalStatus 更新审批状态
func (r *LeaveRepository) UpdateApprovalStatus(id uint, approverID uint, status string, remark string) error {
	now := time.Now()
	return r.db.Model(&model.LeaveApplication{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         status,
			"approver_id":    approverID,
			"approve_remark": remark,
			"approved_at":    &now,
		}).Error
}

// GetLeaveApplicationsByUserAndTimeRange 查询用户在某时间范围内的请假申请
// 用于校验是否有重复请假
func (r *LeaveRepository) GetLeaveApplicationsByUserAndTimeRange(userID uint, startTime, endTime time.Time) ([]model.LeaveApplication, error) {
	var applications []model.LeaveApplication
	if err := r.db.Where("applicant_id = ? AND status != ? AND start_time < ? AND end_time > ?",
		userID, model.ApprovalStatusRejected, endTime, startTime).
		Find(&applications).Error; err != nil {
		return nil, err
	}
	return applications, nil
}
