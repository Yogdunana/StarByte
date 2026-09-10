package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/google/uuid"
)

// Notifier 通知接口
type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

// ActivityService 活动服务接口
type ActivityService interface {
	// 活动 CRUD
	CreateActivity(ctx context.Context, operator uuid.UUID, req *dto.CreateActivityRequest) (*dto.ActivityResponse, error)
	UpdateActivity(ctx context.Context, id uuid.UUID, req *dto.UpdateActivityRequest) (*dto.ActivityResponse, error)
	DeleteActivity(ctx context.Context, id uuid.UUID) error
	GetActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error)
	ListActivities(ctx context.Context, req *dto.ListActivityRequest) ([]*dto.ActivityResponse, int64, error)

	// 活动状态流转
	StartActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error)
	EndActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error)
	CancelActivity(ctx context.Context, id uuid.UUID, reason string) (*dto.ActivityResponse, error)

	// 报名管理
	Register(ctx context.Context, activityID, userID uuid.UUID) (*dto.RegistrationResponse, error)
	CancelRegistration(ctx context.Context, activityID, userID uuid.UUID) error
	ApproveRegistration(ctx context.Context, activityID, userID uuid.UUID, approve bool, reason string) (*dto.RegistrationResponse, error)
	ListRegistrations(ctx context.Context, activityID uuid.UUID) ([]dto.RegistrationResponse, error)

	// 签到
	Checkin(ctx context.Context, activityID, userID uuid.UUID, req *dto.CheckinRequest) (*dto.RegistrationResponse, error)

	// 统计
	GetStats(ctx context.Context, activityID uuid.UUID) (*dto.ActivityStatsResponse, error)

	// 满意度调查
	SubmitSurvey(ctx context.Context, activityID, userID uuid.UUID, req *dto.SurveyRequest) error
}
