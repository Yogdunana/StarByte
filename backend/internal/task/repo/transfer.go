package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

type TransferRepo interface {
	Create(context.Context, *model.TaskTransfer) error
	Save(context.Context, *model.TaskTransfer) error
	Get(context.Context, uuid.UUID) (*model.TaskTransfer, error)
	Lock(context.Context, uuid.UUID) (*model.TaskTransfer, error)
	Pending(context.Context, uuid.UUID) (*model.TaskTransfer, error)
	Signatures(context.Context, uuid.UUID) ([]model.TransferSignature, error)
	Sign(context.Context, *model.TransferSignature) error
	Actors(context.Context) ([]model.TransferActor, error)
	Actor(context.Context, uuid.UUID) (*model.TransferActor, error)
	Department(context.Context, uuid.UUID) (*model.TransferDepartment, error)
}
type transferRepo struct{ db *gorm.DB }

func NewTransferRepo(db *gorm.DB) TransferRepo { return &transferRepo{db} }
func (r *transferRepo) Create(ctx context.Context, t *model.TaskTransfer) error {
	return r.db.WithContext(ctx).Create(t).Error
}
func (r *transferRepo) Save(ctx context.Context, t *model.TaskTransfer) error {
	return r.db.WithContext(ctx).Save(t).Error
}
func (r *transferRepo) Get(ctx context.Context, id uuid.UUID) (*model.TaskTransfer, error) {
	return transferRow(r.db.WithContext(ctx).Where("id=?", id))
}
func (r *transferRepo) Lock(ctx context.Context, id uuid.UUID) (*model.TaskTransfer, error) {
	return transferRow(r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id))
}
func (r *transferRepo) Pending(ctx context.Context, id uuid.UUID) (*model.TaskTransfer, error) {
	return transferRow(r.db.WithContext(ctx).Where("task_id=? AND status='pending'", id))
}
func transferRow(q *gorm.DB) (*model.TaskTransfer, error) {
	var row model.TaskTransfer
	err := q.First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}
func (r *transferRepo) Signatures(ctx context.Context, id uuid.UUID) ([]model.TransferSignature, error) {
	rows := []model.TransferSignature{}
	err := r.db.WithContext(ctx).Where("transfer_id=?", id).Order("created_at,id").Find(&rows).Error
	return rows, err
}
func (r *transferRepo) Sign(ctx context.Context, s *model.TransferSignature) error {
	return r.db.WithContext(ctx).Create(s).Error
}
func (r *transferRepo) Department(ctx context.Context, id uuid.UUID) (*model.TransferDepartment, error) {
	var d model.TransferDepartment
	err := r.db.WithContext(ctx).Table("departments").Where("id=? AND status=0", id).First(&d).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &d, err
}
func (r *transferRepo) Actors(ctx context.Context) ([]model.TransferActor, error) {
	type row struct {
		ID           uuid.UUID
		DepartmentID *uuid.UUID
		Code         string
	}
	rows := []row{}
	err := r.db.WithContext(ctx).Table("users u").Select("u.id,u.department_id,r.code").Joins("JOIN user_roles ur ON ur.user_id=u.id JOIN roles r ON r.id=ur.role_id").Where("u.status=0 AND u.deleted_at IS NULL AND r.status=0 AND (ur.expired_at IS NULL OR ur.expired_at>NOW()) AND r.code IN ?", []string{"minister", "center_director", "vice_president", "president"}).Order("u.id,r.code").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := []model.TransferActor{}
	indices := map[uuid.UUID]int{}
	for _, r := range rows {
		idx, ok := indices[r.ID]
		if !ok {
			idx = len(out)
			indices[r.ID] = idx
			out = append(out, model.TransferActor{ID: r.ID, DepartmentID: r.DepartmentID})
		}
		out[idx].Roles = append(out[idx].Roles, r.Code)
	}
	return out, nil
}
func (r *transferRepo) Actor(ctx context.Context, id uuid.UUID) (*model.TransferActor, error) {
	var a model.TransferActor
	err := r.db.WithContext(ctx).Table("users").Select("id,department_id").Where("id=? AND status=0 AND deleted_at IS NULL", id).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Table("user_roles ur").Select("r.code").Joins("JOIN roles r ON r.id=ur.role_id").Where("ur.user_id=? AND r.status=0 AND (ur.expired_at IS NULL OR ur.expired_at>NOW())", id).Pluck("r.code", &a.Roles).Error
	return &a, err
}
