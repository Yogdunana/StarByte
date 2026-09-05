package model

const (
	StatusOngoing   int16 = 0 // 进行中
	StatusDone      int16 = 1 // 已完成
	StatusAborted   int16 = 2 // 已中止
	TypeCampusClub  int16 = 0 // 校内社团
	TypeCampusOther int16 = 1 // 校内其他
	TypeOffCampus   int16 = 2 // 校外实习
	ConfigKey             = "internship_config"
)

func IsClosed(status int16) bool {
	return status == StatusDone || status == StatusAborted
}
