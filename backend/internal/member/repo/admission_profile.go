package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdmissionProfileRepo interface {
	Capture(context.Context, uuid.UUID, uuid.UUID) error
	Restore(context.Context, uuid.UUID, uuid.UUID, time.Time) error
}
type admissionProfileRepo struct{ db *gorm.DB }

func NewAdmissionProfileRepo(db *gorm.DB) AdmissionProfileRepo { return &admissionProfileRepo{db} }
func (r *admissionProfileRepo) Capture(ctx context.Context, application, user uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(`INSERT INTO admission_profile_snapshots(application_id,profile)
 VALUES (?,COALESCE((SELECT to_jsonb(p) FROM member_profiles p WHERE p.user_id=?),'null'::jsonb))
 ON CONFLICT(application_id) DO NOTHING`, application, user).Error
}
func (r *admissionProfileRepo) Restore(ctx context.Context, application, user uuid.UUID, now time.Time) error {
	var row struct{ Absent bool }
	err := r.db.WithContext(ctx).Table("admission_profile_snapshots").Select("profile = 'null'::jsonb AS absent").Where("application_id=?", application).Take(&row).Error
	if err != nil {
		return err
	}
	if row.Absent {
		return NewAdmissionMaintenanceRepo(r.db).StopProbation(ctx, user, now)
	}
	// PostgreSQL restores its own date/timestamp representation without a lossy JSON time conversion.
	// Contact information and achievements edited during probation remain untouched.
	return r.db.WithContext(ctx).Exec(`UPDATE member_profiles p SET
 status=old.status,member_type=old.member_type,department_id=old.department_id,
 position_id=old.position_id,join_date=old.join_date,leave_date=old.leave_date,updated_at=?
 FROM admission_profile_snapshots s CROSS JOIN LATERAL jsonb_populate_record(NULL::member_profiles,s.profile) old
 WHERE s.application_id=? AND p.user_id=? AND p.status=3`, now, application, user).Error
}
