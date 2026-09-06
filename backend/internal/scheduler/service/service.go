package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchedulerService interface {
	List(ctx context.Context, q dto.ListQuery) ([]dto.TaskResponse, int64, int, int, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error)
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateTaskRequest) (*dto.TaskResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Pause(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error)
	Resume(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error)
	RunNow(ctx context.Context, id uuid.UUID) error
	Logs(ctx context.Context, id uuid.UUID, runID *uuid.UUID) (*dto.LogsResponse, error)
	Handlers() []dto.HandlerInfo
}

type schedulerService struct {
	repo   repo.Repository
	engine *Engine
}

func NewService(r repo.Repository, eng *Engine) SchedulerService {
	return &schedulerService{repo: r, engine: eng}
}

func (s *schedulerService) Handlers() []dto.HandlerInfo {
	src := publicHandlers()
	out := make([]dto.HandlerInfo, 0, len(src))
	for _, h := range src {
		out = append(out, dto.HandlerInfo{Key: h.Key, Description: h.Description})
	}
	return out
}

func (s *schedulerService) List(ctx context.Context, q dto.ListQuery) ([]dto.TaskResponse, int64, int, int, error) {
	page, size := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	list, total, err := s.repo.ListTasks(ctx, strings.TrimSpace(q.Keyword), q.Status, (page-1)*size, size)
	if err != nil {
		return nil, 0, page, size, err
	}
	out := make([]dto.TaskResponse, 0, len(list))
	for i := range list {
		out = append(out, toTaskDTO(&list[i]))
	}
	return out, total, page, size, nil
}

func (s *schedulerService) Get(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error) {
	rec, err := s.getActive(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toTaskDTO(rec)
	return &out, nil
}

func (s *schedulerService) getActive(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	rec, err := s.repo.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewError(response.CodeSchedulerNotFound, "定时任务不存在")
		}
		return nil, err
	}
	if rec.Status == model.StatusDeleted {
		return nil, response.NewError(response.CodeSchedulerNotFound, "定时任务不存在")
	}
	return rec, nil
}

func defaultInt(p *int, fallback, min, max int) int {
	v := fallback
	if p != nil {
		v = *p
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (s *schedulerService) computeNext(rec *model.Task) error {
	if rec.Status != model.StatusActive {
		rec.NextRunAt = nil
		return nil
	}
	if rec.IsCron() {
		n, err := nextRun(rec.CronExpr, rec.Timezone, time.Now())
		if err != nil {
			return err
		}
		rec.NextRunAt = n
		return nil
	}
	rec.NextRunAt = rec.RunAt
	return nil
}
