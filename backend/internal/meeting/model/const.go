package model

const (
	MeetingPending   int16 = 0
	MeetingOngoing   int16 = 1
	MeetingEnded     int16 = 2
	MeetingCancelled int16 = 3

	MeetingTypeRegular int16 = 1
	MeetingTypeAdhoc   int16 = 2
	MeetingTypeOnline  int16 = 3

	VoteEqual    int16 = 1
	VoteWeighted int16 = 2

	VotePending   int16 = 0
	VoteOpen      int16 = 1
	VoteClosed    int16 = 2
	VoteCancelled int16 = 3

	WeightConfigKey = "vote_weight_config"
)
