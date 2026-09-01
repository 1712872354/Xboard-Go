package service

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// UserServerService 用户端服务器服务接口
type UserServerService interface {
	GetUserByID(id uint) (*model.User, error)
	GetUserNodes(user *model.User) ([]model.Node, error)
}

type userServerService struct {
	userRepo repository.UserRepository
	nodeRepo repository.NodeRepository
	planRepo repository.PlanRepository
	db       *gorm.DB
}

// NewUserServerService 创建用户端服务器服务
func NewUserServerService(userRepo repository.UserRepository, nodeRepo repository.NodeRepository, planRepo repository.PlanRepository) UserServerService {
	return &userServerService{
		userRepo: userRepo,
		nodeRepo: nodeRepo,
		planRepo: planRepo,
		db:       database.Get(),
	}
}

// GetUserByID 根据ID获取用户
func (s *userServerService) GetUserByID(id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserNodes 获取用户可用的节点列表
func (s *userServerService) GetUserNodes(user *model.User) ([]model.Node, error) {
	// 检查用户是否过期
	if user.HasExpired() {
		return []model.Node{}, nil
	}

	// 获取用户最近一笔已支付订单的套餐
	var latestOrder model.Order
	err := s.db.Where("user_id = ? AND status = ?", user.ID, model.OrderStatusPaid).
		Order("paid_at DESC").
		Preload("Plan").
		First(&latestOrder).Error

	if err != nil {
		// 没有已支付订单，返回所有可见节点（向后兼容）
		return s.nodeRepo.ListVisible()
	}

	plan := latestOrder.Plan
	if plan.NodeGroup == "" {
		// 套餐未指定节点组，返回所有可见节点
		return s.nodeRepo.ListVisible()
	}

	// 解析NodeGroup（支持逗号分隔的多个组ID）
	groupIDs := parseGroupIDs(plan.NodeGroup)
	if len(groupIDs) == 0 {
		return s.nodeRepo.ListVisible()
	}

	return s.nodeRepo.ListVisibleByGroups(groupIDs)
}
