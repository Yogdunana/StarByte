package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

type AdmissionRepo interface {
	Objections(context.Context, uuid.UUID) ([]model.AdmissionObjection, error)
	OpenObjection(context.Context, uuid.UUID) (*model.AdmissionObjection, error)
	LockApplication(context.Context, uuid.UUID) (*model.MemberApplication, error)
	Actor(context.Context, uuid.UUID) (*model.AdmissionActor, error)
	ParentDepartment(context.Context, uuid.UUID) (*uuid.UUID, error)
	Signatures(context.Context, uuid.UUID) ([]model.AdmissionSignature, error)
	AddSignature(context.Context, *model.AdmissionSignature) error
	InterviewCompleted(context.Context, uuid.UUID, int16, time.Time) (bool, error)
	SaveApplication(context.Context, *model.MemberApplication) error
	WorkflowDefinitionKey(context.Context, uuid.UUID) (string, error)
}
type admissionRepo struct{ db *gorm.DB }

func NewAdmissionRepo(db *gorm.DB) AdmissionRepo { return &admissionRepo{db: db} }
func (r *admissionRepo) LockApplication(ctx context.Context, id uuid.UUID) (*model.MemberApplication, error) {
	var app model.MemberApplication
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&app, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &app, err
}
func (r *admissionRepo) Actor(ctx context.Context, id uuid.UUID) (*model.AdmissionActor, error) {
	var actor model.AdmissionActor
	err := r.db.WithContext(ctx).Table("users").Select("id,department_id").Where("id = ? AND status = 0 AND deleted_at IS NULL", id).Take(&actor).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Table("user_roles ur").Select("r.code").Joins("JOIN roles r ON r.id = ur.role_id").Where("ur.user_id = ? AND r.status = 0 AND (ur.expired_at IS NULL OR ur.expired_at > NOW())", id).Pluck("r.code", &actor.Roles).Error
	return &actor, err
}
func (r *admissionRepo) ParentDepartment(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
	var department struct{ ParentID *uuid.UUID }
	err := r.db.WithContext(ctx).Table("departments").Select("parent_id").Where("id = ? AND status = 0", id).Take(&department).Error
	return department.ParentID, err
}
func (r *admissionRepo) Signatures(ctx context.Context, id uuid.UUID) ([]model.AdmissionSignature, error) {
	rows := []model.AdmissionSignature{}
	err := r.db.WithContext(ctx).Table("admission_signatures s").Select("s.*, u.real_name AS signer_name").Joins("LEFT JOIN users u ON u.id = s.signer_id").Where("application_id = ?", id).Order("s.created_at,s.id").Find(&rows).Error
	return rows, err
}
func (r *admissionRepo) AddSignature(ctx context.Context, item *model.AdmissionSignature) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *admissionRepo) InterviewCompleted(ctx context.Context, id uuid.UUID, round int16, entered time.Time) (bool, error) {
	var count int64
	// Completion and presence are facts. Scores and recommendations are deliberately not consulted.
	err := r.db.WithContext(ctx).Table("interviews").Where("application_id = ? AND round = ? AND status = 3 AND actual_start_time >= ? AND actual_end_time >= ?", id, round, entered, entered).Count(&count).Error
	return count > 0, err
}
func (r *admissionRepo) SaveApplication(ctx context.Context, app *model.MemberApplication) error {
	return r.db.WithContext(ctx).Save(app).Error
}
func (r *admissionRepo) WorkflowDefinitionKey(ctx context.Context, instanceID uuid.UUID) (string, error) {
	var key string
	err := r.db.WithContext(ctx).Raw(
		`SELECT d.key FROM flow_instances i JOIN flow_definitions d ON d.id = i.definition_id WHERE i.id = ?`,
		instanceID,
	).Scan(&key).Error
	return key, err
}

func (r *admissionRepo) Objections(ctx context.Context, id uuid.UUID) ([]model.AdmissionObjection, error) {
	return NewAdmissionMaintenanceRepo(r.db).Objections(ctx, id)
}
func (r *admissionRepo) OpenObjection(ctx context.Context, id uuid.UUID) (*model.AdmissionObjection, error) {
	return NewAdmissionMaintenanceRepo(r.db).OpenObjection(ctx, id)
}
