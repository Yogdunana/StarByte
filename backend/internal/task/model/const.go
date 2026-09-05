package model

const (
	StatusPending   int16 = 0 // 待处理
	StatusDoing     int16 = 1 // 进行中
	StatusDone      int16 = 2 // 已完成
	StatusCancelled int16 = 3 // 已取消
	StatusHeld      int16 = 4 // 已挂起

	PriorityLow    int16 = 0
	PriorityMedium int16 = 1
	PriorityHigh   int16 = 2
	PriorityUrgent int16 = 3

	ActionStatusChange = "status_change"
	ActionAssign       = "assign"
	ActionTransfer     = "transfer"
	ActionComment      = "comment"
	ActionUrge         = "urge"
	ActionCreate       = "create"
)

func IsClosed(status int16) bool {
	return status == StatusDone || status == StatusCancelled
}
