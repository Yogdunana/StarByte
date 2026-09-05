package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *meetingService) AddAgenda(ctx context.Context, meetingID uuid.UUID, req *dto.CreateAgendaRequest) (*dto.AgendaResponse, error) {
	if _, err := s.mustMeeting(ctx, meetingID); err != nil {
		return nil, err
	}
	now := time.Now()
	a := &model.Agenda{
		ID:          uuid.New(),
		MeetingID:   meetingID,
		Title:       req.Title,
		Description: req.Content,
		SortOrder:   req.SortOrder,
		Presenter:   req.Presenter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.Duration > 0 {
		d := req.Duration
		a.Duration = &d
	}
	if err := s.agendas.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("create agenda: %w", err)
	}
	out := mapAgenda(*a)
	return &out, nil
}

func (s *meetingService) UpdateAgenda(ctx context.Context, meetingID, agendaID uuid.UUID, req *dto.UpdateAgendaRequest) (*dto.AgendaResponse, error) {
	a, err := s.mustAgenda(ctx, meetingID, agendaID)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Content != nil {
		a.Description = *req.Content
	}
	if req.Presenter != nil {
		a.Presenter = *req.Presenter
	}
	if req.SortOrder != nil {
		a.SortOrder = *req.SortOrder
	}
	if req.Duration != nil {
		d := *req.Duration
		a.Duration = &d
	}
	a.UpdatedAt = time.Now()
	if err := s.agendas.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update agenda: %w", err)
	}
	out := mapAgenda(*a)
	return &out, nil
}

func (s *meetingService) DeleteAgenda(ctx context.Context, meetingID, agendaID uuid.UUID) error {
	if _, err := s.mustAgenda(ctx, meetingID, agendaID); err != nil {
		return err
	}
	if err := s.agendas.Delete(ctx, agendaID); err != nil {
		return fmt.Errorf("delete agenda: %w", err)
	}
	return nil
}

func (s *meetingService) SortAgendas(ctx context.Context, meetingID uuid.UUID, ids []uuid.UUID) ([]dto.AgendaResponse, error) {
	if _, err := s.mustMeeting(ctx, meetingID); err != nil {
		return nil, err
	}
	exist, err := s.agendas.ListByMeeting(ctx, meetingID)
	if err != nil {
		return nil, fmt.Errorf("list agendas: %w", err)
	}
	byID := map[uuid.UUID]model.Agenda{}
	for _, a := range exist {
		byID[a.ID] = a
	}
	items := make([]model.Agenda, 0, len(ids))
	for i, id := range ids {
		a, ok := byID[id]
		if !ok {
			return nil, response.NewError(response.CodeBadRequest, "议程不属于该会议")
		}
		a.SortOrder = i + 1
		items = append(items, a)
	}
	if err := s.agendas.SaveSort(ctx, items); err != nil {
		return nil, fmt.Errorf("sort agendas: %w", err)
	}
	return s.ListAgendas(ctx, meetingID)
}

func (s *meetingService) ListAgendas(ctx context.Context, meetingID uuid.UUID) ([]dto.AgendaResponse, error) {
	if _, err := s.mustMeeting(ctx, meetingID); err != nil {
		return nil, err
	}
	rows, err := s.agendas.ListByMeeting(ctx, meetingID)
	if err != nil {
		return nil, fmt.Errorf("list agendas: %w", err)
	}
	out := make([]dto.AgendaResponse, 0, len(rows))
	for _, a := range rows {
		out = append(out, mapAgenda(a))
	}
	return out, nil
}

func (s *meetingService) mustAgenda(ctx context.Context, meetingID, agendaID uuid.UUID) (*model.Agenda, error) {
	a, err := s.agendas.GetByID(ctx, agendaID)
	if err != nil {
		return nil, fmt.Errorf("get agenda: %w", err)
	}
	if a == nil || a.MeetingID != meetingID {
		return nil, response.NewError(response.CodeNotFound, "议程不存在")
	}
	return a, nil
}
