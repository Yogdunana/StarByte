package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/file/dto"
	"github.com/Yogdunana/StarByte/backend/internal/file/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileRepo 文件元数据访问
type FileRepo interface {
	Create(ctx context.Context, file *model.File) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.File, error)
	GetByIDWithUploader(ctx context.Context, id uuid.UUID) (*model.FileWithUploader, error)
	List(ctx context.Context, req *dto.ListFilesRequest) ([]model.FileWithUploader, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type fileRepo struct {
	db *gorm.DB
}

// NewFileRepo 创建文件 repo
func NewFileRepo(db *gorm.DB) FileRepo {
	return &fileRepo{db: db}
}

func (r *fileRepo) Create(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&file).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepo) GetByIDWithUploader(ctx context.Context, id uuid.UUID) (*model.FileWithUploader, error) {
	var row model.FileWithUploader
	err := r.withUploader(r.db.WithContext(ctx)).Where("files.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *fileRepo) List(ctx context.Context, req *dto.ListFilesRequest) ([]model.FileWithUploader, int64, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	query := r.withUploader(r.db.WithContext(ctx).Model(&model.File{}))
	if req.Category != "" {
		query = query.Where("files.category = ?", req.Category)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("files.name ILIKE ? OR files.original_name ILIKE ?", like, like)
	}
	if req.UploaderID != "" {
		query = query.Where("files.uploaded_by = ?", req.UploaderID)
	}
	if req.MimeType != "" {
		query = query.Where("files.mime_type = ?", req.MimeType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.FileWithUploader
	err := query.Order("files.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *fileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.File{}).Error
}

func (r *fileRepo) withUploader(db *gorm.DB) *gorm.DB {
	return db.Table("files").
		Select("files.*, COALESCE(NULLIF(users.real_name, ''), users.username, '') AS uploader_name").
		Joins("LEFT JOIN users ON users.id = files.uploaded_by")
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
