package dto

// MemberImportRow 是一行待录入的人。
//
// Row 由前端给出，用来在结果里指回原始行号（Excel 的行号从 1 开始且通常带表头，
// 由调用方决定怎么算），后端只负责原样回传。
type MemberImportRow struct {
	Row        int    `json:"row"`
	StudentNo  string `json:"student_no"`
	RealName   string `json:"real_name"`
	Gender     *int16 `json:"gender"`
	Grade      string `json:"grade"`
	Major      string `json:"major"`
	Phone      string `json:"contact_phone"`
	Email      string `json:"contact_email"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

// MemberImportRequest 批量录入请求。Rows 为空时后端直接返回空结果，不报错。
type MemberImportRequest struct {
	Rows []MemberImportRow `json:"rows"`
}

// MemberImportRowStatus 单行落库结果。
type MemberImportRowStatus string

const (
	// ImportRowCreated 新建了账号与档案。
	ImportRowCreated MemberImportRowStatus = "create"
	// ImportRowUpdated 已存在，且这次把某些字段改了。
	ImportRowUpdated MemberImportRowStatus = "update"
	// ImportRowSkipped 已存在且内容一致，重复导入不会二次写入。
	ImportRowSkipped MemberImportRowStatus = "skip"
	// ImportRowFailed 这一行没过校验，同批其它行照常处理。
	ImportRowFailed MemberImportRowStatus = "error"
)

// MemberImportRowResult 单行结果，预览与提交共用同一种结构。
type MemberImportRowResult struct {
	Row        int                   `json:"row"`
	StudentNo  string                `json:"student_no"`
	RealName   string                `json:"real_name"`
	Department string                `json:"department"`
	Role       string                `json:"role"`
	Status     MemberImportRowStatus `json:"status"`
	Message    string                `json:"message"`
}

// MemberImportResponse 批量录入结果汇总。
type MemberImportResponse struct {
	Total   int                     `json:"total"`
	Created int                     `json:"created"`
	Updated int                     `json:"updated"`
	Skipped int                     `json:"skipped"`
	Failed  int                     `json:"failed"`
	Results []MemberImportRowResult `json:"results"`
}
