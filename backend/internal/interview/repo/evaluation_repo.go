package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EvaluationRepo interface {
	CreateBatch(ctx context.Context, rows []model.Evaluation) error
	Update(ctx context.Context, ev *model.Evaluation) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error)
	ListByInterview(ctx context.Context, interviewID uuid.UUID) ([]model.EvaluationNamed, error)
	HasEvaluatorScores(ctx context.Context, interviewID, evaluatorID uuid.UUID) (bool, error)
	ListDimensions(ctx context.Context) ([]model.Dimension, error)
	GetDimensionByName(ctx context.Context, name string) (*model.Dimension, error)
	GetDimensionByID(ctx context.Context, id uuid.UUID) (*model.Dimension, error)
	CreateDimension(ctx context.Context, d *model.Dimension) error
	UpdateDimension(ctx context.Context, d *model.Dimension) error
	DeleteDimension(ctx context.Context, id uuid.UUID) error
	Stats(ctx context.Context, q *dto.StatsQuery) (model.StatsRow, []model.ScoreBucket, []model.DeptStat, error)
}

type evaluationRepo struct{ db *gorm.DB }

func NewEvaluationRepo(db *gorm.DB) EvaluationRepo {
	return &evaluationRepo{db: db}
}

func (r *evaluationRepo) CreateBatch(ctx context.Context, rows []model.Evaluation) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *evaluationRepo) Update(ctx context.Context, ev *model.Evaluation) error {
	return r.db.WithContext(ctx).Save(ev).Error
}

func (r *evaluationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	var ev model.Evaluation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ev).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *evaluationRepo) ListByInterview(ctx context.Context, interviewID uuid.UUID) ([]model.EvaluationNamed, error) {
	var rows []model.EvaluationNamed
	err := r.db.WithContext(ctx).Table("interview_evaluations AS e").
		Select("e.*, COALESCE(u.real_name, u.username, '') AS evaluator_name").
		Joins("LEFT JOIN users u ON u.id = e.interviewer_id").
		Where("e.interview_id = ?", interviewID).
		Order("e.created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *evaluationRepo) HasEvaluatorScores(ctx context.Context, interviewID, evaluatorID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Evaluation{}).
		Where("interview_id = ? AND interviewer_id = ?", interviewID, evaluatorID).
		Count(&n).Error
	return n > 0, err
}

func (r *evaluationRepo) ListDimensions(ctx context.Context) ([]model.Dimension, error) {
	var rows []model.Dimension
	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *evaluationRepo) GetDimensionByName(ctx context.Context, name string) (*model.Dimension, error) {
	var d model.Dimension
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&d).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *evaluationRepo) GetDimensionByID(ctx context.Context, id uuid.UUID) (*model.Dimension, error) {
	var d model.Dimension
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *evaluationRepo) CreateDimension(ctx context.Context, d *model.Dimension) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *evaluationRepo) UpdateDimension(ctx context.Context, d *model.Dimension) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *evaluationRepo) DeleteDimension(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Dimension{}, "id = ?", id).Error
}

func (r *evaluationRepo) Stats(ctx context.Context, q *dto.StatsQuery) (model.StatsRow, []model.ScoreBucket, []model.DeptStat, error) {
	base := r.statsBase(ctx, q)
	var row model.StatsRow
	err := base.Select(`COUNT(*) AS total,
		COUNT(*) FILTER (WHERE i.result_code = 1) AS pass_count,
		COUNT(*) FILTER (WHERE i.result_code = 2) AS fail_count,
		COUNT(*) FILTER (WHERE i.result_code IN (0, 3)) AS pending_count`).
		Scan(&row).Error
	if err != nil {
		return row, nil, nil, err
	}
	buckets, err := r.scoreBuckets(ctx, q)
	if err != nil {
		return row, nil, nil, err
	}
	depts, err := r.deptStats(ctx, q)
	return row, buckets, depts, err
}

func (r *evaluationRepo) statsBase(ctx context.Context, q *dto.StatsQuery) *gorm.DB {
	db := r.db.WithContext(ctx).Table("interviews AS i").
		Joins("LEFT JOIN interview_sessions s ON s.id = i.session_id").
		Where("i.status <> ?", model.InterviewCancelled)
	if q.DepartmentID != "" {
		db = db.Where("s.department_id = ?", q.DepartmentID)
	}
	if q.Round != nil {
		db = db.Where("s.round = ?", *q.Round)
	}
	if q.StartDate != "" {
		if t, err := time.Parse("2006-01-02", q.StartDate); err == nil {
			db = db.Where("i.created_at >= ?", t)
		}
	}
	if q.EndDate != "" {
		if t, err := time.Parse("2006-01-02", q.EndDate); err == nil {
			db = db.Where("i.created_at < ?", t.Add(24*time.Hour))
		}
	}
	return db
}

func (r *evaluationRepo) scoreBuckets(ctx context.Context, q *dto.StatsQuery) ([]model.ScoreBucket, error) {
	type row struct {
		Bucket string
		Count  int64
	}
	var rows []row
	err := r.statsBase(ctx, q).
		Select(`CASE
			WHEN i.score IS NULL THEN '未评分'
			WHEN i.score < 60 THEN '0-59'
			WHEN i.score < 70 THEN '60-69'
			WHEN i.score < 80 THEN '70-79'
			WHEN i.score < 90 THEN '80-89'
			ELSE '90-100' END AS bucket, COUNT(*) AS count`).
		Group("bucket").Scan(&rows).Error
	out := []model.ScoreBucket{
		{Range: "0-59"}, {Range: "60-69"}, {Range: "70-79"}, {Range: "80-89"}, {Range: "90-100"}, {Range: "未评分"},
	}
	idx := map[string]int{"0-59": 0, "60-69": 1, "70-79": 2, "80-89": 3, "90-100": 4, "未评分": 5}
	for _, item := range rows {
		if i, ok := idx[item.Bucket]; ok {
			out[i].Count = item.Count
		}
	}
	return out, err
}

func (r *evaluationRepo) deptStats(ctx context.Context, q *dto.StatsQuery) ([]model.DeptStat, error) {
	var rows []model.DeptStat
	err := r.statsBase(ctx, q).
		Select("COALESCE(d.name, '未分配') AS department, COUNT(*) AS count, COUNT(*) FILTER (WHERE i.result_code = 1) AS pass_count").
		Joins("LEFT JOIN departments d ON d.id = s.department_id").
		Group("d.name").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}
