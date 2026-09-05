package model

const (
	SessionPending   int16 = 0
	SessionOngoing   int16 = 1
	SessionEnded     int16 = 2
	SessionCancelled int16 = 3

	InterviewPending   int16 = 0
	InterviewCheckedIn int16 = 1
	InterviewOngoing   int16 = 2
	InterviewDone      int16 = 3
	InterviewAbsent    int16 = 4
	InterviewCancelled int16 = 5

	ResultNone    int16 = 0
	ResultPass    int16 = 1
	ResultFail    int16 = 2
	ResultPending int16 = 3

	DefaultDuration = 30
)
