package service

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/logger"

	"gorm.io/gorm"
)

// NodeHealthService 节点健康检查与在线状态管理服务
type NodeHealthService interface {
	// Start 启动后台健康检查（阻塞，应运行在goroutine中）
	Start(ctx context.Context)
	// CheckNode 立即检查单个节点健康状态
	CheckNode(node *model.Node) bool
	// UpdateOnlineUserCount 更新节点在线用户数
	UpdateOnlineUserCount(nodeID uint, count int)
	// CheckMachineOffline 检查服务器机器是否超时离线
	CheckMachineOffline()
}

type nodeHealthService struct {
	nodeRepo    repository.NodeRepository
	machineRepo repository.ServerMachineRepository
	db          *gorm.DB

	// 离线超时时间（秒）：gRPC节点超过此时间未上报则标记离线
	nodeOfflineTimeout time.Duration
	// 机器离线超时时间
	machineOfflineTimeout time.Duration
}

// NewNodeHealthService 创建节点健康检查服务
func NewNodeHealthService(
	nodeRepo repository.NodeRepository,
	machineRepo repository.ServerMachineRepository,
) NodeHealthService {
	return &nodeHealthService{
		nodeRepo:              nodeRepo,
		machineRepo:           machineRepo,
		db:                    nil, // 通过 repo 操作
		nodeOfflineTimeout:    3 * time.Minute,
		machineOfflineTimeout: 5 * time.Minute,
	}
}

// Start 启动后台健康检查
func (s *nodeHealthService) Start(ctx context.Context) {
	// 节点离线检测定时器（每60秒检查一次）
	nodeCheckTicker := time.NewTicker(60 * time.Second)
	defer nodeCheckTicker.Stop()

	// 机器离线检测定时器（每120秒检查一次）
	machineCheckTicker := time.NewTicker(120 * time.Second)
	defer machineCheckTicker.Stop()

	// TCP健康检查定时器（每60秒检查一次）
	healthCheckTicker := time.NewTicker(60 * time.Second)
	defer healthCheckTicker.Stop()

	logger.Sugar().Info("NodeHealthService started")

	for {
		select {
		case <-ctx.Done():
			logger.Sugar().Info("NodeHealthService stopped")
			return
		case <-nodeCheckTicker.C:
			s.checkNodesOnline()
		case <-machineCheckTicker.C:
			s.CheckMachineOffline()
		case <-healthCheckTicker.C:
			s.runHealthChecks()
		}
	}
}

// checkNodesOnline 检查所有在线节点的最后上报时间，超时则标记离线
func (s *nodeHealthService) checkNodesOnline() {
	nodes, err := s.nodeRepo.ListAllOnline()
	if err != nil {
		logger.Sugar().Errorf("Health check: failed to list online nodes: %v", err)
		return
	}

	now := time.Now()
	offlineThreshold := now.Add(-s.nodeOfflineTimeout)

	var wg sync.WaitGroup
	for _, node := range nodes {
		if node.LastOnline != nil && node.LastOnline.Before(offlineThreshold) {
			wg.Add(1)
			go func(nodeID uint) {
				defer wg.Done()
				if err := s.nodeRepo.UpdateStatus(nodeID, 0); err != nil {
					logger.Sugar().Errorf("Health check: failed to mark node %d offline: %v", nodeID, err)
				} else {
					logger.Sugar().Infof("Health check: marked node %d offline (last online: %s)", nodeID, offlineThreshold.Format(time.RFC3339))
				}
			}(node.ID)
		}
	}
	wg.Wait()
}

// runHealthChecks 对配置了健康检查的节点执行TCP连通性检测
func (s *nodeHealthService) runHealthChecks() {
	nodes, err := s.nodeRepo.ListVisible()
	if err != nil {
		logger.Sugar().Errorf("Health check: failed to list visible nodes: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, node := range nodes {
		if node.HealthCheckPort > 0 {
			wg.Add(1)
			go func(n model.Node) {
				defer wg.Done()
				alive := s.CheckNode(&n)
				if !alive && n.Status == 1 {
					// 节点配置了健康检查但不可达，标记离线
					if err := s.nodeRepo.UpdateStatus(n.ID, 0); err != nil {
						logger.Sugar().Errorf("Health check: failed to mark node %d offline: %v", n.ID, err)
					}
				}
			}(node)
		}
	}
	wg.Wait()
}

// CheckNode 检查单个节点的TCP连通性
func (s *nodeHealthService) CheckNode(node *model.Node) bool {
	checkPort := node.HealthCheckPort
	if checkPort <= 0 {
		checkPort = node.Port
	}
	timeout := time.Duration(node.HealthCheckTimeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	addr := net.JoinHostPort(node.Address, fmt.Sprintf("%d", checkPort))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		logger.Sugar().Debugf("Health check failed for node %d (%s): %v", node.ID, addr, err)
		return false
	}
	conn.Close()
	return true
}

// UpdateOnlineUserCount 更新节点在线用户数
func (s *nodeHealthService) UpdateOnlineUserCount(nodeID uint, count int) {
	if err := s.nodeRepo.UpdateOnlineUserCount(nodeID, count); err != nil {
		logger.Sugar().Errorf("Failed to update online user count for node %d: %v", nodeID, err)
	}
}

// CheckMachineOffline 检查服务器机器是否超时离线
func (s *nodeHealthService) CheckMachineOffline() {
	machines, err := s.machineRepo.ListAll()
	if err != nil {
		logger.Sugar().Errorf("Health check: failed to list machines: %v", err)
		return
	}

	now := time.Now()
	offlineThreshold := now.Add(-s.machineOfflineTimeout)

	for _, machine := range machines {
		if machine.Status == 1 && machine.LastCheckAt != nil && machine.LastCheckAt.Before(offlineThreshold) {
			if err := s.machineRepo.UpdateStatus(machine.ID, 0); err != nil {
				logger.Sugar().Errorf("Health check: failed to mark machine %d offline: %v", machine.ID, err)
			} else {
				logger.Sugar().Infof("Health check: marked machine %d offline (last check: %s)",
					machine.ID, machine.LastCheckAt.Format(time.RFC3339))
			}
		}
	}
}
