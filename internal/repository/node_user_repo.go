package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// NodeUserRepository 节点用户仓储接口
type NodeUserRepository interface {
	Create(nodeUser *model.NodeUser) error
	GetByUserAndNode(userID, nodeID uint) (*model.NodeUser, error)
	ListByNodeID(nodeID uint) ([]model.NodeUser, error)
	ListByUserID(userID uint) ([]model.NodeUser, error)
	Delete(id uint) error
	DeleteByUserAndNode(userID, nodeID uint) error
}

type nodeUserRepository struct {
	db *gorm.DB
}

// NewNodeUserRepository 创建节点用户仓储
func NewNodeUserRepository() NodeUserRepository {
	return &nodeUserRepository{
		db: database.Get(),
	}
}

// Create 创建节点用户关联
func (r *nodeUserRepository) Create(nodeUser *model.NodeUser) error {
	return r.db.Create(nodeUser).Error
}

// GetByUserAndNode 根据用户和节点获取关联
func (r *nodeUserRepository) GetByUserAndNode(userID, nodeID uint) (*model.NodeUser, error) {
	var nodeUser model.NodeUser
	err := r.db.Where("user_id = ? AND node_id = ?", userID, nodeID).First(&nodeUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &nodeUser, nil
}

// ListByNodeID 获取节点的所有用户
func (r *nodeUserRepository) ListByNodeID(nodeID uint) ([]model.NodeUser, error) {
	var nodeUsers []model.NodeUser
	err := r.db.Where("node_id = ?", nodeID).Preload("User").Find(&nodeUsers).Error
	return nodeUsers, err
}

// ListByUserID 获取用户的所有节点
func (r *nodeUserRepository) ListByUserID(userID uint) ([]model.NodeUser, error) {
	var nodeUsers []model.NodeUser
	err := r.db.Where("user_id = ?", userID).Preload("Node").Find(&nodeUsers).Error
	return nodeUsers, err
}

// Delete 删除节点用户关联
func (r *nodeUserRepository) Delete(id uint) error {
	return r.db.Delete(&model.NodeUser{}, id).Error
}

// DeleteByUserAndNode 根据用户和节点删除关联
func (r *nodeUserRepository) DeleteByUserAndNode(userID, nodeID uint) error {
	return r.db.Where("user_id = ? AND node_id = ?", userID, nodeID).Delete(&model.NodeUser{}).Error
}
