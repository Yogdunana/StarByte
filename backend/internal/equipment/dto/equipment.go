package dto

import "time"

// ========== 请求 DTO ==========

// ListEquipmentRequest 物资列表查询
type ListEquipmentRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	Category string `form:"category"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1 2 3 4"`
}

// CreateEquipmentRequest 入库
type CreateEquipmentRequest struct {
	Name           string  `json:"name" binding:"required,min=1,max=100"`
	Model          string  `json:"model" binding:"omitempty,max=100"`
	Category       string  `json:"category" binding:"required,oneof=general electronic book tool other"`
	TotalQuantity  int     `json:"total_quantity" binding:"required,min=1"`
	Location       string  `json:"location" binding:"omitempty,max=200"`
	ManagerID      string  `json:"manager_id"`
	PhotoFileID    string  `json:"photo_file_id"`
	QRCode         string  `json:"qr_code" binding:"omitempty,max=500"`
	Description    string  `json:"description"`
	PurchaseDate   string  `json:"purchase_date" binding:"omitempty,datetime=2006-01-02"`
	PurchasePrice  *float64 `json:"purchase_price"`
}

// UpdateEquipmentRequest 更新物资
type UpdateEquipmentRequest struct {
	Name           string  `json:"name" binding:"omitempty,max=100"`
	Model          string  `json:"model" binding:"omitempty,max=100"`
	Category       string  `json:"category" binding:"omitempty,oneof=general electronic book tool other"`
	TotalQuantity  *int    `json:"total_quantity" binding:"omitempty,min=1"`
	Location       string  `json:"location" binding:"omitempty,max=200"`
	ManagerID     *string `json:"manager_id"`
	PhotoFileID    *string `json:"photo_file_id"`
	Status         *int    `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
	QRCode         string  `json:"qr_code" binding:"omitempty,max=500"`
	Description    string  `json:"description"`
	PurchaseDate   *string `json:"purchase_date"`
	PurchasePrice  *float64 `json:"purchase_price"`
}

// BorrowRequest 借用申请
type BorrowRequest struct {
	EquipmentID      string `json:"equipment_id" binding:"required"`
	Quantity         int    `json:"quantity" binding:"required,min=1"`
	ExpectedReturnAt string `json:"expected_return_at" binding:"required"` // RFC3339
	Remark           string `json:"remark" binding:"omitempty,max=500"`
}

// BorrowActionRequest 借用审批
type BorrowActionRequest struct {
	Status int    `json:"status" binding:"required,oneof=1 4"` // 1=批准 4=拒绝
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

// ReturnRequest 归还确认
type ReturnRequest struct {
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

// ListBorrowRequest 借用记录查询
type ListBorrowRequest struct {
	Page       int    `form:"page,default=1" binding:"min=1"`
	PageSize   int    `form:"page_size,default=20" binding:"min=1,max=100"`
	EquipmentID string `form:"equipment_id"`
	BorrowerID string `form:"borrower_id"`
	Status     *int   `form:"status"`
}

// CreateMaintenanceRequest 维修记录
type CreateMaintenanceRequest struct {
	EquipmentID string  `json:"equipment_id" binding:"required"`
	Type        int     `json:"type" binding:"required,oneof=0 1 2"`
	Description string  `json:"description" binding:"required,min=5,max=500"`
	Cost        *float64 `json:"cost"`
	StartDate   string  `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate     string  `json:"end_date" binding:"omitempty"`
	Remark      string  `json:"remark" binding:"omitempty,max=500"`
}

// CreateInventoryRequest 库存盘点
type CreateInventoryRequest struct {
	EquipmentID      string `json:"equipment_id" binding:"required"`
	ActualQuantity   int    `json:"actual_quantity" binding:"required,min=0"`
	Remark           string `json:"remark" binding:"omitempty,max=500"`
}

// ========== 响应 DTO ==========

// EquipmentResponse 物资响应
type EquipmentResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Model             string  `json:"model"`
	Category          string  `json:"category"`
	CategoryText      string  `json:"category_text"`
	TotalQuantity     int     `json:"total_quantity"`
	AvailableQuantity int     `json:"available_quantity"`
	Location          string  `json:"location"`
	ManagerID         string  `json:"manager_id"`
	ManagerName       string  `json:"manager_name"`
	PhotoFileID       string  `json:"photo_file_id"`
	PhotoURL          string  `json:"photo_url"`
	Status            int     `json:"status"`
	StatusText        string  `json:"status_text"`
	QRCode            string  `json:"qr_code"`
	Description       string  `json:"description"`
	PurchaseDate      string  `json:"purchase_date"`
	PurchasePrice     float64 `json:"purchase_price"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// BorrowResponse 借用记录响应
type BorrowResponse struct {
	ID               string `json:"id"`
	EquipmentID      string `json:"equipment_id"`
	EquipmentName    string `json:"equipment_name"`
	BorrowerID       string `json:"borrower_id"`
	BorrowerName     string `json:"borrower_name"`
	Quantity         int    `json:"quantity"`
	BorrowAt         string `json:"borrow_at"`
	ExpectedReturnAt string `json:"expected_return_at"`
	ActualReturnAt   string `json:"actual_return_at"`
	Status           int    `json:"status"`
	StatusText       string `json:"status_text"`
	ApproverID       string `json:"approver_id"`
	ApproverName     string `json:"approver_name"`
	ApprovedAt       string `json:"approved_at"`
	ApprovedRemark   string `json:"approved_remark"`
	ReturnRemark     string `json:"return_remark"`
	ReturnCheckerID  string `json:"return_checker_id"`
	Remark           string `json:"remark"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// MaintenanceResponse 维修记录响应
type MaintenanceResponse struct {
	ID           string  `json:"id"`
	EquipmentID  string  `json:"equipment_id"`
	EquipmentName string `json:"equipment_name"`
	Type         int     `json:"type"`
	TypeText     string  `json:"type_text"`
	Description string  `json:"description"`
	Cost         float64 `json:"cost"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	Status       int     `json:"status"`
	StatusText   string  `json:"status_text"`
	OperatorID   string  `json:"operator_id"`
	Remark       string  `json:"remark"`
	CreatedAt    string  `json:"created_at"`
}

// InventoryResponse 盘点记录响应
type InventoryResponse struct {
	ID               string `json:"id"`
	EquipmentID      string `json:"equipment_id"`
	EquipmentName    string `json:"equipment_name"`
	ExpectedQuantity int    `json:"expected_quantity"`
	ActualQuantity   int    `json:"actual_quantity"`
	Difference       int    `json:"difference"`
	CheckerID       string `json:"checker_id"`
	CheckDate        string `json:"check_date"`
	Remark           string `json:"remark"`
	CreatedAt        string `json:"created_at"`
}

// ========== 文本映射 ==========

var equipmentStatusText = map[int]string{
	0: "可用", 1: "部分借用", 2: "全部借出", 3: "维修中", 4: "已报废",
}

var borrowStatusText = map[int]string{
	0: "待审批", 1: "已批准", 2: "已领取", 3: "已归还", 4: "已拒绝", 5: "逾期",
}

var maintenanceTypeText = map[int]string{
	0: "维修", 1: "保养", 2: "报废",
}

var maintenanceStatusText = map[int]string{
	0: "进行中", 1: "已完成", 2: "已取消",
}

var equipmentCategoryText = map[string]string{
	"general": "通用", "electronic": "电子设备", "book": "书籍", "tool": "工具", "other": "其他",
}

func GetEquipmentStatusText(status int) string {
	if t, ok := equipmentStatusText[status]; ok {
		return t
	}
	return "未知"
}

func GetBorrowStatusText(status int) string {
	if t, ok := borrowStatusText[status]; ok {
		return t
	}
	return "未知"
}

func GetMaintenanceTypeText(t int) string {
	if s, ok := maintenanceTypeText[t]; ok {
		return s
	}
	return "未知"
}

func GetMaintenanceStatusText(s int) string {
	if t, ok := maintenanceStatusText[s]; ok {
		return t
	}
	return "未知"
}

func GetCategoryText(c string) string {
	if t, ok := equipmentCategoryText[c]; ok {
		return t
	}
	return "其他"
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatTimePtr 格式化可空时间
func FormatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDatePtr 格式化可空日期
func FormatDatePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatFloat 格式化浮点数
func FormatFloatPtr(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
