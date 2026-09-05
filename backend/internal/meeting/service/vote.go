package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *meetingService) CreateVote(ctx context.Context, meetingID uuid.UUID, req *dto.CreateVoteRequest) (*dto.VoteResponse, error) {
	m, err := s.mustMeeting(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	if !canCreateVote(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "当前会议不可发起投票")
	}
	if err := validateOptions(req.Options); err != nil {
		return nil, err
	}
	now := time.Now()
	end := (*time.Time)(nil)
	if req.Duration > 0 {
		t := now.Add(time.Duration(req.Duration) * time.Second)
		end = &t
	}
	v := &model.Vote{
		ID: uuid.New(), MeetingID: meetingID, Title: req.Title, Description: req.Description,
		VoteType: req.VoteType, Status: model.VoteOpen, IsAnonymous: req.IsAnonymous,
		StartTime: &now, EndTime: end, CreatedAt: now, UpdatedAt: now,
	}
	opts := make([]model.VoteOption, 0, len(req.Options))
	for i, o := range req.Options {
		opts = append(opts, model.VoteOption{
			ID: uuid.New(), VoteID: v.ID, OptionText: o.Label, OptionKey: o.Key, SortOrder: i + 1,
		})
	}
	if err := s.votes.CreateVote(ctx, v, opts); err != nil {
		return nil, fmt.Errorf("create vote: %w", err)
	}
	return mapVote(v, opts, false), nil
}

func (s *meetingService) ListVotes(ctx context.Context, meetingID, viewer uuid.UUID) ([]*dto.VoteResponse, error) {
	if _, err := s.mustMeeting(ctx, meetingID); err != nil {
		return nil, err
	}
	rows, err := s.votes.ListByMeeting(ctx, meetingID)
	if err != nil {
		return nil, fmt.Errorf("list votes: %w", err)
	}
	out := make([]*dto.VoteResponse, 0, len(rows))
	for i := range rows {
		item, err := s.voteDTO(ctx, &rows[i], viewer)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *meetingService) GetVote(ctx context.Context, id, viewer uuid.UUID) (*dto.VoteResponse, error) {
	v, err := s.mustVote(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.voteDTO(ctx, v, viewer)
}

func (s *meetingService) CastVote(ctx context.Context, voteID, userID uuid.UUID, optionKey string) error {
	v, err := s.ensureVoteOpen(ctx, voteID)
	if err != nil {
		return err
	}
	ok, err := s.attendees.IsAttendee(ctx, v.MeetingID, userID)
	if err != nil {
		return fmt.Errorf("check attendee: %w", err)
	}
	if !ok {
		return response.NewError(response.CodeVoteNoAccess, "无权投票（非参会人）")
	}
	exist, err := s.votes.GetRecord(ctx, voteID, userID)
	if err != nil {
		return fmt.Errorf("get record: %w", err)
	}
	if exist != nil {
		return response.NewError(response.CodeVoteDuplicate, "重复投票")
	}
	opt, err := s.votes.GetOptionByKey(ctx, voteID, optionKey)
	if err != nil {
		return fmt.Errorf("get option: %w", err)
	}
	if opt == nil {
		return response.NewError(response.CodeVoteOptionGone, "投票选项不存在")
	}
	weight := 1.0
	if v.VoteType == model.VoteWeighted {
		cfg, err := s.loadWeight(ctx)
		if err != nil {
			return err
		}
		u, _ := s.meetings.GetUser(ctx, userID)
		code := ""
		if u != nil {
			code = u.PositionCode
		}
		weight = ResolveWeight(cfg, code, v.VoteType)
	}
	rec := &model.VoteRecord{
		ID: uuid.New(), VoteID: voteID, VoterID: userID, OptionID: opt.ID,
		OptionKey: opt.OptionKey, Weight: weight, VotedAt: time.Now(),
	}
	if err := s.votes.CreateRecord(ctx, rec); err != nil {
		return fmt.Errorf("cast vote: %w", err)
	}
	return nil
}

func (s *meetingService) VoteResult(ctx context.Context, id uuid.UUID) (*dto.VoteResultResponse, error) {
	v, err := s.ensureExpiredClosed(ctx, id)
	if err != nil {
		return nil, err
	}
	opts, err := s.votes.ListOptions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	recs, err := s.votes.ListRecords(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}
	plain := make([]model.VoteRecord, 0, len(recs))
	for _, r := range recs {
		plain = append(plain, r.VoteRecord)
	}
	agg, total, totalW := CalculateVoteResult(plain)
	items := make([]dto.VoteResultItem, 0, len(opts))
	for _, o := range opts {
		a := agg[o.OptionKey]
		items = append(items, dto.VoteResultItem{
			OptionKey: o.OptionKey, OptionLabel: o.OptionText, Count: a.Count, WeightTotal: a.Weight,
		})
	}
	return &dto.VoteResultResponse{
		ID: v.ID.String(), Title: v.Title, VoteType: v.VoteType, IsAnonymous: v.IsAnonymous,
		Status: v.Status, Results: items, TotalVoters: total, TotalWeight: totalW,
		StartTime: v.StartTime, EndTime: v.EndTime,
	}, nil
}

func (s *meetingService) CloseVote(ctx context.Context, id uuid.UUID) (*dto.VoteResponse, error) {
	v, err := s.mustVote(ctx, id)
	if err != nil {
		return nil, err
	}
	if v.Status != model.VoteOpen && v.Status != model.VotePending {
		return nil, response.NewError(response.CodeVoteNotOpen, "投票未开始或已结束")
	}
	v.Status = model.VoteClosed
	now := time.Now()
	v.EndTime = &now
	v.UpdatedAt = now
	if err := s.votes.UpdateVote(ctx, v); err != nil {
		return nil, fmt.Errorf("close vote: %w", err)
	}
	return s.voteDTO(ctx, v, uuid.Nil)
}

func (s *meetingService) MyVote(ctx context.Context, voteID, userID uuid.UUID) (*dto.MyVoteResponse, error) {
	v, err := s.mustVote(ctx, voteID)
	if err != nil {
		return nil, err
	}
	if v.IsAnonymous {
		return nil, response.NewError(response.CodeVoteAnonymousHidden, "匿名投票无法查看个人记录")
	}
	rec, err := s.votes.GetRecord(ctx, voteID, userID)
	if err != nil {
		return nil, fmt.Errorf("get record: %w", err)
	}
	if rec == nil {
		return nil, response.NewError(response.CodeNotFound, "尚未投票")
	}
	return &dto.MyVoteResponse{
		VoteID: voteID.String(), OptionKey: rec.OptionKey, Weight: rec.Weight, VotedAt: rec.VotedAt,
	}, nil
}

func (s *meetingService) GetWeightConfig(ctx context.Context) (*dto.VoteWeightConfigResponse, error) {
	cfg, err := s.loadWeight(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.VoteWeightConfigResponse{Weights: cfg.Weights, DefaultWeight: cfg.DefaultWeight}, nil
}

func (s *meetingService) UpdateWeightConfig(ctx context.Context, operator uuid.UUID, req *dto.VoteWeightConfigRequest) (*dto.VoteWeightConfigResponse, error) {
	raw, err := json.Marshal(model.VoteWeightConfig{Weights: req.Weights, DefaultWeight: req.DefaultWeight})
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "权重配置无效")
	}
	exist, err := s.votes.GetConfig(ctx, model.WeightConfigKey)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	now := time.Now()
	if exist == nil {
		exist = &model.SystemConfig{
			ID: uuid.New(), ConfigKey: model.WeightConfigKey, ConfigType: "json",
			Description: "会议加权投票职务权重", Category: "meeting",
		}
	}
	exist.ConfigValue = string(raw)
	exist.UpdatedBy = &operator
	exist.UpdatedAt = now
	if err := s.votes.UpsertConfig(ctx, exist); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}
	return &dto.VoteWeightConfigResponse{Weights: req.Weights, DefaultWeight: req.DefaultWeight}, nil
}
