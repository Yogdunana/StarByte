package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VoteRepo interface {
	CreateVote(ctx context.Context, v *model.Vote, options []model.VoteOption) error
	UpdateVote(ctx context.Context, v *model.Vote) error
	GetVote(ctx context.Context, id uuid.UUID) (*model.Vote, error)
	ListByMeeting(ctx context.Context, meetingID uuid.UUID) ([]model.Vote, error)
	ListOptions(ctx context.Context, voteID uuid.UUID) ([]model.VoteOption, error)
	GetOptionByKey(ctx context.Context, voteID uuid.UUID, key string) (*model.VoteOption, error)
	CreateRecord(ctx context.Context, rec *model.VoteRecord) error
	GetRecord(ctx context.Context, voteID, userID uuid.UUID) (*model.VoteRecord, error)
	ListRecords(ctx context.Context, voteID uuid.UUID) ([]model.VoteRecordNamed, error)
	GetConfig(ctx context.Context, key string) (*model.SystemConfig, error)
	UpsertConfig(ctx context.Context, cfg *model.SystemConfig) error
}

type voteRepo struct{ db *gorm.DB }

func NewVoteRepo(db *gorm.DB) VoteRepo {
	return &voteRepo{db: db}
}

func (r *voteRepo) CreateVote(ctx context.Context, v *model.Vote, options []model.VoteOption) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(v).Error; err != nil {
			return err
		}
		if len(options) == 0 {
			return nil
		}
		return tx.Create(&options).Error
	})
}

func (r *voteRepo) UpdateVote(ctx context.Context, v *model.Vote) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *voteRepo) GetVote(ctx context.Context, id uuid.UUID) (*model.Vote, error) {
	var v model.Vote
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *voteRepo) ListByMeeting(ctx context.Context, meetingID uuid.UUID) ([]model.Vote, error) {
	var rows []model.Vote
	err := r.db.WithContext(ctx).Where("meeting_id = ?", meetingID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *voteRepo) ListOptions(ctx context.Context, voteID uuid.UUID) ([]model.VoteOption, error) {
	var rows []model.VoteOption
	err := r.db.WithContext(ctx).Where("vote_id = ?", voteID).Order("sort_order ASC").Find(&rows).Error
	return rows, err
}

func (r *voteRepo) GetOptionByKey(ctx context.Context, voteID uuid.UUID, key string) (*model.VoteOption, error) {
	var o model.VoteOption
	err := r.db.WithContext(ctx).Where("vote_id = ? AND option_key = ?", voteID, key).First(&o).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *voteRepo) CreateRecord(ctx context.Context, rec *model.VoteRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *voteRepo) GetRecord(ctx context.Context, voteID, userID uuid.UUID) (*model.VoteRecord, error) {
	var rec model.VoteRecord
	err := r.db.WithContext(ctx).Where("vote_id = ? AND voter_id = ?", voteID, userID).First(&rec).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *voteRepo) ListRecords(ctx context.Context, voteID uuid.UUID) ([]model.VoteRecordNamed, error) {
	var rows []model.VoteRecordNamed
	err := r.db.WithContext(ctx).Table("meeting_vote_records AS r").
		Select("r.*, COALESCE(u.real_name, u.username, '') AS voter_name").
		Joins("LEFT JOIN users u ON u.id = r.voter_id").
		Where("r.vote_id = ?", voteID).Find(&rows).Error
	return rows, err
}

func (r *voteRepo) GetConfig(ctx context.Context, key string) (*model.SystemConfig, error) {
	var c model.SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *voteRepo) UpsertConfig(ctx context.Context, cfg *model.SystemConfig) error {
	if cfg.ID == uuid.Nil {
		cfg.ID = uuid.New()
		return r.db.WithContext(ctx).Create(cfg).Error
	}
	return r.db.WithContext(ctx).Save(cfg).Error
}
