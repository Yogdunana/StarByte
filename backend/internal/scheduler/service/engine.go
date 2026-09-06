package service

import (
	"context"
	"os"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/cache"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Engine struct {
	repo     repo.Repository
	rdb      *redis.Client
	alerter  Alerter
	tick     time.Duration
	sleeper  func(time.Duration)
	now      func() time.Time
	workerID string
	stop     chan struct{}
	done     chan struct{}
}

func NewEngine(r repo.Repository, rdb *redis.Client, alerter Alerter) *Engine {
	host, _ := os.Hostname()
	return &Engine{
		repo: r, rdb: rdb, alerter: alerter,
		tick: time.Second, sleeper: time.Sleep, now: time.Now,
		workerID: host, stop: make(chan struct{}), done: make(chan struct{}),
	}
}

func (e *Engine) Start() {
	go e.loop()
	logger.Info("scheduler engine started")
}

func (e *Engine) Stop() {
	select {
	case <-e.stop:
		return
	default:
		close(e.stop)
	}
	<-e.done
	logger.Info("scheduler engine stopped")
}

func (e *Engine) loop() {
	defer close(e.done)
	t := time.NewTicker(e.tick)
	defer t.Stop()
	e.tickOnce()
	for {
		select {
		case <-t.C:
			e.tickOnce()
		case <-e.stop:
			return
		}
	}
}

func (e *Engine) tickOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	due, err := e.repo.ListDue(ctx, e.now(), 50)
	if err != nil {
		logger.Warn("scheduler list due failed", zap.Error(err))
		return
	}
	for i := range due {
		task := due[i]
		go e.dispatch(task, false)
	}
}

func (e *Engine) Trigger(ctx context.Context, id uuid.UUID) error {
	task, err := e.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	go e.dispatch(*task, true)
	return nil
}

func (e *Engine) dispatch(task model.Task, manual bool) {
	if e.rdb == nil {
		e.runLocked(context.Background(), task, manual, nil)
		return
	}
	ttl := time.Duration(task.TimeoutSec+5) * time.Second
	if ttl < 8*time.Second {
		ttl = 8 * time.Second
	}
	lk, err := cache.Acquire(context.Background(), e.rdb, lockName(task.ID.String()), e.workerID, ttl)
	if err == cache.ErrLockBusy {
		return
	}
	if err != nil {
		logger.Warn("scheduler lock failed", zap.Error(err), zap.String("task", task.Code))
		return
	}
	lk.StartWatchdog()
	defer func() { _ = lk.Unlock(context.Background()) }()
	e.runLocked(context.Background(), task, manual, lk)
}
