package repo

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Create(ctx context.Context, a *model.Announcement) error
	Update(ctx context.Context, a *model.Announcement) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Announcement, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Announcement, error)
	GetByIDNamed(ctx context.Context, id, viewer uuid.UUID) (*model.AnnouncementNamed, error)
	List(ctx context.Context, viewer uuid.UUID, staff, manage bool, req *dto.ListAnnouncementRequest) ([]model.AnnouncementNamed, int64, error)
	ListDueDrafts(ctx context.Context, now time.Time, limit int) ([]model.Announcement, error)
	ListExpiredPublished(ctx context.Context, now time.Time, limit int) ([]model.Announcement, error)
	MarkRead(ctx context.Context, announcementID, userID uuid.UUID, at time.Time, duration int) error
	UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	ListReaders(ctx context.Context, announcementID uuid.UUID) ([]model.ReaderNamed, error)
	ListNamedUsers(ctx context.Context, ids []uuid.UUID) ([]model.NamedUser, error)
	CountActiveUsers(ctx context.Context) (int64, error)
	ListActiveUserIDs(ctx context.Context) ([]uuid.UUID, error)
	ListActiveUserIDsAmong(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)
	ListActiveUserIDsByDepartments(ctx context.Context, deptIDs []uuid.UUID) ([]uuid.UUID, error)
	ListActiveUserIDsByRoles(ctx context.Context, roleIDs []uuid.UUID) ([]uuid.UUID, error)
	UserInAudience(ctx context.Context, userID uuid.UUID, audienceType string, audienceIDs []string) (bool, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, a *model.Announcement) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repository) Update(ctx context.Context, a *model.Announcement) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Announcement{}, "id = ?", id).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Announcement, error) {
	return r.getByID(ctx, id, false)
}

func (r *repository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Announcement, error) {
	return r.getByID(ctx, id, true)
}

func (r *repository) getByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*model.Announcement, error) {
	var a model.Announcement
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) namedQuery(ctx context.Context, viewer uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Table("announcements AS a").
		Select(`a.*, COALESCE(u.real_name, u.username, '') AS author_name,
			EXISTS(SELECT 1 FROM announcement_reads rd WHERE rd.announcement_id = a.id AND rd.user_id = ?) AS is_read`, viewer).
		Joins("LEFT JOIN users u ON u.id = a.author_id").
		Where("a.deleted_at IS NULL")
}

func (r *repository) GetByIDNamed(ctx context.Context, id, viewer uuid.UUID) (*model.AnnouncementNamed, error) {
	var row model.AnnouncementNamed
	err := r.namedQuery(ctx, viewer).Where("a.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *repository) List(ctx context.Context, viewer uuid.UUID, staff, manage bool, req *dto.ListAnnouncementRequest) ([]model.AnnouncementNamed, int64, error) {
	if req == nil {
		req = &dto.ListAnnouncementRequest{}
	}
	q := r.namedQuery(ctx, viewer)
	q = applyListFilters(q, viewer, staff, manage, req)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var rows []model.AnnouncementNamed
	err := q.Order("a.pinned DESC, a.sort_order ASC, a.published_at DESC NULLS LAST, a.created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

func applyListFilters(q *gorm.DB, viewer uuid.UUID, staff, manage bool, req *dto.ListAnnouncementRequest) *gorm.DB {
	if req.Category != "" {
		q = q.Where("a.category = ?", req.Category)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("a.title ILIKE ? OR a.content ILIKE ?", like, like)
	}
	if req.PinnedOnly {
		q = q.Where("a.pinned = TRUE")
	}
	if req.UnreadOnly {
		q = q.Where("a.status = ? AND NOT EXISTS (SELECT 1 FROM announcement_reads rd WHERE rd.announcement_id = a.id AND rd.user_id = ?)",
			model.StatusPublished, viewer)
	}
	if req.Status != nil {
		q = q.Where("a.status = ?", *req.Status)
		if *req.Status == model.StatusDraft && !manage {
			q = q.Where("a.author_id = ?", viewer)
		}
		if !manage && *req.Status != model.StatusDraft {
			q = q.Where(audienceVisibleSQL(), viewer, viewer, viewer)
		}
		return q
	}
	if manage {
		return q
	}
	visible := "(a.status IN ? AND (" + audienceVisibleSQL() + ")) OR a.author_id = ?"
	args := []interface{}{[]int16{model.StatusPublished, model.StatusArchived}, viewer, viewer, viewer, viewer}
	if !staff {
		visible = "a.status IN ? AND (" + audienceVisibleSQL() + ")"
		args = []interface{}{[]int16{model.StatusPublished, model.StatusArchived}, viewer, viewer, viewer}
	}
	return q.Where(visible, args...)
}

// audienceVisibleSQL 绑定 3 个 viewer：指定用户 / 部门 / 角色。
func audienceVisibleSQL() string {
	return `(
		COALESCE(a.audience_type, 'all') = 'all'
		OR (a.audience_type = 'users' AND a.audience_ids ? ?::text)
		OR (a.audience_type = 'department' AND EXISTS (
			SELECT 1 FROM users u
			WHERE u.id = ? AND u.deleted_at IS NULL AND u.department_id IS NOT NULL
			  AND a.audience_ids ? u.department_id::text
		))
		OR (a.audience_type = 'role' AND EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = ? AND r.status = 0
			  AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
			  AND a.audience_ids ? ur.role_id::text
		))
	)`
}

func (r *repository) ListDueDrafts(ctx context.Context, now time.Time, limit int) ([]model.Announcement, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []model.Announcement
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", model.StatusDraft, now).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *repository) ListExpiredPublished(ctx context.Context, now time.Time, limit int) ([]model.Announcement, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []model.Announcement
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND status = ? AND expires_at IS NOT NULL AND expires_at <= ?", model.StatusPublished, now).
		Order("expires_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *repository) MarkRead(ctx context.Context, announcementID, userID uuid.UUID, at time.Time, duration int) error {
	if duration < 0 {
		duration = 0
	}
	row := &model.AnnouncementRead{
		ID:              uuid.New(),
		AnnouncementID:  announcementID,
		UserID:          userID,
		ReadAt:          at,
		DurationSeconds: duration,
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "announcement_id"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"duration_seconds": gorm.Expr(
					"GREATEST(announcement_reads.duration_seconds, EXCLUDED.duration_seconds)",
				),
			}),
		}).
		Create(row).Error
}

func (r *repository) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("announcements AS a").
		Where("a.deleted_at IS NULL AND a.status = ?", model.StatusPublished).
		Where(audienceVisibleSQL(), userID, userID, userID).
		Where("NOT EXISTS (SELECT 1 FROM announcement_reads rd WHERE rd.announcement_id = a.id AND rd.user_id = ?)", userID).
		Count(&n).Error
	return n, err
}

func (r *repository) ListReaders(ctx context.Context, announcementID uuid.UUID) ([]model.ReaderNamed, error) {
	var rows []model.ReaderNamed
	err := r.db.WithContext(ctx).Table("announcement_reads AS rd").
		Select("rd.user_id, COALESCE(u.real_name, u.username, '') AS real_name, COALESCE(u.username, '') AS username, rd.read_at, rd.duration_seconds").
		Joins("LEFT JOIN users u ON u.id = rd.user_id").
		Where("rd.announcement_id = ?", announcementID).
		Order("rd.read_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *repository) ListNamedUsers(ctx context.Context, ids []uuid.UUID) ([]model.NamedUser, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, COALESCE(real_name, '') AS real_name, COALESCE(username, '') AS username, department_id").
		Where("id IN ? AND deleted_at IS NULL AND status = 0", ids).
		Order("real_name ASC, username ASC").
		Find(&rows).Error
	return rows, err
}

func (r *repository) CountActiveUsers(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("users").
		Where("deleted_at IS NULL AND status = 0").
		Count(&n).Error
	return n, err
}

func (r *repository) ListActiveUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("users").
		Where("deleted_at IS NULL AND status = 0").
		Pluck("id", &ids).Error
	return ids, err
}

func (r *repository) ListActiveUserIDsAmong(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var out []uuid.UUID
	err := r.db.WithContext(ctx).Table("users").
		Where("deleted_at IS NULL AND status = 0 AND id IN ?", ids).
		Pluck("id", &out).Error
	return out, err
}

func (r *repository) ListActiveUserIDsByDepartments(ctx context.Context, deptIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var out []uuid.UUID
	err := r.db.WithContext(ctx).Table("users").
		Where("deleted_at IS NULL AND status = 0 AND department_id IN ?", deptIDs).
		Pluck("id", &out).Error
	return out, err
}

func (r *repository) ListActiveUserIDsByRoles(ctx context.Context, roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var out []uuid.UUID
	err := r.db.WithContext(ctx).Table("users AS u").
		Joins("JOIN user_roles ur ON ur.user_id = u.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("u.deleted_at IS NULL AND u.status = 0 AND r.status = 0").
		Where("ur.role_id IN ? AND (ur.expired_at IS NULL OR ur.expired_at > NOW())", roleIDs).
		Distinct("u.id").
		Pluck("u.id", &out).Error
	return out, err
}

func (r *repository) UserInAudience(ctx context.Context, userID uuid.UUID, audienceType string, audienceIDs []string) (bool, error) {
	switch model.NormalizeAudience(audienceType) {
	case model.AudienceAll:
		return true, nil
	case model.AudienceUsers:
		want := map[string]struct{}{}
		for _, id := range audienceIDs {
			want[id] = struct{}{}
		}
		_, ok := want[userID.String()]
		return ok, nil
	case model.AudienceDepartment:
		ids, err := parseUUIDs(audienceIDs)
		if err != nil {
			return false, err
		}
		out, err := r.ListActiveUserIDsAmong(ctx, []uuid.UUID{userID})
		if err != nil || len(out) == 0 {
			return false, err
		}
		var dept *uuid.UUID
		err = r.db.WithContext(ctx).Table("users").
			Select("department_id").
			Where("id = ? AND deleted_at IS NULL", userID).
			Scan(&dept).Error
		if err != nil || dept == nil {
			return false, err
		}
		for _, id := range ids {
			if id == *dept {
				return true, nil
			}
		}
		return false, nil
	case model.AudienceRole:
		ids, err := parseUUIDs(audienceIDs)
		if err != nil {
			return false, err
		}
		matched, err := r.ListActiveUserIDsByRoles(ctx, ids)
		if err != nil {
			return false, err
		}
		for _, id := range matched {
			if id == userID {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

func parseUUIDs(raw []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
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
	if err != nil {
		return nil, err
	}
	return &u, nil
}
