package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

type AdmissionMaintenanceRepo interface {
	GrantRole(context.Context, uuid.UUID, string, *uuid.UUID) error
	DueApplications(context.Context, time.Time) ([]uuid.UUID, error)
	OpenObjection(context.Context, uuid.UUID) (*model.AdmissionObjection, error)
	Objections(context.Context, uuid.UUID) ([]model.AdmissionObjection, error)
	SaveObjection(context.Context, *model.AdmissionObjection) error
	ActivateProfile(context.Context, uuid.UUID, time.Time) error
	StopProbation(context.Context, uuid.UUID, time.Time) error
}
type admissionMaintenanceRepo struct{ db *gorm.DB }

func NewAdmissionMaintenanceRepo(db *gorm.DB) AdmissionMaintenanceRepo {
	return &admissionMaintenanceRepo{db: db}
}
func (r *admissionMaintenanceRepo) DueApplications(ctx context.Context, now time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("member_applications").Where("charter_policy = false AND admission_stage = ? AND historical_review_required = false AND probation_until <= ? AND NOT EXISTS (SELECT 1 FROM admission_objections o WHERE o.application_id=member_applications.id AND o.status IN ('center_review','president_review'))", model.AdmissionProbation, now).Order("probation_until,id").Limit(100).Pluck("id", &ids).Error
	return ids, err
}
func (r *admissionMaintenanceRepo) OpenObjection(ctx context.Context, id uuid.UUID) (*model.AdmissionObjection, error) {
	var objection model.AdmissionObjection
	err := r.db.WithContext(ctx).Where("application_id = ? AND status IN ?", id, []string{"center_review", "president_review"}).First(&objection).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &objection, err
}
func (r *admissionMaintenanceRepo) Objections(ctx context.Context, id uuid.UUID) ([]model.AdmissionObjection, error) {
	rows := []model.AdmissionObjection{}
	err := r.db.WithContext(ctx).Where("application_id = ?", id).Order("created_at,id").Find(&rows).Error
	return rows, err
}
func (r *admissionMaintenanceRepo) SaveObjection(ctx context.Context, item *model.AdmissionObjection) error {
	return r.db.WithContext(ctx).Save(item).Error
}
func (r *admissionMaintenanceRepo) ActivateProfile(ctx context.Context, user uuid.UUID, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&model.MemberProfile{}).Where("user_id = ? AND status = ?", user, model.ProfileProbation).Updates(map[string]interface{}{"status": model.ProfileActive, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *admissionMaintenanceRepo) StopProbation(ctx context.Context, user uuid.UUID, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.MemberProfile{}).Where("user_id = ? AND status = ?", user, model.ProfileProbation).Updates(map[string]interface{}{"status": model.ProfileLeft, "leave_date": now, "updated_at": now}).Error
}

func (r *admissionMaintenanceRepo) GrantRole(ctx context.Context, user uuid.UUID, role string, department *uuid.UUID) error {
	var record struct{ ID uuid.UUID }
	if err := r.db.WithContext(ctx).Table("roles").Where("code = ? AND status = 0", role).Select("id").Take(&record).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Exec("INSERT INTO user_roles(user_id,role_id) VALUES (?,?) ON CONFLICT(user_id,role_id) DO UPDATE SET expired_at=NULL", user, record.ID).Error; err != nil {
		return err
	}
	if role == "probationary" {
		return nil
	}
	return r.db.WithContext(ctx).Table("users").Where("id = ?", user).Update("department_id", department).Error
}
