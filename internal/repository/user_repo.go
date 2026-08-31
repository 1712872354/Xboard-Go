package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetBySubscribeToken(token string) (*model.User, error)
	Update(user *model.User) error
	UpdateTraffic(userID uint, traffic int64) error
	List(page, pageSize int, keyword string) ([]model.User, int64, error)
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.Get(),
	}
}

// NewUserRepositoryWithTx 创建带事务的用户仓储
func NewUserRepositoryWithTx(tx *gorm.DB) UserRepository {
	return &userRepository{
		db: tx,
	}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetBySubscribeToken 根据订阅token获取用户
func (r *userRepository) GetBySubscribeToken(token string) (*model.User, error) {
	var user model.User
	err := r.db.Where("subscribe_token = ?", token).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateTraffic 更新用户已用流量（增量）
func (r *userRepository) UpdateTraffic(userID uint, traffic int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("used_traffic", gorm.Expr("used_traffic + ?", traffic)).Error
}

// List 用户列表
func (r *userRepository) List(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})

	if keyword != "" {
		query = query.Where("email LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}
