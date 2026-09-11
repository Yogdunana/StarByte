package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RegistrationRepo interface {
	Create(ctx context.Context, r *model.ActivityRegistration) error
	Update(ctx context.Context, r *model.ActivityRegistration) error
	MarkCheckedIn(ctx context.Context, id uuid.UUID, at time.Time, method int16, lat, lng *float64) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ActivityRegistration, error)
	GetByActivityAndUser(ctx context.Context, activityID, userID uuid.UUID) (*model.ActivityRegistration, error)
	GetByActivityAndUserForUpdate(ctx context.Context, activityID, userID uuid.UUID) (*model.ActivityRegistration, error)
	ListByActivity(ctx context.Context, activityID uuid.UUID) ([]model.RegistrationNamed, error)
	CountByActivityAndStatus(ctx context.Context, activityID uuid.UUID, status int16) (int64, error)
	CountCheckedIn(ctx context.Context, activityID uuid.UUID) (int64, error)
	ListWaitlist(ctx context.Context, activityID uuid.UUID) ([]model.ActivityRegistration, error)
}

type registrationRepo struct{ db *gorm.DB }

func NewRegistrationRepo(db *gorm.DB) RegistrationRepo {
	return &registrationRepo{db: db}
}

func (r *registrationRepo) Create(ctx context.Context, reg *model.ActivityRegistration) error {
	return r.db.WithContext(ctx).Create(reg).Error
}

func (r *registrationRepo) Update(ctx context.Context, reg *model.ActivityRegistration) error {
	return r.db.WithContext(ctx).Save(reg).Error
}

func (r *registrationRepo) MarkCheckedIn(ctx context.Context, id uuid.UUID, at time.Time, method int16, lat, lng *float64) (int64, error) {
	updates := map[string]interface{}{
		"checkin_status": model.CheckinDone,
		"checked_in_at":  at,
		"checkin_method": method,
		"updated_at":     at,
	}
	if lat != nil && lng != nil {
		updates["gps_latitude"] = *lat
		updates["gps_longitude"] = *lng
	}
	res := r.db.WithContext(ctx).Model(&model.ActivityRegistration{}).
		Where("id = ? AND status = ? AND checkin_status = ?", id, model.RegApproved, model.CheckinPending).
		Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *registrationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.ActivityRegistration, error) {
	var reg model.ActivityRegistration
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&reg).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *registrationRepo) GetByActivityAndUser(ctx context.Context, activityID, userID uuid.UUID) (*model.ActivityRegistration, error) {
	return r.getByActivityAndUser(ctx, activityID, userID, false)
}

func (r *registrationRepo) GetByActivityAndUserForUpdate(ctx context.Context, activityID, userID uuid.UUID) (*model.ActivityRegistration, error) {
	return r.getByActivityAndUser(ctx, activityID, userID, true)
}

func (r *registrationRepo) getByActivityAndUser(ctx context.Context, activityID, userID uuid.UUID, forUpdate bool) (*model.ActivityRegistration, error) {
	var reg model.ActivityRegistration
	q := r.db.WithContext(ctx).Where("activity_id = ? AND user_id = ?", activityID, userID)
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&reg).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *registrationRepo) ListByActivity(ctx context.Context, activityID uuid.UUID) ([]model.RegistrationNamed, error) {
	var rows []model.RegistrationNamed
	err := r.db.WithContext(ctx).Table("activity_registrations AS r").
		Select("r.*, COALESCE(u.real_name, u.username, '') AS real_name, u.username").
		Joins("LEFT JOIN users u ON u.id = r.user_id").
		Where("r.activity_id = ?", activityID).
		Order("r.created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *registrationRepo) CountByActivityAndStatus(ctx context.Context, activityID uuid.UUID, status int16) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ActivityRegistration{}).
		Where("activity_id = ? AND status = ?", activityID, status).
		Count(&count).Error
	return count, err
}

func (r *registrationRepo) CountCheckedIn(ctx context.Context, activityID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ActivityRegistration{}).
		Where("activity_id = ? AND checkin_status = ?", activityID, model.CheckinDone).
		Count(&count).Error
	return count, err
}

func (r *registrationRepo) ListWaitlist(ctx context.Context, activityID uuid.UUID) ([]model.ActivityRegistration, error) {
	var rows []model.ActivityRegistration
	err := r.db.WithContext(ctx).
		Where("activity_id = ? AND status = ?", activityID, model.RegWaitlist).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

// SurveyRepo 满意度调查
type SurveyRepo interface {
	Create(ctx context.Context, s *model.ActivitySurvey) error
	GetByActivityAndUser(ctx context.Context, activityID, userID uuid.UUID) (*model.ActivitySurvey, error)
	StatsByActivity(ctx context.Context, activityID uuid.UUID) (count int64, avgRating float64, dist map[int16]int64, err error)
}

type surveyRepo struct{ db *gorm.DB }

func NewSurveyRepo(db *gorm.DB) SurveyRepo {
	return &surveyRepo{db: db}
}

func (r *surveyRepo) Create(ctx context.Context, s *model.ActivitySurvey) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *surveyRepo) GetByActivityAndUser(ctx context.Context, activityID, userID uuid.UUID) (*model.ActivitySurvey, error) {
	var s model.ActivitySurvey
	err := r.db.WithContext(ctx).
		Where("activity_id = ? AND user_id = ?", activityID, userID).
		First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *surveyRepo) StatsByActivity(ctx context.Context, activityID uuid.UUID) (int64, float64, map[int16]int64, error) {
	var rows []model.ActivitySurvey
	if err := r.db.WithContext(ctx).
		Where("activity_id = ?", activityID).
		Find(&rows).Error; err != nil {
		return 0, 0, nil, err
	}
	dist := make(map[int16]int64)
	var totalRating float64
	for _, s := range rows {
		dist[s.Rating]++
		totalRating += float64(s.Rating)
	}
	count := int64(len(rows))
	avg := 0.0
	if count > 0 {
		avg = totalRating / float64(count)
	}
	return count, avg, dist, nil
}
