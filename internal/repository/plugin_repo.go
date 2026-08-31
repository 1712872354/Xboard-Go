package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// PluginRepository 插件仓储接口
type PluginRepository interface {
	Create(plugin *model.Plugin) error
	GetByID(id uint) (*model.Plugin, error)
	GetByName(name string) (*model.Plugin, error)
	Update(plugin *model.Plugin) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.Plugin, int64, error)
	ListEnabled() ([]model.Plugin, error)
	UpdateStatus(id uint, status int) error
}

type pluginRepository struct {
	db *gorm.DB
}

// NewPluginRepository 创建插件仓储
func NewPluginRepository() PluginRepository {
	return &pluginRepository{
		db: database.Get(),
	}
}

// Create 创建插件
func (r *pluginRepository) Create(plugin *model.Plugin) error {
	return r.db.Create(plugin).Error
}

// GetByID 根据ID获取插件
func (r *pluginRepository) GetByID(id uint) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.db.First(&plugin, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plugin, nil
}

// GetByName 根据名称获取插件
func (r *pluginRepository) GetByName(name string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.db.Where("name = ?", name).First(&plugin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plugin, nil
}

// Update 更新插件
func (r *pluginRepository) Update(plugin *model.Plugin) error {
	return r.db.Save(plugin).Error
}

// Delete 删除插件
func (r *pluginRepository) Delete(id uint) error {
	return r.db.Delete(&model.Plugin{}, id).Error
}

// List 插件列表
func (r *pluginRepository) List(page, pageSize int) ([]model.Plugin, int64, error) {
	var plugins []model.Plugin
	var total int64

	query := r.db.Model(&model.Plugin{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&plugins).Error; err != nil {
		return nil, 0, err
	}

	return plugins, total, nil
}

// ListEnabled 获取已启用的插件列表
func (r *pluginRepository) ListEnabled() ([]model.Plugin, error) {
	var plugins []model.Plugin
	err := r.db.Where("status = ?", 1).Order("id ASC").Find(&plugins).Error
	if err != nil {
		return nil, err
	}
	return plugins, nil
}

// UpdateStatus 更新插件状态
func (r *pluginRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Plugin{}).Where("id = ?", id).Update("status", status).Error
}
