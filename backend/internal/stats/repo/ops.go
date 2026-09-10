package repo

import (
	"context"
)

func (r *statsRepo) MeetingAttendanceTrend(ctx context.Context, q Query) ([]Bucket, error) {
	expr := truncExpr("m.start_time", q.Granularity)
	db := r.db.WithContext(ctx).Table("meetings AS m").
		Joins("LEFT JOIN meeting_attendees a ON a.meeting_id = m.id").
		Joins("LEFT JOIN member_profiles p ON p.user_id = m.organizer_id").
		Select(expr + ` AS k, ` + expr + ` AS l,
			CASE WHEN COUNT(a.id) = 0 THEN 0
			ELSE 100.0 * COUNT(*) FILTER (WHERE a.attended) / COUNT(a.id) END AS v`)
	db = applyDept(db, "p.department_id", q)
	db = applyRange(db, "m.start_time", q)
	return scanBuckets(db.Group("k, l").Order("k"))
}

func (r *statsRepo) MeetingByDepartment(ctx context.Context, q Query) ([]Bucket, error) {
	db := r.db.WithContext(ctx).Table("meetings AS m").
		Joins("LEFT JOIN member_profiles p ON p.user_id = m.organizer_id").
		Joins("LEFT JOIN departments d ON d.id = p.department_id").
		Select("COALESCE(p.department_id::text, 'none') AS k, COALESCE(d.name, '未分配') AS l, COUNT(*)::float AS v")
	db = applyDept(db, "p.department_id", q)
	db = applyRange(db, "m.start_time", q)
	return scanBuckets(db.Group("p.department_id, d.name").Order("v DESC"))
}

func (r *statsRepo) MeetingCalendar(ctx context.Context, q Query) ([]Bucket, error) {
	db := r.db.WithContext(ctx).Table("meetings AS m").
		Joins("LEFT JOIN member_profiles p ON p.user_id = m.organizer_id").
		Select("to_char(m.start_time::date, 'YYYY-MM-DD') AS k, to_char(m.start_time::date, 'YYYY-MM-DD') AS l, COUNT(*)::float AS v")
	db = applyDept(db, "p.department_id", q)
	db = applyRange(db, "m.start_time", q)
	return scanBuckets(db.Group("k, l").Order("k"))
}

func (r *statsRepo) TaskByStatus(ctx context.Context, q Query) ([]Bucket, error) {
	db := r.db.WithContext(ctx).Table("tasks").Where("deleted_at IS NULL").
		Select("status::text AS k, CASE status WHEN 0 THEN '待处理' WHEN 1 THEN '进行中' WHEN 2 THEN '已完成' WHEN 3 THEN '已取消' WHEN 4 THEN '已挂起' ELSE status::text END AS l, COUNT(*)::float AS v")
	db = applyDept(db, "department_id", q)
	db = applyRange(db, "created_at", q)
	return scanBuckets(db.Group("status").Order("status"))
}

func (r *statsRepo) TaskOnTimeRate(ctx context.Context, q Query) (float64, error) {
	type agg struct {
		Done   int64
		OnTime int64
	}
	var row agg
	db := r.db.WithContext(ctx).Table("tasks").Where("deleted_at IS NULL").Where("status = 2")
	db = applyDept(db, "department_id", q)
	db = applyRange(db, "created_at", q)
	err := db.Select("COUNT(*) AS done, COUNT(*) FILTER (WHERE due_date IS NULL OR completed_at IS NULL OR completed_at <= due_date) AS on_time").Scan(&row).Error
	if err != nil || row.Done == 0 {
		return 0, err
	}
	return float64(row.OnTime) / float64(row.Done), nil
}

func (r *statsRepo) TaskTrend(ctx context.Context, q Query) ([]string, map[string][]Bucket, error) {
	expr := truncExpr("created_at", q.Granularity)
	db := r.db.WithContext(ctx).Table("tasks").Where("deleted_at IS NULL").
		Select(expr + " AS k, " + expr + " AS l, status::text AS s, COUNT(*)::float AS v")
	db = applyDept(db, "department_id", q)
	db = applyRange(db, "created_at", q)
	var rows []struct {
		K string  `gorm:"column:k"`
		L string  `gorm:"column:l"`
		S string  `gorm:"column:s"`
		V float64 `gorm:"column:v"`
	}
	if err := db.Group("k, l, s").Order("k, s").Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	statusSet := map[string]struct{}{}
	points := map[string][]Bucket{}
	for _, row := range rows {
		statusSet[row.S] = struct{}{}
		points[row.S] = append(points[row.S], Bucket{Key: row.K, Label: row.L, Value: row.V})
	}
	statuses := make([]string, 0, len(statusSet))
	for s := range statusSet {
		statuses = append(statuses, s)
	}
	return statuses, points, nil
}
