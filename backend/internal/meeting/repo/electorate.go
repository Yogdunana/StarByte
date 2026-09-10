package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

type ElectorateRepo interface {
	Candidates(context.Context, uuid.UUID) ([]model.ElectorCandidate, error)
	Create(context.Context, []model.Elector) error
	Get(context.Context, uuid.UUID, uuid.UUID) (*model.Elector, error)
}
type electorateRepo struct{ db *gorm.DB }

func NewElectorateRepo(db *gorm.DB) ElectorateRepo { return &electorateRepo{db} }
func (r *electorateRepo) Candidates(ctx context.Context, meeting uuid.UUID) ([]model.ElectorCandidate, error) {
	var rows []model.ElectorCandidate
	err := r.db.WithContext(ctx).Raw(`SELECT DISTINCT a.user_id, COALESCE(p.code,'') AS position_code,COALESCE(r.code,'') AS role_code
 FROM meeting_attendees a JOIN users u ON u.id=a.user_id AND u.status=0 AND u.deleted_at IS NULL
 LEFT JOIN positions p ON p.id=u.position_id AND p.status=0
 LEFT JOIN user_roles ur ON ur.user_id=u.id AND (ur.expired_at IS NULL OR ur.expired_at>NOW())
 LEFT JOIN roles r ON r.id=ur.role_id AND r.status=0
 WHERE a.meeting_id=?`, meeting).Scan(&rows).Error
	return rows, err
}
func (r *electorateRepo) Create(ctx context.Context, rows []model.Elector) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(rows, 500).Error
}
func (r *electorateRepo) Get(ctx context.Context, vote, user uuid.UUID) (*model.Elector, error) {
	var row model.Elector
	err := r.db.WithContext(ctx).Where("vote_id=? AND user_id=?", vote, user).Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}
