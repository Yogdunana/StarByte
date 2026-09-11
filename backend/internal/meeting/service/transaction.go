package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *meetingService) bind(tx *gorm.DB) *meetingService {
	bound := *s
	bound.db = nil
	bound.access = repo.NewAccessRepo(tx)
	bound.electorate = repo.NewElectorateRepo(tx)
	bound.meetings = repo.NewMeetingRepo(tx)
	bound.attendees = repo.NewAttendeeRepo(tx)
	bound.votes = repo.NewVoteRepo(tx)
	bound.agendas = repo.NewAgendaRepo(tx)
	return &bound
}

// meetingTransaction serializes meeting changes, attendance and votes. Effects
// are dispatched only after commit; rollback cannot send an invitation.
func (s *meetingService) meetingTransaction(ctx context.Context, id uuid.UUID, fn func(*meetingService) error) error {
	if s.db == nil {
		return fn(s)
	}
	pending := []func(){}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if id != uuid.Nil {
			if err := repo.LockMeeting(ctx, tx, id); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return response.NewError(response.CodeMeetingNotFound, "会议不存在")
				}
				return fmt.Errorf("lock meeting: %w", err)
			}
		}
		bound := s.bind(tx)
		bound.afterCommit = &pending
		return fn(bound)
	})
	if err == nil {
		for _, effect := range pending {
			effect()
		}
	}
	return err
}

func meetingMutation[T any](ctx context.Context, s *meetingService, id uuid.UUID, fn func(*meetingService) (T, error)) (T, error) {
	var result T
	err := s.meetingTransaction(ctx, id, func(bound *meetingService) error {
		var err error
		result, err = fn(bound)
		return err
	})
	return result, err
}
