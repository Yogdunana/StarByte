package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/google/uuid"
)

func (r *statsRepo) InternshipRanking(ctx context.Context, q Query) ([]Bucket, error) {
	if q.HideRanking {
		return []Bucket{}, nil
	}
	days := clippedDaysSQL(q)
	db := r.db.WithContext(ctx).Table("internships AS i").
		Joins("JOIN users u ON u.id = i.user_id").
		Select("u.id::text AS k, COALESCE(NULLIF(u.real_name, ''), u.username) AS l, SUM(" + days + ")::float AS v")
	db = applyDept(db, "i.department_id", q)
	db = applyOverlap(db, "i.start_date", "i.end_date", q)
	return scanBuckets(db.Group("u.id, u.real_name, u.username").Having("SUM(" + days + ") > 0").Order("v DESC").Limit(15))
}

func (r *statsRepo) InternshipDeptAvg(ctx context.Context, q Query) ([]Bucket, error) {
	days := clippedDaysSQL(q)
	db := r.db.WithContext(ctx).Table("internships AS i").
		Joins("LEFT JOIN departments d ON d.id = i.department_id").
		Select("COALESCE(i.department_id::text, 'none') AS k, COALESCE(d.name, '未分配') AS l, AVG(" + days + ")::float AS v")
	db = applyDept(db, "i.department_id", q)
	db = applyOverlap(db, "i.start_date", "i.end_date", q)
	return scanBuckets(db.Group("i.department_id, d.name").Order("v DESC"))
}

func (r *statsRepo) InternshipTrend(ctx context.Context, q Query) ([]Bucket, error) {
	clipS := clipStartSQL(q)
	clipE := clipEndSQL(q)
	expr := truncExpr("gs", q.Granularity)
	table := "internships AS i CROSS JOIN LATERAL generate_series(" + clipS + ", (" + clipE + ") - 1, interval '1 day') AS gs"
	db := r.db.WithContext(ctx).Table(table).
		Select(expr + " AS k, " + expr + " AS l, COUNT(*)::float AS v")
	db = applyDept(db, "i.department_id", q)
	db = applyOverlap(db, "i.start_date", "i.end_date", q)
	db = db.Where(clipE + " > " + clipS)
	return scanBuckets(db.Group("k, l").Order("k"))
}

func (r *statsRepo) RankingHidden(ctx context.Context) (bool, error) {
	var raw string
	err := r.db.WithContext(ctx).Table("configs").Select("config_value").Where("config_key = ?", "internship_config").Limit(1).Scan(&raw).Error
	if err != nil {
		return false, err
	}
	if raw == "" {
		return false, nil
	}
	var cfg struct {
		RankingVisible *bool `json:"ranking_visible"`
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil || cfg.RankingVisible == nil {
		return false, nil
	}
	return !*cfg.RankingVisible, nil
}

func (r *statsRepo) Overview(ctx context.Context, userID uuid.UUID) (*dto.OverviewResponse, error) {
	out := &dto.OverviewResponse{TodayMeetings: []dto.OverviewMeeting{}}
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	if err := r.db.WithContext(ctx).Table("member_profiles").Where("status = 0").Count(&out.TotalMembers).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("meetings").Where("start_time >= ? AND start_time < ?", monthStart, monthStart.AddDate(0, 1, 0)).Count(&out.TotalMeetingsThisMonth).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("tasks").Where("status = 1").Count(&out.TotalTasksInProgress).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("internships").Where("status = 0").Count(&out.TotalInternshipsActive).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("member_applications").Where("status IN (0, 1, 2, 5)").Count(&out.PendingApprovals).Error; err != nil {
		return nil, err
	}
	if userID != uuid.Nil {
		if err := r.db.WithContext(ctx).Table("tasks").Where("assignee_id = ? AND status IN (0, 1, 4)", userID).Count(&out.MyTasks.Todo).Error; err != nil {
			return nil, err
		}
		if err := r.db.WithContext(ctx).Table("tasks").
			Where("assignee_id = ? AND status IN (0, 1, 4) AND due_date IS NOT NULL AND due_date < ?", userID, now).
			Count(&out.MyTasks.Overdue).Error; err != nil {
			return nil, err
		}
		if err := r.db.WithContext(ctx).Table("notifications").Where("user_id = ? AND is_read = false", userID).Count(&out.NotificationsUnread).Error; err != nil {
			return nil, err
		}
	}
	var meetings []struct {
		ID        uuid.UUID
		Title     string
		StartTime time.Time
	}
	if err := r.db.WithContext(ctx).Table("meetings").
		Select("id, title, start_time").
		Where("start_time >= ? AND start_time < ?", dayStart, dayEnd).
		Order("start_time").Limit(8).Scan(&meetings).Error; err != nil {
		return nil, err
	}
	for _, m := range meetings {
		out.TodayMeetings = append(out.TodayMeetings, dto.OverviewMeeting{
			ID: m.ID.String(), Title: m.Title, StartTime: m.StartTime.Format(time.RFC3339),
		})
	}
	return out, nil
}
