package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// NodeRepository 节点仓储接口
type NodeRepository interface {
	Create(node *model.Node) error
	GetByID(id uint) (*model.Node, error)
	Update(node *model.Node) error
	Delete(id uint) error
	List(page, pageSize int, groupID uint) ([]model.Node, int64, error)
	ListByGroup(groupID uint) ([]model.Node, error)
	ListOnlineByGroup(groupID uint) ([]model.Node, error)
	ListAllOnline() ([]model.Node, error)
	ListVisible() ([]model.Node, error)
	ListVisibleByGroups(groupIDs []uint) ([]model.Node, error)
	BatchUpdateStatus(ids []uint, status int) error
	BatchDelete(ids []uint) error
	BatchMoveGroup(ids []uint, groupID uint) error
	UpdateSort(id uint, sort int) error
	ResetTraffic(id uint) error
	BatchResetTraffic(ids []uint) error
	UpdateStatus(id uint, status int) error
	UpdateOnlineUserCount(id uint, count int) error
	IncrementTrafficUsed(id uint, delta int64) error
}

type nodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository 创建节点仓储
func NewNodeRepository() NodeRepository {
	return &nodeRepository{
		db: database.Get(),
	}
}

// Create 创建节点
func (r *nodeRepository) Create(node *model.Node) error {
	return r.db.Create(node).Error
}

// GetByID 根据ID获取节点
func (r *nodeRepository) GetByID(id uint) (*model.Node, error) {
	var node model.Node
	err := r.db.First(&node, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &node, nil
}

// Update 更新节点
func (r *nodeRepository) Update(node *model.Node) error {
	return r.db.Save(node).Error
}

// Delete 删除节点
func (r *nodeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Node{}, id).Error
}

// List 节点列表
func (r *nodeRepository) List(page, pageSize int, groupID uint) ([]model.Node, int64, error) {
	var nodes []model.Node
	var total int64

	query := r.db.Model(&model.Node{})

	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

// ListByGroup 按分组获取节点
func (r *nodeRepository) ListByGroup(groupID uint) ([]model.Node, error) {
	var nodes []model.Node
	err := r.db.Where("group_id = ?", groupID).Order("sort ASC, id ASC").Find(&nodes).Error
	return nodes, err
}

// ListOnlineByGroup 按分组获取在线节点
func (r *nodeRepository) ListOnlineByGroup(groupID uint) ([]model.Node, error) {
	var nodes []model.Node
	err := r.db.Where("group_id = ? AND status = ?", groupID, 1).Order("sort ASC, id ASC").Find(&nodes).Error
	return nodes, err
}

// ListAllOnline 获取所有在线节点
func (r *nodeRepository) ListAllOnline() ([]model.Node, error) {
	var nodes []model.Node
	err := r.db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&nodes).Error
	return nodes, err
}

// ListVisible 获取所有可见节点（用于订阅输出）
func (r *nodeRepository) ListVisible() ([]model.Node, error) {
	var nodes []model.Node
	err := r.db.Where("show = ? AND status != ?", 1, 2).Order("sort ASC, id ASC").Find(&nodes).Error
	return nodes, err
}

// ListVisibleByGroups 按多个分组获取可见节点（用于订阅输出）
func (r *nodeRepository) ListVisibleByGroups(groupIDs []uint) ([]model.Node, error) {
	if len(groupIDs) == 0 {
		return r.ListVisible()
	}
	var nodes []model.Node
	err := r.db.Where("show = ? AND status != ? AND group_id IN ?", 1, 2, groupIDs).
		Order("sort ASC, id ASC").Find(&nodes).Error
	return nodes, err
}

// BatchUpdateStatus 批量更新节点状态
func (r *nodeRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.Node{}).Where("id IN ?", ids).Update("status", status).Error
}

// BatchDelete 批量删除节点
func (r *nodeRepository) BatchDelete(ids []uint) error {
	return r.db.Delete(&model.Node{}, ids).Error
}

// BatchMoveGroup 批量移动节点分组
func (r *nodeRepository) BatchMoveGroup(ids []uint, groupID uint) error {
	return r.db.Model(&model.Node{}).Where("id IN ?", ids).Update("group_id", groupID).Error
}

// UpdateSort 更新节点排序值
func (r *nodeRepository) UpdateSort(id uint, sort int) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).Update("sort", sort).Error
}

// ResetTraffic 重置节点流量统计
func (r *nodeRepository) ResetTraffic(id uint) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).Update("traffic_used", 0).Error
}

// BatchResetTraffic 批量重置节点流量统计
func (r *nodeRepository) BatchResetTraffic(ids []uint) error {
	return r.db.Model(&model.Node{}).Where("id IN ?", ids).Update("traffic_used", 0).Error
}

// UpdateStatus 更新单个节点状态
func (r *nodeRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateOnlineUserCount 更新节点在线用户数
func (r *nodeRepository) UpdateOnlineUserCount(id uint, count int) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).Update("online_user_count", count).Error
}

// IncrementTrafficUsed 增加节点累计流量
func (r *nodeRepository) IncrementTrafficUsed(id uint, delta int64) error {
	if delta <= 0 {
		return nil
	}
	return r.db.Model(&model.Node{}).Where("id = ?", id).
		UpdateColumn("traffic_used", gorm.Expr("traffic_used + ?", delta)).Error
}
