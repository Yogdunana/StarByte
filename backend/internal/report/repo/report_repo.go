package repo

import (
	"context"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/report/dto"
	"github.com/Yogdunana/StarByte/backend/internal/report/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportRepo interface {
	Create(ctx context.Context, report *model.Report) error
	Update(ctx context.Context, report *model.Report) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Report, error)
	GetByUserAndPeriod(
		ctx context.Context,
		userID uuid.UUID,
		reportType model.ReportType,
		periodStart time.Time,
		periodEnd time.Time,
	) (*model.Report, error)
	List(
		ctx context.Context,
		req *dto.ListReportRequest,
		viewerID uuid.UUID,
		scope *rbacModel.DataScopeCondition,
	) ([]model.Report, int64, error)
}

type reportRepo struct{ db *gorm.DB }

func NewReportRepo(db *gorm.DB) ReportRepo {
	return &reportRepo{db: db}
}

func (r *reportRepo) Create(ctx context.Context, report *model.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *reportRepo) Update(ctx context.Context, report *model.Report) error {
	return r.db.WithContext(ctx).Save(report).Error
}

func (r *reportRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Report, error) {
	var report model.Report
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&report).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepo) GetByUserAndPeriod(
	ctx context.Context,
	userID uuid.UUID,
	reportType model.ReportType,
	periodStart time.Time,
	periodEnd time.Time,
) (*model.Report, error) {
	var report model.Report
	err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND report_type = ? AND period_start = ? AND period_end = ?",
			userID,
			reportType,
			periodStart,
			periodEnd,
		).
		First(&report).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepo) List(
	ctx context.Context,
	req *dto.ListReportRequest,
	viewerID uuid.UUID,
	scope *rbacModel.DataScopeCondition,
) ([]model.Report, int64, error) {
	q := applyListFilters(r.db.WithContext(ctx).Model(&model.Report{}), req, viewerID, scope)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizePage(req.Page, req.PageSize)
	var reports []model.Report
	err := q.Order("period_start DESC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&reports).Error
	return reports, total, err
}

func applyListFilters(
	q *gorm.DB,
	req *dto.ListReportRequest,
	viewerID uuid.UUID,
	scope *rbacModel.DataScopeCondition,
) *gorm.DB {
	switch {
	case scope == nil:
		// Fail closed: a missing scope must never list rows.
		q = q.Where("1 = 0")
	case scope.IsSelf:
		q = q.Where("user_id = ?", viewerID)
	case !scope.IsEmpty():
		q = q.Where(scope.Query, scope.Args...)
	}
	if req.ReportType != "" {
		q = q.Where("report_type = ?", req.ReportType)
	}
	if req.ReviewStatus != "" {
		q = q.Where("review_status = ?", req.ReviewStatus)
	}
	if req.UserID != "" {
		q = q.Where("user_id = ?", req.UserID)
	}
	if req.DepartmentID != "" {
		q = q.Where("department_id = ?", req.DepartmentID)
	}
	if req.PeriodStart != nil {
		q = q.Where("period_end >= ?", *req.PeriodStart)
	}
	if req.PeriodEnd != nil {
		q = q.Where("period_start <= ?", *req.PeriodEnd)
	}
	return q
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
