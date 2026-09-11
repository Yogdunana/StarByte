package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/duty/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScheduleRepo 排班数据访问接口
type ScheduleRepo interface {
	Create(ctx context.Context, tx *gorm.DB, schedule *model.Schedule) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	Update(ctx context.Context, tx *gorm.DB, schedule *model.Schedule) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int, departmentID, userID uuid.UUID, startDate, endDate *time.Time) ([]model.Schedule, int64, error)
	CountByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time, timeSlot string) (int64, error)
	StatsByUser(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time) ([]UserDutyStat, error)
	StatsByDepartment(ctx context.Context, departmentID uuid.UUID, startDate, endDate *time.Time) ([]UserDutyStat, error)
	FindPendingBefore(ctx context.Context, before time.Time) ([]model.Schedule, error)
}

type scheduleRepo struct {
	db *gorm.DB
}

func NewScheduleRepo(db *gorm.DB) ScheduleRepo {
	return &scheduleRepo{db: db}
}

func (r *scheduleRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *scheduleRepo) Create(ctx context.Context, tx *gorm.DB, schedule *model.Schedule) error {
	return r.getDB(tx).WithContext(ctx).Create(schedule).Error
}

func (r *scheduleRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	var s model.Schedule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

func (r *scheduleRepo) Update(ctx context.Context, tx *gorm.DB, schedule *model.Schedule) error {
	return r.getDB(tx).WithContext(ctx).Save(schedule).Error
}

func (r *scheduleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Schedule{}, id).Error
}

func (r *scheduleRepo) List(ctx context.Context, page, pageSize int, departmentID, userID uuid.UUID, startDate, endDate *time.Time) ([]model.Schedule, int64, error) {
	var schedules []model.Schedule
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Schedule{})

	if departmentID != uuid.Nil {
		query = query.Where("department_id = ?", departmentID)
	}
	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}
	if startDate != nil {
		query = query.Where("duty_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("duty_date <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("duty_date ASC, time_slot ASC").Offset(offset).Limit(pageSize).Find(&schedules).Error
	return schedules, total, err
}

func (r *scheduleRepo) CountByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time, timeSlot string) (int64, error) {
	var count int64
	dateOnly := date.Format("2006-01-02")
	err := r.db.WithContext(ctx).Model(&model.Schedule{}).
		Where("user_id = ? AND duty_date = ? AND time_slot = ? AND deleted_at IS NULL", userID, dateOnly, timeSlot).
		Count(&count).Error
	return count, err
}

func (r *scheduleRepo) StatsByUser(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time) ([]UserDutyStat, error) {
	var stats []UserDutyStat
	query := r.db.WithContext(ctx).Table("duty_schedules").
		Select("user_id, COUNT(*) as total_count, SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as completed, SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END) as absent").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Group("user_id")

	if startDate != nil {
		query = query.Where("duty_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("duty_date <= ?", *endDate)
	}

	err := query.Scan(&stats).Error
	return stats, err
}

func (r *scheduleRepo) StatsByDepartment(ctx context.Context, departmentID uuid.UUID, startDate, endDate *time.Time) ([]UserDutyStat, error) {
	var stats []UserDutyStat
	query := r.db.WithContext(ctx).Table("duty_schedules").
		Select("user_id, COUNT(*) as total_count, SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as completed, SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END) as absent").
		Where("department_id = ? AND deleted_at IS NULL", departmentID).
		Group("user_id")

	if startDate != nil {
		query = query.Where("duty_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("duty_date <= ?", *endDate)
	}

	err := query.Scan(&stats).Error
	return stats, err
}

func (r *scheduleRepo) FindPendingBefore(ctx context.Context, before time.Time) ([]model.Schedule, error) {
	var schedules []model.Schedule
	err := r.db.WithContext(ctx).
		Where("status = 0 AND duty_date <= ? AND deleted_at IS NULL", before.Format("2006-01-02")).
		Find(&schedules).Error
	return schedules, err
}

// UserDutyStat 用户值班统计
type UserDutyStat struct {
	UserID     uuid.UUID `gorm:"column:user_id"`
	TotalCount int64     `gorm:"column:total_count"`
	Completed  int64     `gorm:"column:completed"`
	Absent     int64     `gorm:"column:absent"`
}

// SwapRepo 调班申请数据访问接口
type SwapRepo interface {
	Create(ctx context.Context, tx *gorm.DB, req *model.SwapRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.SwapRequest, error)
	Update(ctx context.Context, tx *gorm.DB, req *model.SwapRequest) error
	List(ctx context.Context, page, pageSize int, requesterID uuid.UUID, status *int) ([]model.SwapRequest, int64, error)
}

type swapRepo struct {
	db *gorm.DB
}

func NewSwapRepo(db *gorm.DB) SwapRepo {
	return &swapRepo{db: db}
}

func (r *swapRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *swapRepo) Create(ctx context.Context, tx *gorm.DB, req *model.SwapRequest) error {
	return r.getDB(tx).WithContext(ctx).Create(req).Error
}

func (r *swapRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.SwapRequest, error) {
	var sr model.SwapRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&sr).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &sr, err
}

func (r *swapRepo) Update(ctx context.Context, tx *gorm.DB, req *model.SwapRequest) error {
	return r.getDB(tx).WithContext(ctx).Save(req).Error
}

func (r *swapRepo) List(ctx context.Context, page, pageSize int, requesterID uuid.UUID, status *int) ([]model.SwapRequest, int64, error) {
	var swaps []model.SwapRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SwapRequest{})

	if requesterID != uuid.Nil {
		query = query.Where("requester_id = ?", requesterID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&swaps).Error
	return swaps, total, err
}
