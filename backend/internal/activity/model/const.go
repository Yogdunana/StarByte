package model

// 活动状态
const (
	ActivityDraft     int16 = 0 // 草稿
	ActivityOpen      int16 = 1 // 报名中
	ActivityOngoing   int16 = 2 // 进行中
	ActivityEnded     int16 = 3 // 已结束
	ActivityCancelled int16 = 4 // 已取消
)

// 报名状态
const (
	RegPending   int16 = 0 // 待审批
	RegApproved  int16 = 1 // 已通过
	RegRejected  int16 = 2 // 已拒绝
	RegWaitlist  int16 = 3 // 候补
	RegCancelled int16 = 4 // 已取消
)

// 签到状态
const (
	CheckinPending int16 = 0 // 未签到
	CheckinDone    int16 = 1 // 已签到
)

// 签到方式
const (
	CheckinMethodQR  int16 = 1 // 二维码
	CheckinMethodGPS int16 = 2 // GPS
)

const DefaultCheckinTokenTTLSeconds = 600
