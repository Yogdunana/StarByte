package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/equipment/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EquipmentRepo 物资数据访问接口
type EquipmentRepo interface {
	Create(ctx context.Context, tx *gorm.DB, e *model.Equipment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Equipment, error)
	Update(ctx context.Context, tx *gorm.DB, e *model.Equipment) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.Equipment, int64, error)
	UpdateAvailableQuantity(ctx context.Context, tx *gorm.DB, id uuid.UUID, delta int) error
}

type equipmentRepo struct {
	db *gorm.DB
}

func NewEquipmentRepo(db *gorm.DB) EquipmentRepo {
	return &equipmentRepo{db: db}
}

func (r *equipmentRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *equipmentRepo) Create(ctx context.Context, tx *gorm.DB, e *model.Equipment) error {
	return r.getDB(tx).WithContext(ctx).Create(e).Error
}

func (r *equipmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Equipment, error) {
	var e model.Equipment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &e, err
}

func (r *equipmentRepo) Update(ctx context.Context, tx *gorm.DB, e *model.Equipment) error {
	return r.getDB(tx).WithContext(ctx).Save(e).Error
}

func (r *equipmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Equipment{}, id).Error
}

func (r *equipmentRepo) List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.Equipment, int64, error) {
	var items []model.Equipment
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Equipment{})

	if keyword != "" {
		query = query.Where("name LIKE ? OR model LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *equipmentRepo) UpdateAvailableQuantity(ctx context.Context, tx *gorm.DB, id uuid.UUID, delta int) error {
	return r.getDB(tx).WithContext(ctx).Model(&model.Equipment{}).
		Where("id = ?", id).
		UpdateColumn("available_quantity", gorm.Expr("available_quantity + ?", delta)).Error
}

// BorrowRepo 借用记录数据访问接口
type BorrowRepo interface {
	Create(ctx context.Context, tx *gorm.DB, b *model.Borrow) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Borrow, error)
	Update(ctx context.Context, tx *gorm.DB, b *model.Borrow) error
	List(ctx context.Context, page, pageSize int, equipmentID, borrowerID uuid.UUID, status *int) ([]model.Borrow, int64, error)
	FindOverdue(ctx context.Context, before time.Time) ([]model.Borrow, error)
}

type borrowRepo struct {
	db *gorm.DB
}

func NewBorrowRepo(db *gorm.DB) BorrowRepo {
	return &borrowRepo{db: db}
}

func (r *borrowRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *borrowRepo) Create(ctx context.Context, tx *gorm.DB, b *model.Borrow) error {
	return r.getDB(tx).WithContext(ctx).Create(b).Error
}

func (r *borrowRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Borrow, error) {
	var b model.Borrow
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&b).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &b, err
}

func (r *borrowRepo) Update(ctx context.Context, tx *gorm.DB, b *model.Borrow) error {
	return r.getDB(tx).WithContext(ctx).Save(b).Error
}

func (r *borrowRepo) List(ctx context.Context, page, pageSize int, equipmentID, borrowerID uuid.UUID, status *int) ([]model.Borrow, int64, error) {
	var borrows []model.Borrow
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Borrow{})

	if equipmentID != uuid.Nil {
		query = query.Where("equipment_id = ?", equipmentID)
	}
	if borrowerID != uuid.Nil {
		query = query.Where("borrower_id = ?", borrowerID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&borrows).Error
	return borrows, total, err
}

func (r *borrowRepo) FindOverdue(ctx context.Context, before time.Time) ([]model.Borrow, error) {
	var borrows []model.Borrow
	err := r.db.WithContext(ctx).
		Where("status IN (1, 2) AND expected_return_at < ? AND deleted_at IS NULL", before).
		Find(&borrows).Error
	return borrows, err
}

// MaintenanceRepo 维修记录数据访问接口
type MaintenanceRepo interface {
	Create(ctx context.Context, tx *gorm.DB, m *model.Maintenance) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Maintenance, error)
	Update(ctx context.Context, tx *gorm.DB, m *model.Maintenance) error
	List(ctx context.Context, page, pageSize int, equipmentID uuid.UUID, status *int) ([]model.Maintenance, int64, error)
}

type maintenanceRepo struct {
	db *gorm.DB
}

func NewMaintenanceRepo(db *gorm.DB) MaintenanceRepo {
	return &maintenanceRepo{db: db}
}

func (r *maintenanceRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *maintenanceRepo) Create(ctx context.Context, tx *gorm.DB, m *model.Maintenance) error {
	return r.getDB(tx).WithContext(ctx).Create(m).Error
}

func (r *maintenanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Maintenance, error) {
	var m model.Maintenance
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &m, err
}

func (r *maintenanceRepo) Update(ctx context.Context, tx *gorm.DB, m *model.Maintenance) error {
	return r.getDB(tx).WithContext(ctx).Save(m).Error
}

func (r *maintenanceRepo) List(ctx context.Context, page, pageSize int, equipmentID uuid.UUID, status *int) ([]model.Maintenance, int64, error) {
	var records []model.Maintenance
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Maintenance{})

	if equipmentID != uuid.Nil {
		query = query.Where("equipment_id = ?", equipmentID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

// InventoryRepo 盘点记录数据访问接口
type InventoryRepo interface {
	Create(ctx context.Context, tx *gorm.DB, inv *model.Inventory) error
	List(ctx context.Context, page, pageSize int, equipmentID uuid.UUID) ([]model.Inventory, int64, error)
}

type inventoryRepo struct {
	db *gorm.DB
}

func NewInventoryRepo(db *gorm.DB) InventoryRepo {
	return &inventoryRepo{db: db}
}

func (r *inventoryRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *inventoryRepo) Create(ctx context.Context, tx *gorm.DB, inv *model.Inventory) error {
	return r.getDB(tx).WithContext(ctx).Create(inv).Error
}

func (r *inventoryRepo) List(ctx context.Context, page, pageSize int, equipmentID uuid.UUID) ([]model.Inventory, int64, error) {
	var records []model.Inventory
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Inventory{})

	if equipmentID != uuid.Nil {
		query = query.Where("equipment_id = ?", equipmentID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("check_date DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}
