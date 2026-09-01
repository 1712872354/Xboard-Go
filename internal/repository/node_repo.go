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
	BatchMoveGroup(ids []uint, groupIDs string) error
	UpdateSort(id uint, sort int) error
	BatchSort(items []BatchSortItem) error
	BatchUpdateFields(ids []uint, show *int, enabled *bool, machineID *uint) error
	ResetTraffic(id uint) error
	BatchResetTraffic(ids []uint) error
	UpdateStatus(id uint, status int) error
	UpdateOnlineUserCount(id uint, count int) error
	IncrementTraffic(id uint, upload, download int64) error
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
		cond, args := model.GroupIDCondition(groupID)
		query = query.Where(cond, args...)
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
	cond, args := model.GroupIDCondition(groupID)
	err := r.db.Where(cond, args...).Order("sort ASC, id ASC").Find(&nodes).Error
	return nodes, err
}

// ListOnlineByGroup 按分组获取在线节点
func (r *nodeRepository) ListOnlineByGroup(groupID uint) ([]model.Node, error) {
	var nodes []model.Node
	cond, args := model.GroupIDCondition(groupID)
	err := r.db.Where("("+cond+") AND status = ?", append(args, 1)...).Order("sort ASC, id ASC").Find(&nodes).Error
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
	cond, args := model.GroupIDsContainAny(groupIDs)
	err := r.db.Where("show = ? AND status != ? AND "+cond, append([]interface{}{1, 2}, args...)...).
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

// BatchMoveGroup 批量移动节点分组（设置 group_ids）
func (r *nodeRepository) BatchMoveGroup(ids []uint, groupIDs string) error {
	return r.db.Model(&model.Node{}).Where("id IN ?", ids).Update("group_ids", groupIDs).Error
}

// UpdateSort 更新节点排序值
func (r *nodeRepository) UpdateSort(id uint, sort int) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).Update("sort", sort).Error
}

// BatchSortItem 批量排序项
type BatchSortItem struct {
	ID    uint
	Order int
}

// BatchSort 批量更新节点排序值
func (r *nodeRepository) BatchSort(items []BatchSortItem) error {
	tx := r.db.Begin()
	for _, item := range items {
		if err := tx.Model(&model.Node{}).Where("id = ?", item.ID).Update("sort", item.Order).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// BatchUpdateFields 批量更新节点属性
func (r *nodeRepository) BatchUpdateFields(ids []uint, show *int, enabled *bool, machineID *uint) error {
	updates := map[string]interface{}{}
	if show != nil {
		updates["show"] = *show
	}
	if enabled != nil {
		updates["enabled"] = *enabled
	}
	if machineID != nil {
		updates["machine_id"] = *machineID
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.Node{}).Where("id IN ?", ids).Updates(updates).Error
}

// ResetTraffic 重置节点流量统计
func (r *nodeRepository) ResetTraffic(id uint) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).
		Updates(map[string]interface{}{"upload_traffic": 0, "download_traffic": 0}).Error
}

// BatchResetTraffic 批量重置节点流量统计
func (r *nodeRepository) BatchResetTraffic(ids []uint) error {
	return r.db.Model(&model.Node{}).Where("id IN ?", ids).
		Updates(map[string]interface{}{"upload_traffic": 0, "download_traffic": 0}).Error
}

// UpdateStatus 更新单个节点状态
func (r *nodeRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateOnlineUserCount 更新节点在线用户数
func (r *nodeRepository) UpdateOnlineUserCount(id uint, count int) error {
	return r.db.Model(&model.Node{}).Where("id = ?", id).Update("online_user_count", count).Error
}

// IncrementTraffic 增加节点上/下行流量
func (r *nodeRepository) IncrementTraffic(id uint, upload, download int64) error {
	updates := map[string]interface{}{}
	if upload > 0 {
		updates["upload_traffic"] = gorm.Expr("upload_traffic + ?", upload)
	}
	if download > 0 {
		updates["download_traffic"] = gorm.Expr("download_traffic + ?", download)
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.Node{}).Where("id = ?", id).UpdateColumns(updates).Error
}
