package dto

import "time"

type VoteOptionInput struct {
	Key   string `json:"key" binding:"required,max=64"`
	Label string `json:"label" binding:"required,max=200"`
}

type CreateVoteRequest struct {
	Title       string            `json:"title" binding:"required,max=200"`
	Description string            `json:"description"`
	VoteType    int16             `json:"vote_type" binding:"required,oneof=1 2"`
	IsAnonymous bool              `json:"is_anonymous"`
	Options     []VoteOptionInput `json:"options" binding:"required,min=2,dive"`
	Duration    int               `json:"duration"`
}

type CastVoteRequest struct {
	OptionKey string `json:"option_key" binding:"required"`
}

type VoteOptionResponse struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type VoteResponse struct {
	ID          string               `json:"id"`
	MeetingID   string               `json:"meeting_id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	VoteType    int16                `json:"vote_type"`
	IsAnonymous bool                 `json:"is_anonymous"`
	Options     []VoteOptionResponse `json:"options"`
	Status      int16                `json:"status"`
	StartTime   *time.Time           `json:"start_time,omitempty"`
	EndTime     *time.Time           `json:"end_time,omitempty"`
	HasVoted    bool                 `json:"has_voted"`
	CreatedAt   time.Time            `json:"created_at"`
}

type VoteResultItem struct {
	OptionKey   string  `json:"option_key"`
	OptionLabel string  `json:"option_label"`
	Count       int     `json:"count"`
	WeightTotal float64 `json:"weight_total"`
}

type VoteResultResponse struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	VoteType    int16            `json:"vote_type"`
	IsAnonymous bool             `json:"is_anonymous"`
	Status      int16            `json:"status"`
	Results     []VoteResultItem `json:"results"`
	TotalVoters int              `json:"total_voters"`
	TotalWeight float64          `json:"total_weight"`
	StartTime   *time.Time       `json:"start_time,omitempty"`
	EndTime     *time.Time       `json:"end_time,omitempty"`
}

type MyVoteResponse struct {
	VoteID    string    `json:"vote_id"`
	OptionKey string    `json:"option_key"`
	Weight    float64   `json:"weight"`
	VotedAt   time.Time `json:"voted_at"`
}

type VoteWeightConfigRequest struct {
	Weights       map[string]float64 `json:"weights" binding:"required"`
	DefaultWeight float64            `json:"default_weight" binding:"required,gt=0"`
}

type VoteWeightConfigResponse struct {
	Weights       map[string]float64 `json:"weights"`
	DefaultWeight float64            `json:"default_weight"`
}
