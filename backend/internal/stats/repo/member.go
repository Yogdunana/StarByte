package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
)

func (r *statsRepo) MemberSummary(ctx context.Context, q Query) (total, active, newMonth int64, err error) {
	base := r.db.WithContext(ctx).Table("member_profiles")
	base = applyDept(base, "department_id", q)
	if err = base.Count(&total).Error; err != nil {
		return
	}
	if err = applyDept(r.db.WithContext(ctx).Table("member_profiles").Where("status = 0"), "department_id", q).Count(&active).Error; err != nil {
		return
	}
	start := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	if err = applyDept(r.db.WithContext(ctx).Table("member_profiles").Where("created_at >= ?", start), "department_id", q).Count(&newMonth).Error; err != nil {
		return
	}
	return
}

func (r *statsRepo) MemberByDepartment(ctx context.Context, q Query) ([]Bucket, error) {
	db := r.db.WithContext(ctx).Table("member_profiles AS p").
		Select("COALESCE(p.department_id::text, 'none') AS k, COALESCE(d.name, '未分配') AS l, COUNT(*)::float AS v").
		Joins("LEFT JOIN departments d ON d.id = p.department_id")
	db = applyDept(db, "p.department_id", q)
	return scanBuckets(db.Group("p.department_id, d.name").Order("v DESC"))
}

func (r *statsRepo) MemberByGrade(ctx context.Context, q Query) ([]Bucket, error) {
	db := r.db.WithContext(ctx).Table("member_profiles").
		Select("COALESCE(NULLIF(grade, ''), '未填写') AS k, COALESCE(NULLIF(grade, ''), '未填写') AS l, COUNT(*)::float AS v")
	db = applyDept(db, "department_id", q)
	return scanBuckets(db.Group("k, l").Order("k"))
}

func (r *statsRepo) MemberTrend(ctx context.Context, q Query) ([]Bucket, error) {
	expr := truncExpr("created_at", q.Granularity)
	db := r.db.WithContext(ctx).Table("member_profiles").
		Select(expr + " AS k, " + expr + " AS l, COUNT(*)::float AS v")
	db = applyDept(db, "department_id", q)
	db = applyRange(db, "created_at", q)
	return scanBuckets(db.Group("k, l").Order("k"))
}

func (r *statsRepo) InterviewSummary(ctx context.Context, q Query) (total int64, passRate, avgScore float64, err error) {
	type agg struct {
		Total int64
		Pass  int64
		Avg   *float64
	}
	var row agg
	db := r.ivJoin(ctx, q).Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE i.result_code = 1) AS pass, AVG(i.score) AS avg")
	if err = db.Scan(&row).Error; err != nil {
		return
	}
	total = row.Total
	if total > 0 {
		passRate = float64(row.Pass) / float64(total)
	}
	if row.Avg != nil {
		avgScore = *row.Avg
	}
	return
}

func (r *statsRepo) ivJoin(ctx context.Context, q Query) *gorm.DB {
	db := r.db.WithContext(ctx).Table("interviews AS i").
		Joins("LEFT JOIN member_applications a ON a.id = i.application_id")
	db = applyDept(db, "a.department_id", q)
	db = applyRange(db, "COALESCE(i.scheduled_at, i.created_at)", q)
	return db
}

func (r *statsRepo) InterviewByDepartment(ctx context.Context, q Query) ([]Bucket, error) {
	db := r.ivJoin(ctx, q).
		Select("COALESCE(a.department_id::text, 'none') AS k, COALESCE(d.name, '未分配') AS l, COUNT(*)::float AS v").
		Joins("LEFT JOIN departments d ON d.id = a.department_id")
	return scanBuckets(db.Group("a.department_id, d.name").Order("v DESC"))
}

func (r *statsRepo) InterviewTrend(ctx context.Context, q Query) ([]Bucket, error) {
	expr := truncExpr("COALESCE(i.scheduled_at, i.created_at)", q.Granularity)
	db := r.ivJoin(ctx, q).Select(expr + " AS k, " + expr + " AS l, COUNT(*)::float AS v")
	return scanBuckets(db.Group("k, l").Order("k"))
}

func (r *statsRepo) InterviewScoreHist(ctx context.Context, q Query) ([]Bucket, error) {
	db := r.ivJoin(ctx, q).Where("i.score IS NOT NULL").
		Select("(FLOOR(i.score / 10) * 10)::int::text AS k, (FLOOR(i.score / 10) * 10)::int::text || '-' || (FLOOR(i.score / 10) * 10 + 9)::int::text AS l, COUNT(*)::float AS v")
	return scanBuckets(db.Group("k, l").Order("k"))
}
