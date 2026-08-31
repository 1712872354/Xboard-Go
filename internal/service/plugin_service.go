package service

import (
	"errors"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// PluginService 插件服务接口
type PluginService interface {
	Create(name, title, description, version, author, homepage, config string) (*model.Plugin, error)
	GetByID(id uint) (*model.Plugin, error)
	GetByName(name string) (*model.Plugin, error)
	Update(id uint, name, title, description, version, author, homepage, config string, status int) (*model.Plugin, error)
	Delete(id uint) error
	List(page, pageSize int) ([]model.Plugin, int64, error)
	ListEnabled() ([]model.Plugin, error)
	UpdateStatus(id uint, status int) error
	Enable(id uint) error
	Disable(id uint) error
}

type pluginService struct {
	pluginRepo repository.PluginRepository
}

// NewPluginService 创建插件服务
func NewPluginService(pluginRepo repository.PluginRepository) PluginService {
	return &pluginService{
		pluginRepo: pluginRepo,
	}
}

// Create 创建插件
func (s *pluginService) Create(name, title, description, version, author, homepage, config string) (*model.Plugin, error) {
	// 检查插件名称是否已存在
	existing, err := s.pluginRepo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("plugin name already exists")
	}

	now := time.Now()
	plugin := &model.Plugin{
		Name:        name,
		Title:       title,
		Description: description,
		Version:     version,
		Author:      author,
		Homepage:    homepage,
		Config:      config,
		Status:      0,
		InstalledAt: &now,
	}

	if err := s.pluginRepo.Create(plugin); err != nil {
		return nil, err
	}

	return plugin, nil
}

// GetByID 根据ID获取插件
func (s *pluginService) GetByID(id uint) (*model.Plugin, error) {
	plugin, err := s.pluginRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, errors.New("plugin not found")
	}
	return plugin, nil
}

// GetByName 根据名称获取插件
func (s *pluginService) GetByName(name string) (*model.Plugin, error) {
	plugin, err := s.pluginRepo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, errors.New("plugin not found")
	}
	return plugin, nil
}

// Update 更新插件
func (s *pluginService) Update(id uint, name, title, description, version, author, homepage, config string, status int) (*model.Plugin, error) {
	plugin, err := s.pluginRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, errors.New("plugin not found")
	}

	// 如果修改了名称，检查新名称是否已存在
	if name != plugin.Name {
		existing, err := s.pluginRepo.GetByName(name)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("plugin name already exists")
		}
	}

	plugin.Name = name
	plugin.Title = title
	plugin.Description = description
	plugin.Version = version
	plugin.Author = author
	plugin.Homepage = homepage
	plugin.Config = config
	plugin.Status = status

	if err := s.pluginRepo.Update(plugin); err != nil {
		return nil, err
	}

	return plugin, nil
}

// Delete 删除插件
func (s *pluginService) Delete(id uint) error {
	plugin, err := s.pluginRepo.GetByID(id)
	if err != nil {
		return err
	}
	if plugin == nil {
		return errors.New("plugin not found")
	}

	return s.pluginRepo.Delete(id)
}

// List 插件列表
func (s *pluginService) List(page, pageSize int) ([]model.Plugin, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.pluginRepo.List(page, pageSize)
}

// ListEnabled 获取已启用的插件列表
func (s *pluginService) ListEnabled() ([]model.Plugin, error) {
	return s.pluginRepo.ListEnabled()
}

// UpdateStatus 更新插件状态
func (s *pluginService) UpdateStatus(id uint, status int) error {
	plugin, err := s.pluginRepo.GetByID(id)
	if err != nil {
		return err
	}
	if plugin == nil {
		return errors.New("plugin not found")
	}

	return s.pluginRepo.UpdateStatus(id, status)
}

// Enable 启用插件
func (s *pluginService) Enable(id uint) error {
	return s.UpdateStatus(id, 1)
}

// Disable 禁用插件
func (s *pluginService) Disable(id uint) error {
	return s.UpdateStatus(id, 0)
}
