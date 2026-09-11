package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/activity/repo"
	"gorm.io/gorm"
)

func (s *activityService) withTx(ctx context.Context, fn func(*activityService) error) error {
	if s.db == nil {
		return fn(s)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(s.bind(tx))
	})
}

func (s *activityService) bind(tx *gorm.DB) *activityService {
	cp := *s
	cp.db = tx
	cp.activities = repo.NewActivityRepo(tx)
	cp.regs = repo.NewRegistrationRepo(tx)
	cp.surveys = repo.NewSurveyRepo(tx)
	return &cp
}
