package service

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// ServerRouteService 服务器路由服务接口
type ServerRouteService interface {
	Create(groupID uint, name, match, action, target string, sort int) (*model.ServerRoute, error)
	GetByID(id uint) (*model.ServerRoute, error)
	Update(id uint, groupID uint, name, match, action, target string, sort, status int) (*model.ServerRoute, error)
	Delete(id uint) error
	List(page, pageSize int, groupID uint) ([]model.ServerRoute, int64, error)
	ListByGroup(groupID uint) ([]model.ServerRoute, error)
}

type serverRouteService struct {
	serverRouteRepo repository.ServerRouteRepository
}

// NewServerRouteService 创建服务器路由服务
func NewServerRouteService(serverRouteRepo repository.ServerRouteRepository) ServerRouteService {
	return &serverRouteService{
		serverRouteRepo: serverRouteRepo,
	}
}

// Create 创建服务器路由
func (s *serverRouteService) Create(groupID uint, name, match, action, target string, sort int) (*model.ServerRoute, error) {
	route := &model.ServerRoute{
		GroupID: groupID,
		Name:    name,
		Match:   match,
		Action:  action,
		Target:  target,
		Sort:    sort,
		Status:  1,
	}

	if err := s.serverRouteRepo.Create(route); err != nil {
		return nil, err
	}

	return route, nil
}

// GetByID 根据ID获取服务器路由
func (s *serverRouteService) GetByID(id uint) (*model.ServerRoute, error) {
	route, err := s.serverRouteRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return nil, errors.New("server route not found")
	}
	return route, nil
}

// Update 更新服务器路由
func (s *serverRouteService) Update(id uint, groupID uint, name, match, action, target string, sort, status int) (*model.ServerRoute, error) {
	route, err := s.serverRouteRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return nil, errors.New("server route not found")
	}

	route.GroupID = groupID
	route.Name = name
	route.Match = match
	route.Action = action
	route.Target = target
	route.Sort = sort
	route.Status = status

	if err := s.serverRouteRepo.Update(route); err != nil {
		return nil, err
	}

	return route, nil
}

// Delete 删除服务器路由
func (s *serverRouteService) Delete(id uint) error {
	route, err := s.serverRouteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if route == nil {
		return errors.New("server route not found")
	}

	return s.serverRouteRepo.Delete(id)
}

// List 服务器路由列表
func (s *serverRouteService) List(page, pageSize int, groupID uint) ([]model.ServerRoute, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.serverRouteRepo.List(page, pageSize, groupID)
}

// ListByGroup 根据分组获取路由列表
func (s *serverRouteService) ListByGroup(groupID uint) ([]model.ServerRoute, error) {
	return s.serverRouteRepo.ListByGroup(groupID)
}
