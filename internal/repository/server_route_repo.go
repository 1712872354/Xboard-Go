package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// ServerRouteRepository 服务器路由仓储接口
type ServerRouteRepository interface {
	Create(route *model.ServerRoute) error
	GetByID(id uint) (*model.ServerRoute, error)
	Update(route *model.ServerRoute) error
	Delete(id uint) error
	List(page, pageSize int, groupID uint) ([]model.ServerRoute, int64, error)
	ListByGroup(groupID uint) ([]model.ServerRoute, error)
}

type serverRouteRepository struct {
	db *gorm.DB
}

// NewServerRouteRepository 创建服务器路由仓储
func NewServerRouteRepository() ServerRouteRepository {
	return &serverRouteRepository{
		db: database.Get(),
	}
}

// Create 创建服务器路由
func (r *serverRouteRepository) Create(route *model.ServerRoute) error {
	return r.db.Create(route).Error
}

// GetByID 根据ID获取服务器路由
func (r *serverRouteRepository) GetByID(id uint) (*model.ServerRoute, error) {
	var route model.ServerRoute
	err := r.db.First(&route, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

// Update 更新服务器路由
func (r *serverRouteRepository) Update(route *model.ServerRoute) error {
	return r.db.Save(route).Error
}

// Delete 删除服务器路由
func (r *serverRouteRepository) Delete(id uint) error {
	return r.db.Delete(&model.ServerRoute{}, id).Error
}

// List 服务器路由列表
func (r *serverRouteRepository) List(page, pageSize int, groupID uint) ([]model.ServerRoute, int64, error) {
	var routes []model.ServerRoute
	var total int64

	query := r.db.Model(&model.ServerRoute{})

	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&routes).Error; err != nil {
		return nil, 0, err
	}

	return routes, total, nil
}

// ListByGroup 根据分组获取路由列表
func (r *serverRouteRepository) ListByGroup(groupID uint) ([]model.ServerRoute, error) {
	var routes []model.ServerRoute
	err := r.db.Where("group_id = ? AND status = ?", groupID, 1).Order("sort ASC, id DESC").Find(&routes).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}
