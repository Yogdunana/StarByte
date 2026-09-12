package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	WithTx(ctx context.Context, fn func(Repository) error) error

	GetAllLeaveTypes(ctx context.Context) ([]model.LeaveType, error)
	GetLeaveTypeByID(ctx context.Context, id uuid.UUID) (*model.LeaveType, error)

	GetLeaveBalance(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error)
	GetLeaveBalanceForUpdate(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error)
	GetLeaveBalancesByUser(ctx context.Context, userID uuid.UUID, year int) ([]model.LeaveBalance, error)
	CreateLeaveBalance(ctx context.Context, balance *model.LeaveBalance) error
	DeductLeaveBalance(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID, usedDays float64) error

	CreateLeaveApplication(ctx context.Context, application *model.LeaveApplication) error
	GetLeaveApplicationByID(ctx context.Context, id uuid.UUID) (*model.ApplicationNamed, error)
	GetLeaveApplicationsByUser(ctx context.Context, userID uuid.UUID, status string, page, pageSize int) ([]model.ApplicationNamed, int64, error)
	GetLeaveApplicationsByStatus(ctx context.Context, status string, page, pageSize int) ([]model.ApplicationNamed, int64, error)
	UpdateApprovalStatus(ctx context.Context, id uuid.UUID, approverID uuid.UUID, status, remark string, at time.Time) error
	GetLeaveApplicationsByUserAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]model.LeaveApplication, error)
	CountStats(ctx context.Context) (total int64, byStatus map[string]int64, byType []TypeCount, err error)
}

type TypeCount struct {
	LeaveTypeID uuid.UUID
	Code        string
	Name        string
	Count       int64
	Days        float64
}

type leaveRepo struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &leaveRepo{db: db} }

func (r *leaveRepo) WithTx(ctx context.Context, fn func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&leaveRepo{db: tx})
	})
}

func (r *leaveRepo) GetAllLeaveTypes(ctx context.Context) ([]model.LeaveType, error) {
	var rows []model.LeaveType
	err := r.db.WithContext(ctx).Order("code").Find(&rows).Error
	return rows, err
}

func (r *leaveRepo) GetLeaveTypeByID(ctx context.Context, id uuid.UUID) (*model.LeaveType, error) {
	var row model.LeaveType
	err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *leaveRepo) findBalance(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID, forUpdate bool) (*model.LeaveBalance, error) {
	q := r.db.WithContext(ctx).Preload("LeaveType").
		Where("user_id = ? AND year = ? AND leave_type_id = ?", userID, year, leaveTypeID)
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var row model.LeaveBalance
	err := q.First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *leaveRepo) GetLeaveBalance(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error) {
	return r.findBalance(ctx, userID, year, leaveTypeID, false)
}

func (r *leaveRepo) GetLeaveBalanceForUpdate(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error) {
	return r.findBalance(ctx, userID, year, leaveTypeID, true)
}

func (r *leaveRepo) GetLeaveBalancesByUser(ctx context.Context, userID uuid.UUID, year int) ([]model.LeaveBalance, error) {
	var rows []model.LeaveBalance
	err := r.db.WithContext(ctx).Preload("LeaveType").
		Where("user_id = ? AND year = ?", userID, year).
		Order("leave_type_id").
		Find(&rows).Error
	return rows, err
}

func (r *leaveRepo) CreateLeaveBalance(ctx context.Context, balance *model.LeaveBalance) error {
	return r.db.WithContext(ctx).Create(balance).Error
}

func (r *leaveRepo) DeductLeaveBalance(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID, usedDays float64) error {
	return r.db.WithContext(ctx).Model(&model.LeaveBalance{}).
		Where("user_id = ? AND year = ? AND leave_type_id = ?", userID, year, leaveTypeID).
		Updates(map[string]interface{}{
			"used_days":      gorm.Expr("used_days + ?", usedDays),
			"remaining_days": gorm.Expr("remaining_days - ?", usedDays),
			"updated_at":     time.Now(),
		}).Error
}

func (r *leaveRepo) CreateLeaveApplication(ctx context.Context, application *model.LeaveApplication) error {
	return r.db.WithContext(ctx).Create(application).Error
}

func namedSelect() string {
	return `leave_applications.*,
		COALESCE(applicant.real_name, applicant.username, '') AS applicant_name,
		COALESCE(approver.real_name, approver.username, '') AS approver_name`
}

func (r *leaveRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&model.LeaveApplication{}).
		Select(namedSelect()).
		Joins("LEFT JOIN users applicant ON applicant.id = leave_applications.applicant_id").
		Joins("LEFT JOIN users approver ON approver.id = leave_applications.approver_id").
		Preload("LeaveType")
}

func (r *leaveRepo) GetLeaveApplicationByID(ctx context.Context, id uuid.UUID) (*model.ApplicationNamed, error) {
	var row model.ApplicationNamed
	err := r.namedQuery(ctx).Where("leave_applications.id = ?", id).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func paginateNamed(countDB, listDB *gorm.DB, page, pageSize int) ([]model.ApplicationNamed, int64, error) {
	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var rows []model.ApplicationNamed
	err := listDB.Order("leave_applications.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *leaveRepo) GetLeaveApplicationsByUser(ctx context.Context, userID uuid.UUID, status string, page, pageSize int) ([]model.ApplicationNamed, int64, error) {
	countQ := r.db.WithContext(ctx).Model(&model.LeaveApplication{}).Where("applicant_id = ?", userID)
	listQ := r.namedQuery(ctx).Where("leave_applications.applicant_id = ?", userID)
	if status != "" {
		countQ = countQ.Where("status = ?", status)
		listQ = listQ.Where("leave_applications.status = ?", status)
	}
	return paginateNamed(countQ, listQ, page, pageSize)
}

func (r *leaveRepo) GetLeaveApplicationsByStatus(ctx context.Context, status string, page, pageSize int) ([]model.ApplicationNamed, int64, error) {
	countQ := r.db.WithContext(ctx).Model(&model.LeaveApplication{})
	listQ := r.namedQuery(ctx)
	if status != "" {
		countQ = countQ.Where("status = ?", status)
		listQ = listQ.Where("leave_applications.status = ?", status)
	}
	return paginateNamed(countQ, listQ, page, pageSize)
}

func (r *leaveRepo) UpdateApprovalStatus(ctx context.Context, id uuid.UUID, approverID uuid.UUID, status, remark string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&model.LeaveApplication{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         status,
			"approver_id":    approverID,
			"approve_remark": remark,
			"approved_at":    at,
			"updated_at":     at,
		}).Error
}

func (r *leaveRepo) GetLeaveApplicationsByUserAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]model.LeaveApplication, error) {
	var rows []model.LeaveApplication
	err := r.db.WithContext(ctx).
		Where("applicant_id = ? AND status <> ? AND start_time < ? AND end_time > ?",
			userID, model.ApprovalStatusRejected, endTime, startTime).
		Find(&rows).Error
	return rows, err
}

func (r *leaveRepo) CountStats(ctx context.Context) (int64, map[string]int64, []TypeCount, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.LeaveApplication{}).Count(&total).Error; err != nil {
		return 0, nil, nil, err
	}
	type statusRow struct {
		Status string
		Count  int64
	}
	var statusRows []statusRow
	if err := r.db.WithContext(ctx).Model(&model.LeaveApplication{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusRows).Error; err != nil {
		return 0, nil, nil, err
	}
	byStatus := map[string]int64{
		model.ApprovalStatusPending:  0,
		model.ApprovalStatusApproved: 0,
		model.ApprovalStatusRejected: 0,
	}
	for _, row := range statusRows {
		byStatus[row.Status] = row.Count
	}
	var typeRows []TypeCount
	err := r.db.WithContext(ctx).
		Table("leave_applications AS a").
		Select("a.leave_type_id, t.code, t.name, count(*) AS count, coalesce(sum(a.duration_days),0) AS days").
		Joins("JOIN leave_types t ON t.id = a.leave_type_id").
		Where("a.deleted_at IS NULL").
		Group("a.leave_type_id, t.code, t.name").
		Order("t.code").
		Scan(&typeRows).Error
	return total, byStatus, typeRows, err
}
