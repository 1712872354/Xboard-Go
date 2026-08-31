package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"

	"gorm.io/gorm"
)

// UniProxyService 节点通讯服务接口（流量上报 + 心跳）
type UniProxyService interface {
	PushTraffic(nodeID uint, data []UniProxyTrafficItem) error
	ReportAlive(nodeID uint, aliveIPs []string) error
}

type uniProxyService struct {
	nodeRepo       repository.NodeRepository
	nodeUserRepo   repository.NodeUserRepository
	userRepo       repository.UserRepository
	trafficLogRepo repository.TrafficLogRepository
	trafficService TrafficService
	db             *gorm.DB
}

// NewUniProxyService 创建 UniProxy 服务
func NewUniProxyService(
	nodeRepo repository.NodeRepository,
	nodeUserRepo repository.NodeUserRepository,
	userRepo repository.UserRepository,
	trafficLogRepo repository.TrafficLogRepository,
	trafficService TrafficService,
) UniProxyService {
	return &uniProxyService{
		nodeRepo:       nodeRepo,
		nodeUserRepo:   nodeUserRepo,
		userRepo:       userRepo,
		trafficLogRepo: trafficLogRepo,
		trafficService: trafficService,
		db:             database.Get(),
	}
}

// UniProxyTrafficItem 流量上报项
type UniProxyTrafficItem struct {
	UserID   uint  `json:"user_id"`
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

// PushTraffic 接收节点上报的流量数据
// 使用 Redis Lua 脚本实现原子扣减，高性能且避免并发超扣
func (s *uniProxyService) PushTraffic(nodeID uint, data []UniProxyTrafficItem) error {
	if len(data) == 0 {
		return nil
	}

	node, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return errors.New("node not found")
	}

	rate := node.Rate
	if rate <= 0 {
		rate = 1.0
	}

	ctx := context.Background()
	now := time.Now()
	var trafficLogs []model.TrafficLog
	overLimitUsers := make([]uint, 0)

	for _, item := range data {
		if item.UserID == 0 {
			continue
		}

		// 按倍率换算实际扣减流量
		actualUpload := int64(float64(item.Upload) / rate)
		actualDownload := int64(float64(item.Download) / rate)

		// 使用 Redis 原子扣减
		isOverLimit, err := s.trafficService.AddTraffic(ctx, item.UserID, actualUpload, actualDownload, nodeID)
		if err != nil {
			logger.Sugar().Errorf("Failed to add traffic via redis: user_id=%d, err=%v", item.UserID, err)
			// Redis 失败，降级到数据库直接写入（在流量日志中记录）
		}

		if isOverLimit {
			overLimitUsers = append(overLimitUsers, item.UserID)
		}

		// 记录流量日志（异步批量写入优化：先收集，再批量写入）
		trafficLogs = append(trafficLogs, model.TrafficLog{
			UserID:     item.UserID,
			NodeID:     nodeID,
			Upload:     actualUpload,
			Download:   actualDownload,
			RecordedAt: now,
		})
	}

	// 批量写入流量日志
	if len(trafficLogs) > 0 {
		if err := s.db.Create(&trafficLogs).Error; err != nil {
			logger.Sugar().Errorf("Failed to batch create traffic logs: %v", err)
		}
	}

	if len(overLimitUsers) > 0 {
		logger.Sugar().Infof("%d users reached traffic limit from node %d", len(overLimitUsers), nodeID)
	}

	return nil
}

// ReportAlive 节点心跳上报
func (s *uniProxyService) ReportAlive(nodeID uint, aliveIPs []string) error {
	node, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return errors.New("node not found")
	}

	now := time.Now()
	node.Status = 1
	node.LastOnline = &now

	if err := s.nodeRepo.Update(node); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	logger.Sugar().Infof("Node %d reported alive, online IPs: %d", nodeID, len(aliveIPs))
	return nil
}
