package handler

import (
	"net/http"
	"strconv"

	"github.com/Yogdunana/StarByte/backend/internal/equipment/dto"
	"github.com/Yogdunana/StarByte/backend/internal/equipment/service"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EquipmentHandler 物资处理器
type EquipmentHandler struct {
	svc service.EquipmentService
}

func NewEquipmentHandler(svc service.EquipmentService) *EquipmentHandler {
	return &EquipmentHandler{svc: svc}
}

// ListEquipment 物资列表
func (h *EquipmentHandler) ListEquipment(c *gin.Context) {
	var req dto.ListEquipmentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListEquipment(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// CreateEquipment 入库
func (h *EquipmentHandler) CreateEquipment(c *gin.Context) {
	var req dto.CreateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	creatorID := authmiddleware.GetUserID(c)
	result, err := h.svc.CreateEquipment(c.Request.Context(), &req, creatorID)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{
		Code:      response.CodeSuccess,
		Message:   "success",
		Data:      result,
		RequestID: c.GetString("request_id"),
	})
}

// UpdateEquipment 更新物资
func (h *EquipmentHandler) UpdateEquipment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的物资ID")
		return
	}
	var req dto.UpdateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateEquipment(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// DeleteEquipment 删除物资
func (h *EquipmentHandler) DeleteEquipment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的物资ID")
		return
	}
	err = h.svc.DeleteEquipment(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

// CreateBorrow 借用申请
func (h *EquipmentHandler) CreateBorrow(c *gin.Context) {
	var req dto.BorrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	borrowerID := authmiddleware.GetUserID(c)
	result, err := h.svc.CreateBorrow(c.Request.Context(), &req, borrowerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{
		Code:      response.CodeSuccess,
		Message:   "success",
		Data:      result,
		RequestID: c.GetString("request_id"),
	})
}

// ActionBorrow 借用审批
func (h *EquipmentHandler) ActionBorrow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的借用记录ID")
		return
	}
	var req dto.BorrowActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	approverID := authmiddleware.GetUserID(c)
	result, err := h.svc.ActionBorrow(c.Request.Context(), id, &req, approverID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ReturnBorrow 归还确认
func (h *EquipmentHandler) ReturnBorrow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的借用记录ID")
		return
	}
	var req dto.ReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	checkerID := authmiddleware.GetUserID(c)
	result, err := h.svc.ReturnBorrow(c.Request.Context(), id, &req, checkerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ListBorrows 借用记录列表
func (h *EquipmentHandler) ListBorrows(c *gin.Context) {
	var req dto.ListBorrowRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListBorrows(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// CreateMaintenance 维修记录
func (h *EquipmentHandler) CreateMaintenance(c *gin.Context) {
	var req dto.CreateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	operatorID := authmiddleware.GetUserID(c)
	result, err := h.svc.CreateMaintenance(c.Request.Context(), &req, operatorID)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{
		Code:      response.CodeSuccess,
		Message:   "success",
		Data:      result,
		RequestID: c.GetString("request_id"),
	})
}

// ListMaintenance 维修记录列表
func (h *EquipmentHandler) ListMaintenance(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	equipmentID := c.Query("equipment_id")
	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}
	list, total, err := h.svc.ListMaintenance(c.Request.Context(), page, pageSize, equipmentID, status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, pageSize)
}

// CreateInventory 库存盘点
func (h *EquipmentHandler) CreateInventory(c *gin.Context) {
	var req dto.CreateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	checkerID := authmiddleware.GetUserID(c)
	result, err := h.svc.CreateInventory(c.Request.Context(), &req, checkerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{
		Code:      response.CodeSuccess,
		Message:   "success",
		Data:      result,
		RequestID: c.GetString("request_id"),
	})
}

// ListInventory 盘点记录列表
func (h *EquipmentHandler) ListInventory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	equipmentID := c.Query("equipment_id")
	list, total, err := h.svc.ListInventory(c.Request.Context(), page, pageSize, equipmentID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, pageSize)
}
