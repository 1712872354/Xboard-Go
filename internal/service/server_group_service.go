package service

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// ServerGroupService 服务器分组服务接口
type ServerGroupService interface {
	Create(name, description, planIDs string, sort int) (*model.ServerGroup, error)
	GetByID(id uint) (*model.ServerGroup, error)
	Update(id uint, name, description, planIDs string, sort, status int) (*model.ServerGroup, error)
	Delete(id uint) error
	List(page, pageSize int) ([]model.ServerGroup, int64, error)
	ListAll() ([]model.ServerGroup, error)
}

type serverGroupService struct {
	serverGroupRepo repository.ServerGroupRepository
}

// NewServerGroupService 创建服务器分组服务
func NewServerGroupService(serverGroupRepo repository.ServerGroupRepository) ServerGroupService {
	return &serverGroupService{
		serverGroupRepo: serverGroupRepo,
	}
}

// Create 创建服务器分组
func (s *serverGroupService) Create(name, description, planIDs string, sort int) (*model.ServerGroup, error) {
	group := &model.ServerGroup{
		Name:        name,
		Description: description,
		PlanIDs:     planIDs,
		Sort:        sort,
		Status:      1,
	}

	if err := s.serverGroupRepo.Create(group); err != nil {
		return nil, err
	}

	return group, nil
}

// GetByID 根据ID获取服务器分组
func (s *serverGroupService) GetByID(id uint) (*model.ServerGroup, error) {
	group, err := s.serverGroupRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("server group not found")
	}
	return group, nil
}

// Update 更新服务器分组
func (s *serverGroupService) Update(id uint, name, description, planIDs string, sort, status int) (*model.ServerGroup, error) {
	group, err := s.serverGroupRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("server group not found")
	}

	group.Name = name
	group.Description = description
	group.PlanIDs = planIDs
	group.Sort = sort
	group.Status = status

	if err := s.serverGroupRepo.Update(group); err != nil {
		return nil, err
	}

	return group, nil
}

// Delete 删除服务器分组
func (s *serverGroupService) Delete(id uint) error {
	group, err := s.serverGroupRepo.GetByID(id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("server group not found")
	}

	return s.serverGroupRepo.Delete(id)
}

// List 服务器分组列表
func (s *serverGroupService) List(page, pageSize int) ([]model.ServerGroup, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.serverGroupRepo.List(page, pageSize)
}

// ListAll 获取所有服务器分组
func (s *serverGroupService) ListAll() ([]model.ServerGroup, error) {
	return s.serverGroupRepo.ListAll()
}
