package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/internal/activity/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notifier 通知接口
type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

// ActivityService 活动服务接口
type ActivityService interface {
	CreateActivity(ctx context.Context, operator uuid.UUID, req *dto.CreateActivityRequest) (*dto.ActivityResponse, error)
	UpdateActivity(ctx context.Context, id uuid.UUID, req *dto.UpdateActivityRequest) (*dto.ActivityResponse, error)
	DeleteActivity(ctx context.Context, id uuid.UUID) error
	GetActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error)
	// CanAccessActivity 判定 viewer 是否有权访问该活动（数据范围校验）。
	// scope 为 nil 表示不限制；活动不存在返回 (false, nil)。
	CanAccessActivity(ctx context.Context, viewer, activityID uuid.UUID, scope *rbacModel.DataScopeCondition) (bool, error)
	ListActivities(ctx context.Context, viewer uuid.UUID, req *dto.ListActivityRequest, scope *rbacModel.DataScopeCondition) ([]*dto.ActivityResponse, int64, error)

	StartActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error)
	EndActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error)
	CancelActivity(ctx context.Context, id uuid.UUID, reason string) (*dto.ActivityResponse, error)

	Register(ctx context.Context, activityID, userID uuid.UUID) (*dto.RegistrationResponse, error)
	GetMyRegistration(ctx context.Context, activityID, userID uuid.UUID) (*dto.RegistrationResponse, error)
	CancelRegistration(ctx context.Context, activityID, userID uuid.UUID) error
	ApproveRegistration(ctx context.Context, activityID, userID uuid.UUID, approve bool, reason string) (*dto.RegistrationResponse, error)
	ListRegistrations(ctx context.Context, activityID uuid.UUID) ([]dto.RegistrationResponse, error)

	Checkin(ctx context.Context, activityID, userID uuid.UUID, req *dto.CheckinRequest) (*dto.RegistrationResponse, error)
	IssueCheckinQR(ctx context.Context, activityID uuid.UUID) (*dto.QRCodeResponse, error)

	GetStats(ctx context.Context, activityID uuid.UUID) (*dto.ActivityStatsResponse, error)
	SubmitSurvey(ctx context.Context, activityID, userID uuid.UUID, req *dto.SurveyRequest) error
}

type activityService struct {
	activities repo.ActivityRepo
	regs       repo.RegistrationRepo
	surveys    repo.SurveyRepo
	notify     Notifier
	db         *gorm.DB
	now        func() time.Time
	tokenTTL   time.Duration
}

func NewActivityService(
	activities repo.ActivityRepo,
	regs repo.RegistrationRepo,
	surveys repo.SurveyRepo,
	notify Notifier,
	db *gorm.DB,
) ActivityService {
	return &activityService{
		activities: activities,
		regs:       regs,
		surveys:    surveys,
		notify:     notify,
		db:         db,
		now:        time.Now,
		tokenTTL:   time.Duration(model.DefaultCheckinTokenTTLSeconds) * time.Second,
	}
}

func (s *activityService) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

type pendingNotify struct {
	users    []uuid.UUID
	template string
	activity *model.Activity
}

func (s *activityService) flushNotifies(ctx context.Context, notes ...*pendingNotify) {
	for _, n := range notes {
		if n != nil {
			s.notifyActivity(ctx, n.users, n.template, n.activity)
		}
	}
}
