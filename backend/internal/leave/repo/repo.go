package repo

import (
	"context"
	"time"

	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	WithTx(ctx context.Context, fn func(Repository) error) error
	DB() *gorm.DB

	GetAllLeaveTypes(ctx context.Context) ([]model.LeaveType, error)
	GetLeaveTypeByID(ctx context.Context, id uuid.UUID) (*model.LeaveType, error)
	GetLeaveTypeByCode(ctx context.Context, code string) (*model.LeaveType, error)
	CreateLeaveType(ctx context.Context, row *model.LeaveType) error
	UpdateLeaveType(ctx context.Context, row *model.LeaveType) error

	GetLeaveBalance(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error)
	GetLeaveBalanceForUpdate(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error)
	GetLeaveBalancesByUser(ctx context.Context, userID uuid.UUID, year int) ([]model.LeaveBalance, error)
	CreateLeaveBalance(ctx context.Context, balance *model.LeaveBalance) error
	DeductLeaveBalance(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID, usedDays float64) error

	CreateLeaveApplication(ctx context.Context, application *model.LeaveApplication) error
	GetLeaveApplicationByID(ctx context.Context, id uuid.UUID) (*model.ApplicationNamed, error)
	GetLeaveApplicationByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.ApplicationNamed, error)
	LockApplicant(ctx context.Context, userID uuid.UUID) error
	GetLeaveApplicationsByUser(ctx context.Context, userID uuid.UUID, status string, page, pageSize int, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, int64, error)
	GetLeaveApplicationsByStatus(ctx context.Context, status string, page, pageSize int, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, int64, error)
	UpdateApprovalStatus(ctx context.Context, id uuid.UUID, approverID uuid.UUID, status, remark string, at time.Time) error
	SaveApplication(ctx context.Context, app *model.LeaveApplication) error
	GetLeaveApplicationsByUserAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]model.LeaveApplication, error)
	ListCalendar(ctx context.Context, from, to time.Time, userID *uuid.UUID, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, error)
	CountStats(ctx context.Context, scope *rbacModel.DataScopeCondition) (total int64, byStatus map[string]int64, byType []TypeCount, err error)
	CountPersonalStats(ctx context.Context, userID uuid.UUID) (total int64, days float64, byType []TypeCount, err error)
	CountDepartmentStats(ctx context.Context, scope *rbacModel.DataScopeCondition) ([]DeptCount, error)
	CountMonthlyStats(ctx context.Context, year int, scope *rbacModel.DataScopeCondition) ([]MonthCount, error)
	GetUserDepartmentID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
}

type TypeCount struct {
	LeaveTypeID uuid.UUID
	Code        string
	Name        string
	Count       int64
	Days        float64
}

type DeptCount struct {
	DepartmentID   *uuid.UUID
	DepartmentName string
	Count          int64
	Days           float64
}

type MonthCount struct {
	Month string
	Count int64
	Days  float64
}

type leaveRepo struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &leaveRepo{db: db} }

func (r *leaveRepo) DB() *gorm.DB { return r.db }

func (r *leaveRepo) WithTx(ctx context.Context, fn func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&leaveRepo{db: tx})
	})
}

func (r *leaveRepo) GetAllLeaveTypes(ctx context.Context) ([]model.LeaveType, error) {
	var rows []model.LeaveType
	err := r.db.WithContext(ctx).Order("sort_order, code").Find(&rows).Error
	return rows, err
}

func (r *leaveRepo) GetLeaveTypeByCode(ctx context.Context, code string) (*model.LeaveType, error) {
	var row model.LeaveType
	err := r.db.WithContext(ctx).First(&row, "code = ?", code).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *leaveRepo) CreateLeaveType(ctx context.Context, row *model.LeaveType) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *leaveRepo) UpdateLeaveType(ctx context.Context, row *model.LeaveType) error {
	return r.db.WithContext(ctx).Omit("CreatedAt").Save(row).Error
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
	return r.db.WithContext(ctx).Omit("LeaveType").Create(balance).Error
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
		COALESCE(approver.real_name, approver.username, '') AS approver_name,
		applicant.department_id AS applicant_department_id`
}

func applyApplicantScope(q *gorm.DB, scope *rbacModel.DataScopeCondition) *gorm.DB {
	if scope == nil || strings.TrimSpace(scope.Query) == "" {
		return q
	}
	return q.Where(scope.Query, scope.Args...)
}

func (r *leaveRepo) joinApplicant(q *gorm.DB, scope *rbacModel.DataScopeCondition) *gorm.DB {
	if scope != nil && strings.Contains(scope.Query, "applicant.") {
		q = q.Joins("JOIN users applicant ON applicant.id = leave_applications.applicant_id")
	}
	return applyApplicantScope(q, scope)
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

func (r *leaveRepo) GetLeaveApplicationByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.ApplicationNamed, error) {
	var lock model.LeaveApplication
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&lock, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.GetLeaveApplicationByID(ctx, id)
}

func (r *leaveRepo) LockApplicant(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(`SELECT pg_advisory_xact_lock(hashtext(?))`, "leave-applicant:"+userID.String()).Error
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

func (r *leaveRepo) GetLeaveApplicationsByUser(ctx context.Context, userID uuid.UUID, status string, page, pageSize int, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, int64, error) {
	countQ := r.joinApplicant(r.db.WithContext(ctx).Model(&model.LeaveApplication{}).Where("leave_applications.applicant_id = ?", userID), scope)
	listQ := applyApplicantScope(r.namedQuery(ctx).Where("leave_applications.applicant_id = ?", userID), scope)
	if status != "" {
		countQ = countQ.Where("leave_applications.status = ?", status)
		listQ = listQ.Where("leave_applications.status = ?", status)
	}
	return paginateNamed(countQ, listQ, page, pageSize)
}

func (r *leaveRepo) GetLeaveApplicationsByStatus(ctx context.Context, status string, page, pageSize int, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, int64, error) {
	countQ := r.joinApplicant(r.db.WithContext(ctx).Model(&model.LeaveApplication{}), scope)
	listQ := applyApplicantScope(r.namedQuery(ctx), scope)
	if status != "" {
		countQ = countQ.Where("leave_applications.status = ?", status)
		listQ = listQ.Where("leave_applications.status = ?", status)
	}
	return paginateNamed(countQ, listQ, page, pageSize)
}

func (r *leaveRepo) UpdateApprovalStatus(ctx context.Context, id uuid.UUID, approverID uuid.UUID, status, remark string, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.LeaveApplication{}).
		Where("id = ? AND status = ?", id, model.ApprovalStatusPending).
		Updates(map[string]interface{}{
			"status":         status,
			"approver_id":    approverID,
			"approve_remark": remark,
			"approved_at":    at,
			"updated_at":     at,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotPending
	}
	return nil
}

func (r *leaveRepo) SaveApplication(ctx context.Context, app *model.LeaveApplication) error {
	return r.db.WithContext(ctx).Omit("LeaveType").Save(app).Error
}

func (r *leaveRepo) GetLeaveApplicationsByUserAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]model.LeaveApplication, error) {
	var rows []model.LeaveApplication
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("applicant_id = ? AND status <> ? AND start_time < ? AND end_time > ?",
			userID, model.ApprovalStatusRejected, endTime, startTime).
		Find(&rows).Error
	return rows, err
}

func (r *leaveRepo) GetUserDepartmentID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	var row struct {
		DepartmentID *uuid.UUID
	}
	err := r.db.WithContext(ctx).Raw(`SELECT department_id FROM users WHERE id = ? AND deleted_at IS NULL`, userID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return row.DepartmentID, nil
}

func (r *leaveRepo) CountStats(ctx context.Context, scope *rbacModel.DataScopeCondition) (int64, map[string]int64, []TypeCount, error) {
	var total int64
	if err := r.joinApplicant(r.db.WithContext(ctx).Model(&model.LeaveApplication{}), scope).Count(&total).Error; err != nil {
		return 0, nil, nil, err
	}
	type statusRow struct {
		Status string
		Count  int64
	}
	var statusRows []statusRow
	if err := r.joinApplicant(r.db.WithContext(ctx).Model(&model.LeaveApplication{}), scope).
		Select("leave_applications.status, count(*) as count").
		Group("leave_applications.status").
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
	typeQ := r.db.WithContext(ctx).
		Table("leave_applications AS a").
		Select("a.leave_type_id, t.code, t.name, count(*) AS count, coalesce(sum(a.duration_days),0) AS days").
		Joins("JOIN leave_types t ON t.id = a.leave_type_id").
		Where("a.deleted_at IS NULL")
	if scope != nil && strings.Contains(scope.Query, "applicant.") {
		typeQ = typeQ.Joins("JOIN users applicant ON applicant.id = a.applicant_id")
	}
	if scope != nil && strings.Contains(scope.Query, "leave_applications.applicant_id") {
		scope = &rbacModel.DataScopeCondition{
			Query:  strings.ReplaceAll(scope.Query, "leave_applications.applicant_id", "a.applicant_id"),
			Args:   scope.Args,
			IsSelf: scope.IsSelf,
		}
	}
	err := applyApplicantScope(typeQ, scope).
		Group("a.leave_type_id, t.code, t.name").
		Order("t.code").
		Scan(&typeRows).Error
	return total, byStatus, typeRows, err
}

func (r *leaveRepo) ListCalendar(ctx context.Context, from, to time.Time, userID *uuid.UUID, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, error) {
	q := applyApplicantScope(r.namedQuery(ctx), scope).
		Where("leave_applications.start_time < ? AND leave_applications.end_time > ?", to, from).
		Where("leave_applications.status <> ?", model.ApprovalStatusRejected)
	if userID != nil {
		q = q.Where("leave_applications.applicant_id = ?", *userID)
	}
	var rows []model.ApplicationNamed
	err := q.Order("leave_applications.start_time").Limit(500).Find(&rows).Error
	return rows, err
}

func (r *leaveRepo) CountPersonalStats(ctx context.Context, userID uuid.UUID) (int64, float64, []TypeCount, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.LeaveApplication{}).
		Where("applicant_id = ?", userID).Count(&total).Error; err != nil {
		return 0, 0, nil, err
	}
	var days float64
	if err := r.db.WithContext(ctx).Model(&model.LeaveApplication{}).
		Where("applicant_id = ? AND status <> ?", userID, model.ApprovalStatusRejected).
		Select("coalesce(sum(duration_days),0)").Scan(&days).Error; err != nil {
		return 0, 0, nil, err
	}
	var typeRows []TypeCount
	err := r.db.WithContext(ctx).
		Table("leave_applications AS a").
		Select("a.leave_type_id, t.code, t.name, count(*) AS count, coalesce(sum(a.duration_days),0) AS days").
		Joins("JOIN leave_types t ON t.id = a.leave_type_id").
		Where("a.deleted_at IS NULL AND a.applicant_id = ?", userID).
		Group("a.leave_type_id, t.code, t.name").
		Order("t.code").
		Scan(&typeRows).Error
	return total, days, typeRows, err
}

func (r *leaveRepo) CountDepartmentStats(ctx context.Context, scope *rbacModel.DataScopeCondition) ([]DeptCount, error) {
	q := r.db.WithContext(ctx).
		Table("leave_applications AS a").
		Select("applicant.department_id, coalesce(d.name,'') AS department_name, count(*) AS count, coalesce(sum(a.duration_days),0) AS days").
		Joins("JOIN users applicant ON applicant.id = a.applicant_id").
		Joins("LEFT JOIN departments d ON d.id = applicant.department_id").
		Where("a.deleted_at IS NULL AND a.status <> ?", model.ApprovalStatusRejected)
	if scope != nil && strings.Contains(scope.Query, "leave_applications.applicant_id") {
		scope = &rbacModel.DataScopeCondition{
			Query:  strings.ReplaceAll(scope.Query, "leave_applications.applicant_id", "a.applicant_id"),
			Args:   scope.Args,
			IsSelf: scope.IsSelf,
		}
	}
	var rows []DeptCount
	err := applyApplicantScope(q, scope).
		Group("applicant.department_id, d.name").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *leaveRepo) CountMonthlyStats(ctx context.Context, year int, scope *rbacModel.DataScopeCondition) ([]MonthCount, error) {
	q := r.db.WithContext(ctx).
		Table("leave_applications AS a").
		Select("to_char(timezone('Asia/Shanghai', a.start_time), 'YYYY-MM') AS month, count(*) AS count, coalesce(sum(a.duration_days),0) AS days").
		Where("a.deleted_at IS NULL")
	if year > 0 {
		q = q.Where("extract(year from timezone('Asia/Shanghai', a.start_time)) = ?", year)
	}
	if scope != nil && strings.Contains(scope.Query, "applicant.") {
		q = q.Joins("JOIN users applicant ON applicant.id = a.applicant_id")
	}
	if scope != nil && strings.Contains(scope.Query, "leave_applications.applicant_id") {
		scope = &rbacModel.DataScopeCondition{
			Query:  strings.ReplaceAll(scope.Query, "leave_applications.applicant_id", "a.applicant_id"),
			Args:   scope.Args,
			IsSelf: scope.IsSelf,
		}
	}
	var rows []MonthCount
	err := applyApplicantScope(q, scope).
		Group("month").
		Order("month").
		Scan(&rows).Error
	return rows, err
}
