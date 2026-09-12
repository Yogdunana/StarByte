package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	CreateCategory(ctx context.Context, c *model.Category) error
	UpdateCategory(ctx context.Context, c *model.Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	GetCategory(ctx context.Context, id uuid.UUID) (*model.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*model.Category, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	CountDocsInCategory(ctx context.Context, id uuid.UUID) (int64, error)

	CreateDoc(ctx context.Context, d *model.Doc) error
	UpdateDoc(ctx context.Context, d *model.Doc) error
	DeleteDoc(ctx context.Context, id uuid.UUID) error
	GetDoc(ctx context.Context, id uuid.UUID) (*model.Doc, error)
	GetDocBySlug(ctx context.Context, kind, slug string) (*model.Doc, error)
	GetDocNamed(ctx context.Context, id uuid.UUID) (*model.DocNamed, error)
	GetDocNamedBySlug(ctx context.Context, kind, slug string) (*model.DocNamed, error)
	ListDocs(ctx context.Context, req *dto.ListDocRequest, scope ReadableScope) ([]model.DocNamed, int64, error)
	SearchDocs(ctx context.Context, query, kind string, page, pageSize int, scope ReadableScope) ([]model.DocNamed, int64, error)
	SlugTaken(ctx context.Context, kind, slug string, except *uuid.UUID) (bool, error)

	CreateVersion(ctx context.Context, v *model.DocVersion) error
	ListVersions(ctx context.Context, docID uuid.UUID) ([]model.VersionNamed, error)
	GetVersion(ctx context.Context, docID uuid.UUID, version int) (*model.VersionNamed, error)

	AddAttachment(ctx context.Context, a *model.Attachment) error
	RemoveAttachment(ctx context.Context, docID, fileID uuid.UUID) error
	ListAttachments(ctx context.Context, docID uuid.UUID) ([]model.AttachmentNamed, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateCategory(ctx context.Context, c *model.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *repository) UpdateCategory(ctx context.Context, c *model.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *repository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, "id = ?", id).Error
}

func (r *repository) GetCategory(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var c model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &c, err
}

func (r *repository) GetCategoryBySlug(ctx context.Context, slug string) (*model.Category, error) {
	var c model.Category
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &c, err
}

func (r *repository) ListCategories(ctx context.Context) ([]model.Category, error) {
	var rows []model.Category
	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *repository) CountDocsInCategory(ctx context.Context, id uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Doc{}).Where("category_id = ?", id).Count(&n).Error
	return n, err
}

func (r *repository) CreateDoc(ctx context.Context, d *model.Doc) error {
	return r.db.WithContext(ctx).Omit("SearchTSV").Create(d).Error
}

func (r *repository) UpdateDoc(ctx context.Context, d *model.Doc) error {
	return r.db.WithContext(ctx).Omit("SearchTSV").Save(d).Error
}

func (r *repository) DeleteDoc(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Doc{}, "id = ?", id).Error
}

func (r *repository) GetDoc(ctx context.Context, id uuid.UUID) (*model.Doc, error) {
	var d model.Doc
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &d, err
}

func (r *repository) GetDocBySlug(ctx context.Context, kind, slug string) (*model.Doc, error) {
	var d model.Doc
	q := r.db.WithContext(ctx).Where("slug = ?", slug)
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	err := q.First(&d).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &d, err
}

func (r *repository) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("knowledge_docs AS d").
		Select(`d.*, COALESCE(u.real_name, u.username, '') AS author_name,
			COALESCE(c.name, '') AS category_name`).
		Joins("LEFT JOIN users u ON u.id = d.author_id").
		Joins("LEFT JOIN knowledge_categories c ON c.id = d.category_id AND c.deleted_at IS NULL").
		Where("d.deleted_at IS NULL")
}

func (r *repository) GetDocNamed(ctx context.Context, id uuid.UUID) (*model.DocNamed, error) {
	var row model.DocNamed
	err := r.namedQuery(ctx).Where("d.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) GetDocNamedBySlug(ctx context.Context, kind, slug string) (*model.DocNamed, error) {
	var row model.DocNamed
	q := r.namedQuery(ctx).Where("d.slug = ?", slug)
	if kind != "" {
		q = q.Where("d.kind = ?", kind)
	}
	err := q.First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) ListDocs(ctx context.Context, req *dto.ListDocRequest, scope ReadableScope) ([]model.DocNamed, int64, error) {
	if req == nil {
		req = &dto.ListDocRequest{}
	}
	q := applyReadable(r.namedQuery(ctx), scope)
	if req.Kind != "" {
		q = q.Where("d.kind = ?", req.Kind)
	}
	if req.Status != nil {
		q = q.Where("d.status = ?", *req.Status)
	}
	if req.Visibility != "" {
		q = q.Where("d.visibility = ?", req.Visibility)
	}
	if req.CategoryID != "" {
		q = q.Where("d.category_id = ?", req.CategoryID)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("d.title ILIKE ? OR d.summary ILIKE ? OR d.content ILIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	var rows []model.DocNamed
	q = q.Order("d.updated_at DESC")
	if pageSize < 0 {
		err := q.Find(&rows).Error
		return rows, total, err
	}
	if pageSize == 0 {
		pageSize = 20
	}
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *repository) SearchDocs(ctx context.Context, query, kind string, page, pageSize int, scope ReadableScope) ([]model.DocNamed, int64, error) {
	q := applyReadable(r.namedQuery(ctx), scope)
	if kind != "" {
		q = q.Where("d.kind = ?", kind)
	}
	kw := strings.TrimSpace(query)
	if kw != "" {
		like := "%" + kw + "%"
		q = q.Where(
			"d.search_tsv @@ plainto_tsquery('simple', ?) OR d.title ILIKE ? OR d.summary ILIKE ? OR d.content ILIKE ?",
			kw, like, like, like,
		)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var rows []model.DocNamed
	err := q.Order("d.updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *repository) SlugTaken(ctx context.Context, kind, slug string, except *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Doc{}).Where("kind = ? AND slug = ?", kind, slug)
	if except != nil {
		q = q.Where("id <> ?", *except)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *repository) CreateVersion(ctx context.Context, v *model.DocVersion) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(v).Error
}

func (r *repository) ListVersions(ctx context.Context, docID uuid.UUID) ([]model.VersionNamed, error) {
	var rows []model.VersionNamed
	err := r.db.WithContext(ctx).Table("knowledge_doc_versions AS v").
		Select("v.*, COALESCE(u.real_name, u.username, '') AS editor_name").
		Joins("LEFT JOIN users u ON u.id = v.editor_id").
		Where("v.doc_id = ?", docID).
		Order("v.version DESC").
		Find(&rows).Error
	return rows, err
}

func (r *repository) GetVersion(ctx context.Context, docID uuid.UUID, version int) (*model.VersionNamed, error) {
	var row model.VersionNamed
	err := r.db.WithContext(ctx).Table("knowledge_doc_versions AS v").
		Select("v.*, COALESCE(u.real_name, u.username, '') AS editor_name").
		Joins("LEFT JOIN users u ON u.id = v.editor_id").
		Where("v.doc_id = ? AND v.version = ?", docID, version).
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) AddAttachment(ctx context.Context, a *model.Attachment) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(a).Error
}

func (r *repository) RemoveAttachment(ctx context.Context, docID, fileID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("doc_id = ? AND file_id = ?", docID, fileID).Delete(&model.Attachment{}).Error
}

func (r *repository) ListAttachments(ctx context.Context, docID uuid.UUID) ([]model.AttachmentNamed, error) {
	var rows []model.AttachmentNamed
	err := r.db.WithContext(ctx).Table("knowledge_attachments AS a").
		Select("a.*, COALESCE(f.original_name, f.name, '') AS file_name, COALESCE(f.size, 0) AS file_size").
		Joins("LEFT JOIN files f ON f.id = a.file_id").
		Where("a.doc_id = ?", docID).
		Order("a.created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *repository) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var u model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, COALESCE(real_name, '') AS real_name, COALESCE(username, '') AS username").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &u, err
}
