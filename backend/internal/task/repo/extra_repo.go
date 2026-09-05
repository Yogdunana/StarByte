package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogRepo interface {
	Create(ctx context.Context, log *model.TaskLog) error
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]model.TaskLogNamed, error)
}

type logRepo struct{ db *gorm.DB }

func NewLogRepo(db *gorm.DB) LogRepo { return &logRepo{db: db} }

func (r *logRepo) Create(ctx context.Context, log *model.TaskLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *logRepo) ListByTask(ctx context.Context, taskID uuid.UUID) ([]model.TaskLogNamed, error) {
	var rows []model.TaskLogNamed
	err := r.db.WithContext(ctx).Table("task_logs AS l").
		Select("l.*, COALESCE(u.real_name, u.username, '') AS operator_name").
		Joins("LEFT JOIN users u ON u.id = l.operator_id").
		Where("l.task_id = ?", taskID).
		Order("l.created_at DESC").Find(&rows).Error
	return rows, err
}

type CommentRepo interface {
	Create(ctx context.Context, c *model.TaskComment) error
	Update(ctx context.Context, c *model.TaskComment) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.TaskComment, error)
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]model.TaskCommentNamed, error)
}

type commentRepo struct{ db *gorm.DB }

func NewCommentRepo(db *gorm.DB) CommentRepo { return &commentRepo{db: db} }

func (r *commentRepo) Create(ctx context.Context, c *model.TaskComment) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *commentRepo) Update(ctx context.Context, c *model.TaskComment) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *commentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.TaskComment{}, "id = ?", id).Error
}

func (r *commentRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskComment, error) {
	var c model.TaskComment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *commentRepo) ListByTask(ctx context.Context, taskID uuid.UUID) ([]model.TaskCommentNamed, error) {
	var rows []model.TaskCommentNamed
	err := r.db.WithContext(ctx).Table("task_comments AS c").
		Select("c.*, COALESCE(u.real_name, u.username, '') AS author_name").
		Joins("LEFT JOIN users u ON u.id = c.author_id").
		Where("c.task_id = ?", taskID).
		Order("c.created_at ASC").Find(&rows).Error
	return rows, err
}

type AttachmentRepo interface {
	Create(ctx context.Context, a *model.TaskAttachment) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.TaskAttachment, error)
	GetNamed(ctx context.Context, id uuid.UUID) (*model.TaskAttachmentNamed, error)
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]model.TaskAttachmentNamed, error)
}

type attachmentRepo struct{ db *gorm.DB }

func NewAttachmentRepo(db *gorm.DB) AttachmentRepo { return &attachmentRepo{db: db} }

func (r *attachmentRepo) Create(ctx context.Context, a *model.TaskAttachment) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *attachmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.TaskAttachment{}, "id = ?", id).Error
}

func (r *attachmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskAttachment, error) {
	var a model.TaskAttachment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *attachmentRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("task_attachments AS a").
		Select(`a.*, COALESCE(f.original_name, f.name, '') AS file_name,
			COALESCE(f.path, '') AS file_path, COALESCE(f.size, 0) AS file_size,
			COALESCE(f.mime_type, '') AS file_type, COALESCE(a.uploaded_by::text, '') AS uploader_id`).
		Joins("LEFT JOIN files f ON f.id = a.file_id")
}

func (r *attachmentRepo) GetNamed(ctx context.Context, id uuid.UUID) (*model.TaskAttachmentNamed, error) {
	var row model.TaskAttachmentNamed
	err := r.namedQuery(ctx).Where("a.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *attachmentRepo) ListByTask(ctx context.Context, taskID uuid.UUID) ([]model.TaskAttachmentNamed, error) {
	var rows []model.TaskAttachmentNamed
	err := r.namedQuery(ctx).Where("a.task_id = ?", taskID).Order("a.created_at DESC").Find(&rows).Error
	return rows, err
}
