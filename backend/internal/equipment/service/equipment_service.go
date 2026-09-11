package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/equipment/dto"
	"github.com/Yogdunana/StarByte/backend/internal/equipment/model"
	"github.com/Yogdunana/StarByte/backend/internal/equipment/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EquipmentNotifier 通知接口
type EquipmentNotifier interface {
	NotifyBorrowApproved(ctx context.Context, userID uuid.UUID, equipmentName string) error
	NotifyBorrowRejected(ctx context.Context, userID uuid.UUID, equipmentName, remark string) error
	NotifyOverdue(ctx context.Context, userID uuid.UUID, equipmentName string, expectedReturn time.Time) error
	NotifyMaintenanceStart(ctx context.Context, userID uuid.UUID, equipmentName string) error
}

// EquipmentService 物资服务接口
type EquipmentService interface {
	ListEquipment(ctx context.Context, req *dto.ListEquipmentRequest) ([]dto.EquipmentResponse, int64, error)
	CreateEquipment(ctx context.Context, req *dto.CreateEquipmentRequest, creatorID string) (*dto.EquipmentResponse, error)
	UpdateEquipment(ctx context.Context, id uuid.UUID, req *dto.UpdateEquipmentRequest) (*dto.EquipmentResponse, error)
	DeleteEquipment(ctx context.Context, id uuid.UUID) error

	CreateBorrow(ctx context.Context, req *dto.BorrowRequest, borrowerID string) (*dto.BorrowResponse, error)
	ActionBorrow(ctx context.Context, id uuid.UUID, req *dto.BorrowActionRequest, approverID string) (*dto.BorrowResponse, error)
	ReturnBorrow(ctx context.Context, id uuid.UUID, req *dto.ReturnRequest, checkerID string) (*dto.BorrowResponse, error)
	ListBorrows(ctx context.Context, req *dto.ListBorrowRequest) ([]dto.BorrowResponse, int64, error)

	CreateMaintenance(ctx context.Context, req *dto.CreateMaintenanceRequest, operatorID string) (*dto.MaintenanceResponse, error)
	ListMaintenance(ctx context.Context, page, pageSize int, equipmentID string, status *int) ([]dto.MaintenanceResponse, int64, error)

	CreateInventory(ctx context.Context, req *dto.CreateInventoryRequest, checkerID string) (*dto.InventoryResponse, error)
	ListInventory(ctx context.Context, page, pageSize int, equipmentID string) ([]dto.InventoryResponse, int64, error)
}

type equipmentService struct {
	db           *gorm.DB
	equipRepo    repo.EquipmentRepo
	borrowRepo   repo.BorrowRepo
	maintRepo    repo.MaintenanceRepo
	inventoryRepo repo.InventoryRepo
	notifier     EquipmentNotifier
}

func NewEquipmentService(
	db *gorm.DB,
	equipRepo repo.EquipmentRepo,
	borrowRepo repo.BorrowRepo,
	maintRepo repo.MaintenanceRepo,
	inventoryRepo repo.InventoryRepo,
	notifier EquipmentNotifier,
) EquipmentService {
	return &equipmentService{
		db:            db,
		equipRepo:     equipRepo,
		borrowRepo:    borrowRepo,
		maintRepo:     maintRepo,
		inventoryRepo: inventoryRepo,
		notifier:      notifier,
	}
}

// ========== 物资管理 ==========

func (s *equipmentService) ListEquipment(ctx context.Context, req *dto.ListEquipmentRequest) ([]dto.EquipmentResponse, int64, error) {
	items, total, err := s.equipRepo.List(ctx, req.Page, req.PageSize, req.Keyword, req.Category, req.Status)
	if err != nil {
		return nil, 0, fmt.Errorf("list equipment: %w", err)
	}
	result := make([]dto.EquipmentResponse, 0, len(items))
	for _, e := range items {
		result = append(result, *s.equipmentToResponse(&e))
	}
	return result, total, nil
}

func (s *equipmentService) CreateEquipment(ctx context.Context, req *dto.CreateEquipmentRequest, creatorID string) (*dto.EquipmentResponse, error) {
	creatorUUID, _ := uuid.Parse(creatorID)

	var managerID *uuid.UUID
	if req.ManagerID != "" {
		parsed, err := parseOptionalUUID(req.ManagerID, "负责人ID")
		if err != nil {
			return nil, err
		}
		managerID = parsed
	}

	var photoFileID *uuid.UUID
	if req.PhotoFileID != "" {
		parsed, err := parseOptionalUUID(req.PhotoFileID, "照片文件ID")
		if err != nil {
			return nil, err
		}
		photoFileID = parsed
	}

	var purchaseDate *time.Time
	if req.PurchaseDate != "" {
		t, err := time.Parse("2006-01-02", req.PurchaseDate)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "purchase_date 格式无效")
		}
		purchaseDate = &t
	}

	equipment := &model.Equipment{
		ID:                uuid.New(),
		Name:              req.Name,
		Model:             req.Model,
		Category:          req.Category,
		TotalQuantity:     req.TotalQuantity,
		AvailableQuantity: req.TotalQuantity,
		Location:          req.Location,
		ManagerID:         managerID,
		PhotoFileID:       photoFileID,
		QRCode:            req.QRCode,
		Description:       req.Description,
		PurchaseDate:      purchaseDate,
		PurchasePrice:     req.PurchasePrice,
		Status:            model.EquipmentStatusAvailable,
		CreatedBy:         &creatorUUID,
	}

	err := s.equipRepo.Create(ctx, nil, equipment)
	if err != nil {
		return nil, fmt.Errorf("create equipment: %w", err)
	}

	return s.equipmentToResponse(equipment), nil
}

func (s *equipmentService) UpdateEquipment(ctx context.Context, id uuid.UUID, req *dto.UpdateEquipmentRequest) (*dto.EquipmentResponse, error) {
	equipment, err := s.equipRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get equipment: %w", err)
	}
	if equipment == nil {
		return nil, response.NewError(27001, "物资不存在")
	}

	if req.Name != "" {
		equipment.Name = req.Name
	}
	if req.Model != "" {
		equipment.Model = req.Model
	}
	if req.Category != "" {
		equipment.Category = req.Category
	}
	if req.TotalQuantity != nil {
		diff := *req.TotalQuantity - equipment.TotalQuantity
		equipment.TotalQuantity = *req.TotalQuantity
		equipment.AvailableQuantity += diff
		if equipment.AvailableQuantity < 0 {
			equipment.AvailableQuantity = 0
		}
	}
	if req.Location != "" {
		equipment.Location = req.Location
	}
	if req.ManagerID != nil {
		if *req.ManagerID == "" {
			equipment.ManagerID = nil
		} else {
			parsed, err := parseOptionalUUID(*req.ManagerID, "负责人ID")
			if err != nil {
				return nil, err
			}
			equipment.ManagerID = parsed
		}
	}
	if req.PhotoFileID != nil {
		if *req.PhotoFileID == "" {
			equipment.PhotoFileID = nil
		} else {
			parsed, err := parseOptionalUUID(*req.PhotoFileID, "照片文件ID")
			if err != nil {
				return nil, err
			}
			equipment.PhotoFileID = parsed
		}
	}
	if req.Status != nil {
		equipment.Status = *req.Status
	}
	if req.QRCode != "" {
		equipment.QRCode = req.QRCode
	}
	if req.Description != "" {
		equipment.Description = req.Description
	}
	if req.PurchaseDate != nil {
		t, err := time.Parse("2006-01-02", *req.PurchaseDate)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "purchase_date 格式无效")
		}
		equipment.PurchaseDate = &t
	}
	if req.PurchasePrice != nil {
		equipment.PurchasePrice = req.PurchasePrice
	}

	err = s.equipRepo.Update(ctx, nil, equipment)
	if err != nil {
		return nil, fmt.Errorf("update equipment: %w", err)
	}
	return s.equipmentToResponse(equipment), nil
}

func (s *equipmentService) DeleteEquipment(ctx context.Context, id uuid.UUID) error {
	equipment, err := s.equipRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get equipment: %w", err)
	}
	if equipment == nil {
		return response.NewError(27001, "物资不存在")
	}
	if equipment.AvailableQuantity < equipment.TotalQuantity {
		return response.NewError(27002, "物资有未归还的借用记录，不可删除")
	}
	return s.equipRepo.Delete(ctx, id)
}

// ========== 借用管理 ==========

func (s *equipmentService) CreateBorrow(ctx context.Context, req *dto.BorrowRequest, borrowerID string) (*dto.BorrowResponse, error) {
	bID, err := parseRequiredUUID(borrowerID, "借用人ID")
	if err != nil {
		return nil, err
	}

	equipID, err := parseRequiredUUID(req.EquipmentID, "物资ID")
	if err != nil {
		return nil, err
	}

	equipment, err := s.equipRepo.GetByID(ctx, equipID)
	if err != nil {
		return nil, fmt.Errorf("get equipment: %w", err)
	}
	if equipment == nil {
		return nil, response.NewError(27001, "物资不存在")
	}
	if equipment.AvailableQuantity < req.Quantity {
		return nil, response.NewError(27003, "可用数量不足")
	}

	expectedReturn, err := time.Parse(time.RFC3339, req.ExpectedReturnAt)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "expected_return_at 格式无效，应为 RFC3339")
	}
	if expectedReturn.Before(time.Now()) {
		return nil, response.NewError(response.CodeBadRequest, "预计归还时间不能早于当前时间")
	}

	borrow := &model.Borrow{
		ID:               uuid.New(),
		EquipmentID:      equipID,
		BorrowerID:       bID,
		Quantity:         req.Quantity,
		BorrowAt:         time.Now(),
		ExpectedReturnAt: expectedReturn,
		Status:           model.BorrowStatusPending,
		Remark:           req.Remark,
	}

	err = s.borrowRepo.Create(ctx, nil, borrow)
	if err != nil {
		return nil, fmt.Errorf("create borrow: %w", err)
	}

	return s.borrowToResponse(borrow, equipment.Name), nil
}

func (s *equipmentService) ActionBorrow(ctx context.Context, id uuid.UUID, req *dto.BorrowActionRequest, approverID string) (*dto.BorrowResponse, error) {
	borrow, err := s.borrowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get borrow: %w", err)
	}
	if borrow == nil {
		return nil, response.NewError(27004, "借用记录不存在")
	}
	if borrow.Status != model.BorrowStatusPending {
		return nil, response.NewError(27005, "借用申请已处理，不可重复操作")
	}

	approverUUID, _ := uuid.Parse(approverID)
	now := time.Now()
	borrow.Status = req.Status
	borrow.ApproverID = &approverUUID
	borrow.ApprovedAt = &now
	borrow.ApprovedRemark = req.Remark

	equipment, err := s.equipRepo.GetByID(ctx, borrow.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("get equipment: %w", err)
	}
	if equipment == nil {
		return nil, response.NewError(27001, "物资不存在")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.borrowRepo.Update(ctx, tx, borrow); err != nil {
			return err
		}
		if req.Status == model.BorrowStatusApproved {
			if equipment.AvailableQuantity < borrow.Quantity {
				return response.NewError(27003, "可用数量不足")
			}
			if err := s.equipRepo.UpdateAvailableQuantity(ctx, tx, borrow.EquipmentID, -borrow.Quantity); err != nil {
				return err
			}
			equipment.AvailableQuantity -= borrow.Quantity
			if equipment.AvailableQuantity == 0 {
				equipment.Status = model.EquipmentStatusAllBorrowed
			} else {
				equipment.Status = model.EquipmentStatusPartBorrowed
			}
			if err := s.equipRepo.Update(ctx, tx, equipment); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("action borrow: %w", err)
	}

	if s.notifier != nil {
		go func() {
			if req.Status == model.BorrowStatusApproved {
				_ = s.notifier.NotifyBorrowApproved(context.Background(), borrow.BorrowerID, equipment.Name)
			} else {
				_ = s.notifier.NotifyBorrowRejected(context.Background(), borrow.BorrowerID, equipment.Name, req.Remark)
			}
		}()
	}

	return s.borrowToResponse(borrow, equipment.Name), nil
}

func (s *equipmentService) ReturnBorrow(ctx context.Context, id uuid.UUID, req *dto.ReturnRequest, checkerID string) (*dto.BorrowResponse, error) {
	borrow, err := s.borrowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get borrow: %w", err)
	}
	if borrow == nil {
		return nil, response.NewError(27004, "借用记录不存在")
	}
	if borrow.Status != model.BorrowStatusTaken && borrow.Status != model.BorrowStatusOverdue {
		return nil, response.NewError(27006, "借用记录状态不允许归还")
	}

	checkerUUID, _ := uuid.Parse(checkerID)
	now := time.Now()
	borrow.Status = model.BorrowStatusReturned
	borrow.ActualReturnAt = &now
	borrow.ReturnRemark = req.Remark
	borrow.ReturnCheckerID = &checkerUUID

	equipment, err := s.equipRepo.GetByID(ctx, borrow.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("get equipment: %w", err)
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.borrowRepo.Update(ctx, tx, borrow); err != nil {
			return err
		}
		if err := s.equipRepo.UpdateAvailableQuantity(ctx, tx, borrow.EquipmentID, borrow.Quantity); err != nil {
			return err
		}
		if equipment != nil {
			equipment.AvailableQuantity += borrow.Quantity
			if equipment.AvailableQuantity >= equipment.TotalQuantity {
				equipment.Status = model.EquipmentStatusAvailable
			} else if equipment.Status == model.EquipmentStatusAllBorrowed {
				equipment.Status = model.EquipmentStatusPartBorrowed
			}
			if err := s.equipRepo.Update(ctx, tx, equipment); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("return borrow: %w", err)
	}

	equipName := ""
	if equipment != nil {
		equipName = equipment.Name
	}
	return s.borrowToResponse(borrow, equipName), nil
}

func (s *equipmentService) ListBorrows(ctx context.Context, req *dto.ListBorrowRequest) ([]dto.BorrowResponse, int64, error) {
	var equipID uuid.UUID
	if req.EquipmentID != "" {
		parsed, err := parseRequiredUUID(req.EquipmentID, "物资ID")
		if err != nil {
			return nil, 0, err
		}
		equipID = parsed
	}

	var borrowerID uuid.UUID
	if req.BorrowerID != "" {
		parsed, err := parseRequiredUUID(req.BorrowerID, "借用人ID")
		if err != nil {
			return nil, 0, err
		}
		borrowerID = parsed
	}

	borrows, total, err := s.borrowRepo.List(ctx, req.Page, req.PageSize, equipID, borrowerID, req.Status)
	if err != nil {
		return nil, 0, fmt.Errorf("list borrows: %w", err)
	}

	result := make([]dto.BorrowResponse, 0, len(borrows))
	for _, b := range borrows {
		result = append(result, *s.borrowToResponse(&b, ""))
	}
	return result, total, nil
}

// ========== 维修管理 ==========

func (s *equipmentService) CreateMaintenance(ctx context.Context, req *dto.CreateMaintenanceRequest, operatorID string) (*dto.MaintenanceResponse, error) {
	opUUID, _ := uuid.Parse(operatorID)
	equipID, err := parseRequiredUUID(req.EquipmentID, "物资ID")
	if err != nil {
		return nil, err
	}

	equipment, err := s.equipRepo.GetByID(ctx, equipID)
	if err != nil {
		return nil, fmt.Errorf("get equipment: %w", err)
	}
	if equipment == nil {
		return nil, response.NewError(27001, "物资不存在")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "start_date 格式无效")
	}

	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "end_date 格式无效")
		}
		endDate = &t
	}

	maintenance := &model.Maintenance{
		ID:          uuid.New(),
		EquipmentID: equipID,
		Type:        req.Type,
		Description: req.Description,
		Cost:        req.Cost,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      model.MaintenanceStatusInProgress,
		OperatorID:  &opUUID,
		Remark:      req.Remark,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.maintRepo.Create(ctx, tx, maintenance); err != nil {
			return err
		}
		if req.Type == model.MaintenanceTypeScrap {
			equipment.Status = model.EquipmentStatusScrapped
		} else {
			equipment.Status = model.EquipmentStatusRepairing
		}
		return s.equipRepo.Update(ctx, tx, equipment)
	})

	if err != nil {
		return nil, fmt.Errorf("create maintenance: %w", err)
	}

	return s.maintenanceToResponse(maintenance, equipment.Name), nil
}

func (s *equipmentService) ListMaintenance(ctx context.Context, page, pageSize int, equipmentID string, status *int) ([]dto.MaintenanceResponse, int64, error) {
	var equipID uuid.UUID
	if equipmentID != "" {
		parsed, err := parseRequiredUUID(equipmentID, "物资ID")
		if err != nil {
			return nil, 0, err
		}
		equipID = parsed
	}

	records, total, err := s.maintRepo.List(ctx, page, pageSize, equipID, status)
	if err != nil {
		return nil, 0, fmt.Errorf("list maintenance: %w", err)
	}

	result := make([]dto.MaintenanceResponse, 0, len(records))
	for _, m := range records {
		result = append(result, *s.maintenanceToResponse(&m, ""))
	}
	return result, total, nil
}

// ========== 库存盘点 ==========

func (s *equipmentService) CreateInventory(ctx context.Context, req *dto.CreateInventoryRequest, checkerID string) (*dto.InventoryResponse, error) {
	checkerUUID, _ := uuid.Parse(checkerID)
	equipID, err := parseRequiredUUID(req.EquipmentID, "物资ID")
	if err != nil {
		return nil, err
	}

	equipment, err := s.equipRepo.GetByID(ctx, equipID)
	if err != nil {
		return nil, fmt.Errorf("get equipment: %w", err)
	}
	if equipment == nil {
		return nil, response.NewError(27001, "物资不存在")
	}

	now := time.Now()
	inventory := &model.Inventory{
		ID:               uuid.New(),
		EquipmentID:      equipID,
		ExpectedQuantity: equipment.TotalQuantity,
		ActualQuantity:   req.ActualQuantity,
		Difference:       req.ActualQuantity - equipment.TotalQuantity,
		CheckerID:        &checkerUUID,
		CheckDate:        now,
		Remark:           req.Remark,
	}

	err = s.inventoryRepo.Create(ctx, nil, inventory)
	if err != nil {
		return nil, fmt.Errorf("create inventory: %w", err)
	}

	return s.inventoryToResponse(inventory, equipment.Name), nil
}

func (s *equipmentService) ListInventory(ctx context.Context, page, pageSize int, equipmentID string) ([]dto.InventoryResponse, int64, error) {
	var equipID uuid.UUID
	if equipmentID != "" {
		parsed, err := parseRequiredUUID(equipmentID, "物资ID")
		if err != nil {
			return nil, 0, err
		}
		equipID = parsed
	}

	records, total, err := s.inventoryRepo.List(ctx, page, pageSize, equipID)
	if err != nil {
		return nil, 0, fmt.Errorf("list inventory: %w", err)
	}

	result := make([]dto.InventoryResponse, 0, len(records))
	for _, inv := range records {
		result = append(result, *s.inventoryToResponse(&inv, ""))
	}
	return result, total, nil
}

// ========== 工具函数 ==========

func (s *equipmentService) equipmentToResponse(e *model.Equipment) *dto.EquipmentResponse {
	resp := &dto.EquipmentResponse{
		ID:                e.ID.String(),
		Name:              e.Name,
		Model:             e.Model,
		Category:          e.Category,
		CategoryText:      dto.GetCategoryText(e.Category),
		TotalQuantity:     e.TotalQuantity,
		AvailableQuantity: e.AvailableQuantity,
		Location:          e.Location,
		Status:            e.Status,
		StatusText:        dto.GetEquipmentStatusText(e.Status),
		QRCode:            e.QRCode,
		Description:       e.Description,
		CreatedAt:         e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         e.UpdatedAt.Format(time.RFC3339),
	}
	if e.ManagerID != nil {
		resp.ManagerID = e.ManagerID.String()
	}
	if e.PhotoFileID != nil {
		resp.PhotoFileID = e.PhotoFileID.String()
	}
	if e.PurchaseDate != nil {
		resp.PurchaseDate = e.PurchaseDate.Format("2006-01-02")
	}
	if e.PurchasePrice != nil {
		resp.PurchasePrice = *e.PurchasePrice
	}
	return resp
}

func (s *equipmentService) borrowToResponse(b *model.Borrow, equipName string) *dto.BorrowResponse {
	resp := &dto.BorrowResponse{
		ID:               b.ID.String(),
		EquipmentID:      b.EquipmentID.String(),
		EquipmentName:    equipName,
		BorrowerID:       b.BorrowerID.String(),
		Quantity:         b.Quantity,
		BorrowAt:         b.BorrowAt.Format(time.RFC3339),
		ExpectedReturnAt: b.ExpectedReturnAt.Format(time.RFC3339),
		Status:           b.Status,
		StatusText:       dto.GetBorrowStatusText(b.Status),
		ApprovedRemark:   b.ApprovedRemark,
		ReturnRemark:     b.ReturnRemark,
		Remark:           b.Remark,
		CreatedAt:        b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        b.UpdatedAt.Format(time.RFC3339),
	}
	if b.ActualReturnAt != nil {
		resp.ActualReturnAt = b.ActualReturnAt.Format(time.RFC3339)
	}
	if b.ApproverID != nil {
		resp.ApproverID = b.ApproverID.String()
	}
	if b.ApprovedAt != nil {
		resp.ApprovedAt = b.ApprovedAt.Format(time.RFC3339)
	}
	if b.ReturnCheckerID != nil {
		resp.ReturnCheckerID = b.ReturnCheckerID.String()
	}
	return resp
}

func (s *equipmentService) maintenanceToResponse(m *model.Maintenance, equipName string) *dto.MaintenanceResponse {
	resp := &dto.MaintenanceResponse{
		ID:            m.ID.String(),
		EquipmentID:   m.EquipmentID.String(),
		EquipmentName: equipName,
		Type:          m.Type,
		TypeText:      dto.GetMaintenanceTypeText(m.Type),
		Description:   m.Description,
		StartDate:     m.StartDate.Format("2006-01-02"),
		Status:        m.Status,
		StatusText:    dto.GetMaintenanceStatusText(m.Status),
		Remark:        m.Remark,
		CreatedAt:     m.CreatedAt.Format(time.RFC3339),
	}
	if m.Cost != nil {
		resp.Cost = *m.Cost
	}
	if m.EndDate != nil {
		resp.EndDate = m.EndDate.Format("2006-01-02")
	}
	if m.OperatorID != nil {
		resp.OperatorID = m.OperatorID.String()
	}
	return resp
}

func (s *equipmentService) inventoryToResponse(inv *model.Inventory, equipName string) *dto.InventoryResponse {
	resp := &dto.InventoryResponse{
		ID:               inv.ID.String(),
		EquipmentID:      inv.EquipmentID.String(),
		EquipmentName:    equipName,
		ExpectedQuantity: inv.ExpectedQuantity,
		ActualQuantity:   inv.ActualQuantity,
		Difference:       inv.Difference,
		CheckDate:        inv.CheckDate.Format("2006-01-02"),
		Remark:           inv.Remark,
		CreatedAt:        inv.CreatedAt.Format(time.RFC3339),
	}
	if inv.CheckerID != nil {
		resp.CheckerID = inv.CheckerID.String()
	}
	return resp
}

func parseRequiredUUID(value, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "参数错误: 无效的"+field)
	}
	return id, nil
}

func parseOptionalUUID(value, field string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := parseRequiredUUID(value, field)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
