package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepo 用户数据访问接口
type UserRepo interface {
	Create(ctx context.Context, tx *gorm.DB, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, tx *gorm.DB, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int, keyword string, status *int, departmentID uuid.UUID) ([]model.User, int64, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户 Repo
func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *userRepo) Create(ctx context.Context, tx *gorm.DB, user *model.User) error {
	return r.getDB(tx).WithContext(ctx).Create(user).Error
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *userRepo) Update(ctx context.Context, tx *gorm.DB, user *model.User) error {
	return r.getDB(tx).WithContext(ctx).Save(user).Error
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *userRepo) List(ctx context.Context, page, pageSize int, keyword string, status *int, departmentID uuid.UUID) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{})

	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if departmentID != uuid.Nil {
		query = query.Where("department_id = ?", departmentID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error

	return users, total, err
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"last_login_ip": ip,
		}).Error
}
