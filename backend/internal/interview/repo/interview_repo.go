package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InterviewRepo interface {
	Create(ctx context.Context, iv *model.Interview) error
	Update(ctx context.Context, iv *model.Interview) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Interview, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.InterviewWithNames, error)
	List(ctx context.Context, req *dto.ListInterviewRequest, scope *rbacModel.DataScopeCondition) ([]model.InterviewWithNames, int64, error)
	ListMine(ctx context.Context, userID uuid.UUID, status *int16) ([]model.InterviewWithNames, error)
	FindBySessionApplicant(ctx context.Context, sessionID, applicantID uuid.UUID) (*model.Interview, error)
	ReplaceInterviewers(ctx context.Context, interviewID uuid.UUID, rows []model.Interviewer) error
	ListInterviewers(ctx context.Context, interviewIDs []uuid.UUID) ([]model.InterviewerNamed, error)
	HasInterviewerConflict(ctx context.Context, interviewerID uuid.UUID, start, end time.Time, exclude uuid.UUID) (bool, error)
	MarkAbsentBySession(ctx context.Context, sessionID uuid.UUID) error
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
	GetApplication(ctx context.Context, id uuid.UUID) (*model.ApplicationBrief, error)
}

type interviewRepo struct{ db *gorm.DB }

func NewInterviewRepo(db *gorm.DB) InterviewRepo {
	return &interviewRepo{db: db}
}

func (r *interviewRepo) Create(ctx context.Context, iv *model.Interview) error {
	return r.db.WithContext(ctx).Create(iv).Error
}

func (r *interviewRepo) Update(ctx context.Context, iv *model.Interview) error {
	return r.db.WithContext(ctx).Save(iv).Error
}

func (r *interviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Interview, error) {
	var iv model.Interview
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&iv).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &iv, nil
}

func (r *interviewRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("interviews AS i").
		Select(`i.*, COALESCE(a.real_name, u.real_name, '') AS applicant_name,
			COALESCE(a.student_no, '') AS student_no,
			COALESCE(s.title, '') AS session_title,
			s.department_id,
			COALESCE(d.name, '') AS department_name`).
		Joins("LEFT JOIN interview_sessions s ON s.id = i.session_id").
		Joins("LEFT JOIN users u ON u.id = i.applicant_id").
		Joins("LEFT JOIN member_applications a ON a.id = i.application_id").
		Joins("LEFT JOIN departments d ON d.id = s.department_id")
}

func (r *interviewRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.InterviewWithNames, error) {
	var row model.InterviewWithNames
	err := r.namedQuery(ctx).Where("i.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *interviewRepo) List(ctx context.Context, req *dto.ListInterviewRequest, scope *rbacModel.DataScopeCondition) ([]model.InterviewWithNames, int64, error) {
	q := r.namedQuery(ctx)
	q = applyScope(q, scope)
	q = filterInterviews(q, req)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.InterviewWithNames
	err := q.Order("i.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func filterInterviews(q *gorm.DB, req *dto.ListInterviewRequest) *gorm.DB {
	if req.SessionID != "" {
		q = q.Where("i.session_id = ?", req.SessionID)
	}
	if req.Status != nil {
		q = q.Where("i.status = ?", *req.Status)
	}
	if req.Result != nil {
		q = q.Where("i.result_code = ?", *req.Result)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		q = q.Where("a.real_name ILIKE ? OR a.student_no ILIKE ? OR u.real_name ILIKE ?", like, like, like)
	}
	return q
}

func (r *interviewRepo) ListMine(ctx context.Context, userID uuid.UUID, status *int16) ([]model.InterviewWithNames, error) {
	q := r.namedQuery(ctx).Where(
		"i.applicant_id = ? OR i.id IN (SELECT interview_id FROM interview_interviewers WHERE interviewer_id = ?)",
		userID, userID,
	)
	if status != nil {
		q = q.Where("i.status = ?", *status)
	}
	var rows []model.InterviewWithNames
	err := q.Order("i.scheduled_at DESC NULLS LAST, i.created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *interviewRepo) FindBySessionApplicant(ctx context.Context, sessionID, applicantID uuid.UUID) (*model.Interview, error) {
	var iv model.Interview
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND applicant_id = ? AND status <> ?", sessionID, applicantID, model.InterviewCancelled).
		First(&iv).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &iv, nil
}

func (r *interviewRepo) ReplaceInterviewers(ctx context.Context, interviewID uuid.UUID, rows []model.Interviewer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("interview_id = ?", interviewID).Delete(&model.Interviewer{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

func (r *interviewRepo) ListInterviewers(ctx context.Context, interviewIDs []uuid.UUID) ([]model.InterviewerNamed, error) {
	if len(interviewIDs) == 0 {
		return nil, nil
	}
	var rows []model.InterviewerNamed
	err := r.db.WithContext(ctx).Table("interview_interviewers AS ii").
		Select("ii.interview_id, ii.interviewer_id, COALESCE(u.real_name, u.username, '') AS name").
		Joins("LEFT JOIN users u ON u.id = ii.interviewer_id").
		Where("ii.interview_id IN ?", interviewIDs).
		Find(&rows).Error
	return rows, err
}

func (r *interviewRepo) HasInterviewerConflict(ctx context.Context, interviewerID uuid.UUID, start, end time.Time, exclude uuid.UUID) (bool, error) {
	var n int64
	q := r.db.WithContext(ctx).Table("interview_interviewers AS ii").
		Joins("JOIN interviews i ON i.id = ii.interview_id").
		Where("ii.interviewer_id = ? AND i.status NOT IN ? AND i.scheduled_at IS NOT NULL",
			interviewerID, []int16{model.InterviewCancelled, model.InterviewAbsent}).
		Where("i.scheduled_at < ? AND (i.scheduled_at + (COALESCE(NULLIF(i.duration, 0), ?) || ' minutes')::interval) > ?",
			end, model.DefaultDuration, start)
	if exclude != uuid.Nil {
		q = q.Where("i.id <> ?", exclude)
	}
	err := q.Count(&n).Error
	return n > 0, err
}

func (r *interviewRepo) MarkAbsentBySession(ctx context.Context, sessionID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Interview{}).
		Where("session_id = ? AND status IN ?", sessionID, []int16{model.InterviewPending, model.InterviewCheckedIn}).
		Updates(map[string]interface{}{"status": model.InterviewAbsent}).Error
}

func (r *interviewRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var u model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, COALESCE(real_name, '') AS real_name, username").
		Where("id = ?", id).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *interviewRepo) GetApplication(ctx context.Context, id uuid.UUID) (*model.ApplicationBrief, error) {
	var a model.ApplicationBrief
	err := r.db.WithContext(ctx).Table("member_applications").
		Select("id, user_id, real_name, student_no, department_id, status").
		Where("id = ?", id).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}
