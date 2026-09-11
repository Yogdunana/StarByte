package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	notification "github.com/Yogdunana/StarByte/backend/internal/notification/model"
)

type PermissionRefresh struct {
	UserID   uuid.UUID
	Revision uuid.UUID
}
type AdmissionJobsRepo interface {
	LockApplicant(context.Context, uuid.UUID) error
	ValidDepartment(context.Context, uuid.UUID) (bool, error)
	QueuePermissionRefresh(context.Context, uuid.UUID) error
	PermissionRefreshes(context.Context) ([]PermissionRefresh, error)
	CompletePermissionRefresh(context.Context, PermissionRefresh) error
	OverdueApplications(context.Context, time.Time) ([]uuid.UUID, error)
	Reviewers(context.Context) ([]model.AdmissionActor, error)
	Notify(context.Context, uuid.UUID, string, string, string, time.Time) error
	MarkReminded(context.Context, *model.MemberApplication) error
}
type admissionJobsRepo struct{ db *gorm.DB }

func NewAdmissionJobsRepo(db *gorm.DB) AdmissionJobsRepo { return &admissionJobsRepo{db} }
func (r *admissionJobsRepo) QueuePermissionRefresh(ctx context.Context, user uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(`INSERT INTO admission_permission_refreshes(user_id,revision) VALUES (?,?)
 ON CONFLICT(user_id) DO UPDATE SET revision=EXCLUDED.revision,created_at=NOW()`, user, uuid.New()).Error
}
func (r *admissionJobsRepo) PermissionRefreshes(ctx context.Context) ([]PermissionRefresh, error) {
	var rows []PermissionRefresh
	err := r.db.WithContext(ctx).Table("admission_permission_refreshes").Order("created_at,user_id").Limit(100).Find(&rows).Error
	return rows, err
}
func (r *admissionJobsRepo) CompletePermissionRefresh(ctx context.Context, item PermissionRefresh) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM admission_permission_refreshes WHERE user_id=? AND revision=?", item.UserID, item.Revision).Error
}
func (r *admissionJobsRepo) OverdueApplications(ctx context.Context, before time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("member_applications a").Where(`a.historical_review_required=false AND a.admission_stage IN ? AND a.stage_entered_at<=?
 AND NOT EXISTS (SELECT 1 FROM admission_reminders r WHERE r.application_id=a.id AND r.revision=a.admission_revision AND r.stage=a.admission_stage)`, []string{model.AdmissionMaterials, model.AdmissionRound1, model.AdmissionRound2, model.AdmissionPresident}, before).Order("a.stage_entered_at,a.id").Limit(100).Pluck("a.id", &ids).Error
	return ids, err
}
func (r *admissionJobsRepo) Reviewers(ctx context.Context) ([]model.AdmissionActor, error) {
	var rows []struct {
		ID           uuid.UUID
		DepartmentID *uuid.UUID
		Code         string
	}
	err := r.db.WithContext(ctx).Table("users u").Select("u.id,u.department_id,r.code").Joins("JOIN user_roles ur ON ur.user_id=u.id JOIN roles r ON r.id=ur.role_id").Where("u.status=0 AND u.deleted_at IS NULL AND r.status=0 AND (ur.expired_at IS NULL OR ur.expired_at>NOW()) AND r.code IN ?", []string{"minister", "vice_president", "center_director", "president"}).Order("u.id,r.code").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := []model.AdmissionActor{}
	indices := map[uuid.UUID]int{}
	for _, row := range rows {
		index, exists := indices[row.ID]
		if !exists {
			index = len(result)
			indices[row.ID] = index
			result = append(result, model.AdmissionActor{ID: row.ID, DepartmentID: row.DepartmentID})
		}
		result[index].Roles = append(result[index].Roles, row.Code)
	}
	return result, nil
}
func (r *admissionJobsRepo) Notify(ctx context.Context, user uuid.UUID, key, title, content string, now time.Time) error {
	item := notification.Notification{ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("admission:"+key+":"+user.String())), UserID: user, Title: title, Content: content, Category: "member", Priority: "high", ActionURL: "/member/application", SenderName: "入会审批", CreatedAt: now, UpdatedAt: now}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error
}
func (r *admissionJobsRepo) MarkReminded(ctx context.Context, app *model.MemberApplication) error {
	return r.db.WithContext(ctx).Exec("INSERT INTO admission_reminders(application_id,revision,stage) VALUES (?,?,?) ON CONFLICT DO NOTHING", app.ID, app.AdmissionRevision, app.AdmissionStage).Error
}

func (r *admissionJobsRepo) LockApplicant(ctx context.Context, id uuid.UUID) error {
	var row struct{ ID uuid.UUID }
	return r.db.WithContext(ctx).Table("users").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND status=0 AND deleted_at IS NULL", id).Take(&row).Error
}
func (r *admissionJobsRepo) ValidDepartment(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("departments").Where("id=? AND status=0 AND parent_id IS NOT NULL", id).Count(&count).Error
	return count == 1, err
}
