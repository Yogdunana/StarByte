package repo

import (
	"context"
	"strings"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	CreateCalendar(ctx context.Context, row *model.Calendar) error
	UpdateCalendar(ctx context.Context, row *model.Calendar) error
	DeleteCalendar(ctx context.Context, id uuid.UUID) error
	GetCalendar(ctx context.Context, id uuid.UUID) (*model.Calendar, error)
	GetCalendarNamed(ctx context.Context, id, viewer uuid.UUID) (*model.CalendarNamed, error)
	ListCalendars(ctx context.Context, viewer uuid.UUID, req *dto.ListCalendarRequest, scope *rbacModel.DataScopeCondition) ([]model.CalendarNamed, int64, error)
	PersonalCalendar(ctx context.Context, owner uuid.UUID) (*model.Calendar, error)

	AddMember(ctx context.Context, row *model.CalendarMember) error
	RemoveMember(ctx context.Context, calendarID, userID uuid.UUID) error
	GetMember(ctx context.Context, calendarID, userID uuid.UUID) (*model.CalendarMember, error)
	ListMembers(ctx context.Context, calendarID uuid.UUID) ([]model.CalendarMemberNamed, error)

	CreateEvent(ctx context.Context, row *model.Event) error
	CreateEventWithDetails(ctx context.Context, row *model.Event, attendees []model.Attendee, reminders []model.Reminder) error
	UpdateEvent(ctx context.Context, row *model.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	GetEvent(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetEventNamed(ctx context.Context, id uuid.UUID) (*model.EventNamed, error)
	ListEvents(ctx context.Context, viewer uuid.UUID, req *dto.ListEventRequest, scope *rbacModel.DataScopeCondition) ([]model.EventNamed, int64, error)
	Overlapping(ctx context.Context, calendarID, exclude uuid.UUID, start, end time.Time) ([]model.Event, error)

	ReplaceAttendees(ctx context.Context, eventID uuid.UUID, rows []model.Attendee) error
	ListAttendees(ctx context.Context, eventID uuid.UUID) ([]model.AttendeeNamed, error)
	GetAttendee(ctx context.Context, eventID, userID uuid.UUID) (*model.Attendee, error)
	UpdateAttendee(ctx context.Context, row *model.Attendee) error

	ReplaceReminders(ctx context.Context, eventID uuid.UUID, rows []model.Reminder) error
	ListReminders(ctx context.Context, eventID uuid.UUID) ([]model.Reminder, error)
	ListDueReminders(ctx context.Context, now time.Time, limit int) ([]model.DueReminder, error)
	MarkReminderTriggered(ctx context.Context, id uuid.UUID, at time.Time) error

	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)

	CalendarBySource(ctx context.Context, owner uuid.UUID, source, sourceKey string) (*model.Calendar, error)
	ReplaceOriginEvents(ctx context.Context, calendarID uuid.UUID, origin string, rows []model.Event) error

	GetGoogleAccount(ctx context.Context, userID uuid.UUID) (*model.GoogleAccount, error)
	UpsertGoogleAccount(ctx context.Context, row *model.GoogleAccount) error
	DeleteGoogleAccount(ctx context.Context, userID uuid.UUID) error
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateCalendar(ctx context.Context, row *model.Calendar) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repository) UpdateCalendar(ctx context.Context, row *model.Calendar) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *repository) DeleteCalendar(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Calendar{}, "id = ?", id).Error
}

func (r *repository) GetCalendar(ctx context.Context, id uuid.UUID) (*model.Calendar, error) {
	var row model.Calendar
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) calendarNamed(ctx context.Context, viewer uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Table("calendars AS c").
		Select(`c.*,
			COALESCE(u.real_name, u.username, '') AS owner_name,
			COALESCE(d.name, '') AS department_name,
			COALESCE(m.role, 0) AS member_role`).
		Joins("LEFT JOIN users u ON u.id = c.owner_id").
		Joins("LEFT JOIN departments d ON d.id = c.department_id").
		Joins("LEFT JOIN calendar_members m ON m.calendar_id = c.id AND m.user_id = ?", viewer)
}

func (r *repository) GetCalendarNamed(ctx context.Context, id, viewer uuid.UUID) (*model.CalendarNamed, error) {
	var row model.CalendarNamed
	err := r.calendarNamed(ctx, viewer).Where("c.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) ListCalendars(ctx context.Context, viewer uuid.UUID, req *dto.ListCalendarRequest, scope *rbacModel.DataScopeCondition) ([]model.CalendarNamed, int64, error) {
	q := applyCalendarVisibility(r.calendarNamed(ctx, viewer), viewer, scope)
	if req.CalendarType != nil {
		q = q.Where("c.calendar_type = ?", *req.CalendarType)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("c.name ILIKE ? OR c.description ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.CalendarNamed
	err := q.Order("c.calendar_type ASC, c.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *repository) PersonalCalendar(ctx context.Context, owner uuid.UUID) (*model.Calendar, error) {
	var row model.Calendar
	err := r.db.WithContext(ctx).Where(
		"owner_id = ? AND calendar_type = ? AND source = ?",
		owner, model.CalendarPersonal, model.SourcePersonal,
	).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) AddMember(ctx context.Context, row *model.CalendarMember) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "calendar_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role"}),
	}).Create(row).Error
}

func (r *repository) RemoveMember(ctx context.Context, calendarID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("calendar_id = ? AND user_id = ?", calendarID, userID).Delete(&model.CalendarMember{}).Error
}

func (r *repository) GetMember(ctx context.Context, calendarID, userID uuid.UUID) (*model.CalendarMember, error) {
	var row model.CalendarMember
	err := r.db.WithContext(ctx).Where("calendar_id = ? AND user_id = ?", calendarID, userID).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) ListMembers(ctx context.Context, calendarID uuid.UUID) ([]model.CalendarMemberNamed, error) {
	var rows []model.CalendarMemberNamed
	err := r.db.WithContext(ctx).Table("calendar_members AS m").
		Select("m.*, COALESCE(u.real_name, u.username, '') AS real_name, COALESCE(u.username, '') AS username").
		Joins("LEFT JOIN users u ON u.id = m.user_id").
		Where("m.calendar_id = ?", calendarID).
		Order("m.created_at").Find(&rows).Error
	return rows, err
}

func (r *repository) CreateEvent(ctx context.Context, row *model.Event) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repository) CreateEventWithDetails(ctx context.Context, row *model.Event, attendees []model.Attendee, reminders []model.Reminder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		if len(attendees) > 0 {
			if err := tx.Create(&attendees).Error; err != nil {
				return err
			}
		}
		if len(reminders) > 0 {
			if err := tx.Create(&reminders).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) UpdateEvent(ctx context.Context, row *model.Event) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *repository) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Event{}, "id = ?", id).Error
}

func (r *repository) GetEvent(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var row model.Event
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) eventNamed(ctx context.Context, viewer uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Table("schedule_events AS e").
		Select(`e.*,
			c.name AS calendar_name, c.color AS calendar_color, c.calendar_type, c.source AS calendar_source, c.owner_id, c.department_id,
			COALESCE(u.real_name, u.username, '') AS creator_name,
			COALESCE(cm.role, 0) AS member_role,
			(SELECT COUNT(*) FROM schedule_event_attendees a WHERE a.event_id = e.id) AS attendee_count`).
		Joins("JOIN calendars c ON c.id = e.calendar_id").
		Joins("LEFT JOIN users u ON u.id = e.created_by").
		Joins("LEFT JOIN calendar_members cm ON cm.calendar_id = e.calendar_id AND cm.user_id = ?", viewer)
}

func (r *repository) GetEventNamed(ctx context.Context, id uuid.UUID) (*model.EventNamed, error) {
	var row model.EventNamed
	err := r.eventNamed(ctx, uuid.Nil).Where("e.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) ListEvents(ctx context.Context, viewer uuid.UUID, req *dto.ListEventRequest, scope *rbacModel.DataScopeCondition) ([]model.EventNamed, int64, error) {
	q := applyEventVisibility(r.eventNamed(ctx, viewer), viewer, scope)
	if req.CalendarID != "" {
		q = q.Where("e.calendar_id = ?", req.CalendarID)
	}
	if req.Status != nil {
		q = q.Where("e.status = ?", *req.Status)
	}
	if !req.Start.IsZero() {
		q = q.Where("e.end_at >= ? OR (e.recurrence <> 'none' AND (e.recurrence_until IS NULL OR e.recurrence_until >= ?))", req.Start, req.Start)
	}
	if !req.End.IsZero() {
		q = q.Where("e.start_at <= ?", req.End)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("e.title ILIKE ? OR e.location ILIKE ? OR e.description ILIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.EventNamed
	err := q.Order("e.start_at ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *repository) Overlapping(ctx context.Context, calendarID, exclude uuid.UUID, start, end time.Time) ([]model.Event, error) {
	q := r.db.WithContext(ctx).Where(
		"calendar_id = ? AND status = ? AND start_at < ? AND (end_at > ? OR (recurrence <> 'none' AND recurrence <> '' AND (recurrence_until IS NULL OR recurrence_until >= ?)))",
		calendarID, model.EventConfirmed, end, start, start,
	)
	if exclude != uuid.Nil {
		q = q.Where("id <> ?", exclude)
	}
	var rows []model.Event
	err := q.Find(&rows).Error
	return rows, err
}

func (r *repository) ReplaceAttendees(ctx context.Context, eventID uuid.UUID, rows []model.Attendee) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", eventID).Delete(&model.Attendee{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

func (r *repository) ListAttendees(ctx context.Context, eventID uuid.UUID) ([]model.AttendeeNamed, error) {
	var rows []model.AttendeeNamed
	err := r.db.WithContext(ctx).Table("schedule_event_attendees AS a").
		Select("a.*, COALESCE(u.real_name, u.username, '') AS real_name, COALESCE(u.username, '') AS username").
		Joins("LEFT JOIN users u ON u.id = a.user_id").
		Where("a.event_id = ?", eventID).
		Order("a.created_at").Find(&rows).Error
	return rows, err
}

func (r *repository) GetAttendee(ctx context.Context, eventID, userID uuid.UUID) (*model.Attendee, error) {
	var row model.Attendee
	err := r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) UpdateAttendee(ctx context.Context, row *model.Attendee) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *repository) ReplaceReminders(ctx context.Context, eventID uuid.UUID, rows []model.Reminder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", eventID).Delete(&model.Reminder{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

func (r *repository) ListReminders(ctx context.Context, eventID uuid.UUID) ([]model.Reminder, error) {
	var rows []model.Reminder
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Order("minutes_before").Find(&rows).Error
	return rows, err
}

func (r *repository) ListDueReminders(ctx context.Context, now time.Time, limit int) ([]model.DueReminder, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []model.DueReminder
	err := r.db.WithContext(ctx).Table("schedule_reminders AS r").
		Select("r.*, e.title, e.start_at, e.end_at, e.recurrence, e.recurrence_until, c.owner_id, e.created_by").
		Joins("JOIN schedule_events e ON e.id = r.event_id").
		Joins("JOIN calendars c ON c.id = e.calendar_id").
		Where(
			"e.status = ? AND (r.triggered_at IS NULL OR (e.recurrence <> 'none' AND e.recurrence <> '' AND (e.recurrence_until IS NULL OR e.recurrence_until >= ?)))",
			model.EventConfirmed, now,
		).
		Order("e.start_at").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *repository) MarkReminderTriggered(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Reminder{}).Where("id = ?", id).Update("triggered_at", at).Error
}

func (r *repository) CalendarBySource(ctx context.Context, owner uuid.UUID, source, sourceKey string) (*model.Calendar, error) {
	var row model.Calendar
	q := r.db.WithContext(ctx).Where("owner_id = ? AND source = ?", owner, source)
	if source == model.SourceImport && sourceKey != "" {
		q = q.Where("source_key = ?", sourceKey)
	}
	err := q.First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) ReplaceOriginEvents(ctx context.Context, calendarID uuid.UUID, origin string, rows []model.Event) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("calendar_id = ? AND origin = ?", calendarID, origin).Delete(&model.Event{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.CreateInBatches(rows, 100).Error
	})
}

func (r *repository) GetGoogleAccount(ctx context.Context, userID uuid.UUID) (*model.GoogleAccount, error) {
	var row model.GoogleAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) UpsertGoogleAccount(ctx context.Context, row *model.GoogleAccount) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"calendar_id", "access_token", "refresh_token", "token_expiry", "google_email", "updated_at"}),
	}).Create(row).Error
}

func (r *repository) DeleteGoogleAccount(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.GoogleAccount{}).Error
}

func (r *repository) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var row model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, COALESCE(real_name, username, '') AS real_name, COALESCE(username, '') AS username, department_id").
		Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func applyCalendarVisibility(q *gorm.DB, viewer uuid.UUID, scope *rbacModel.DataScopeCondition) *gorm.DB {
	if scope == nil || scope.IsEmpty() {
		return q
	}
	if scope.IsSelf {
		return q.Where("c.owner_id = ? OR m.user_id = ?", viewer, viewer)
	}
	if scope.Query == "1 = 0" {
		return q.Where("1 = 0")
	}
	deptQ := strings.ReplaceAll(scope.Query, "department_id", "c.department_id")
	return q.Where(
		"c.owner_id = ? OR m.user_id = ? OR (c.calendar_type = ? AND ("+deptQ+"))",
		append([]interface{}{viewer, viewer, model.CalendarDepartment}, scope.Args...)...,
	)
}

func applyEventVisibility(q *gorm.DB, viewer uuid.UUID, scope *rbacModel.DataScopeCondition) *gorm.DB {
	attendee := "EXISTS (SELECT 1 FROM schedule_event_attendees a WHERE a.event_id = e.id AND a.user_id = ?)"
	member := "EXISTS (SELECT 1 FROM calendar_members cm WHERE cm.calendar_id = e.calendar_id AND cm.user_id = ?)"
	if scope == nil || scope.IsEmpty() {
		return q
	}
	if scope.IsSelf {
		return q.Where("c.owner_id = ? OR "+member+" OR "+attendee, viewer, viewer, viewer)
	}
	if scope.Query == "1 = 0" {
		return q.Where("1 = 0")
	}
	deptQ := strings.ReplaceAll(scope.Query, "department_id", "c.department_id")
	return q.Where(
		"c.owner_id = ? OR "+member+" OR "+attendee+" OR (c.calendar_type = ? AND ("+deptQ+"))",
		append([]interface{}{viewer, viewer, viewer, model.CalendarDepartment}, scope.Args...)...,
	)
}

func normalizePage(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return page, size
}
