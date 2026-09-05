package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgendaRepo interface {
	Create(ctx context.Context, a *model.Agenda) error
	Update(ctx context.Context, a *model.Agenda) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Agenda, error)
	ListByMeeting(ctx context.Context, meetingID uuid.UUID) ([]model.Agenda, error)
	SaveSort(ctx context.Context, items []model.Agenda) error
}

type agendaRepo struct{ db *gorm.DB }

func NewAgendaRepo(db *gorm.DB) AgendaRepo {
	return &agendaRepo{db: db}
}

func (r *agendaRepo) Create(ctx context.Context, a *model.Agenda) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *agendaRepo) Update(ctx context.Context, a *model.Agenda) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *agendaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Agenda{}, "id = ?", id).Error
}

func (r *agendaRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Agenda, error) {
	var a model.Agenda
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *agendaRepo) ListByMeeting(ctx context.Context, meetingID uuid.UUID) ([]model.Agenda, error) {
	var rows []model.Agenda
	err := r.db.WithContext(ctx).Where("meeting_id = ?", meetingID).Order("sort_order ASC, created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *agendaRepo) SaveSort(ctx context.Context, items []model.Agenda) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range items {
			if err := tx.Model(&model.Agenda{}).Where("id = ?", items[i].ID).
				Update("sort_order", items[i].SortOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
