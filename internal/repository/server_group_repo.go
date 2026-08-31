package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// ServerGroupRepository 服务器分组仓储接口
type ServerGroupRepository interface {
	Create(group *model.ServerGroup) error
	GetByID(id uint) (*model.ServerGroup, error)
	Update(group *model.ServerGroup) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.ServerGroup, int64, error)
	ListAll() ([]model.ServerGroup, error)
}

type serverGroupRepository struct {
	db *gorm.DB
}

// NewServerGroupRepository 创建服务器分组仓储
func NewServerGroupRepository() ServerGroupRepository {
	return &serverGroupRepository{
		db: database.Get(),
	}
}

// Create 创建服务器分组
func (r *serverGroupRepository) Create(group *model.ServerGroup) error {
	return r.db.Create(group).Error
}

// GetByID 根据ID获取服务器分组
func (r *serverGroupRepository) GetByID(id uint) (*model.ServerGroup, error) {
	var group model.ServerGroup
	err := r.db.First(&group, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// Update 更新服务器分组
func (r *serverGroupRepository) Update(group *model.ServerGroup) error {
	return r.db.Save(group).Error
}

// Delete 删除服务器分组
func (r *serverGroupRepository) Delete(id uint) error {
	return r.db.Delete(&model.ServerGroup{}, id).Error
}

// List 服务器分组列表
func (r *serverGroupRepository) List(page, pageSize int) ([]model.ServerGroup, int64, error) {
	var groups []model.ServerGroup
	var total int64

	query := r.db.Model(&model.ServerGroup{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// ListAll 获取所有服务器分组
func (r *serverGroupRepository) ListAll() ([]model.ServerGroup, error) {
	var groups []model.ServerGroup
	err := r.db.Where("status = ?", 1).Order("sort ASC, id DESC").Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}
