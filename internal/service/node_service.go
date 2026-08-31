package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// NodeService 节点服务接口
type NodeService interface {
	CreateNode(name, nodeType, address string, port int, serverInfo string, groupID uint, rate float64, parentID uint) (*model.Node, error)
	CreateNodeEx(req CreateNodeRequest) (*model.Node, error)
	GetNodeByID(id uint) (*model.Node, error)
	UpdateNode(id uint, name, nodeType, address string, port int, serverInfo string, groupID uint, rate float64, status int, parentID uint) (*model.Node, error)
	UpdateNodeEx(id uint, req UpdateNodeRequest) (*model.Node, error)
	DeleteNode(id uint) error
	ListNodes(page, pageSize int, groupID uint) ([]model.Node, int64, error)
	ListOnlineNodes() ([]model.Node, error)
	UpdateNodeStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) error
	BatchDelete(ids []uint) error
	BatchMoveGroup(ids []uint, groupID uint) error
	// 复制节点
	CopyNode(id uint) (*model.Node, error)
	// 更新节点排序
	UpdateNodeSort(id uint, sort int) error
	// 重置节点流量
	ResetNodeTraffic(id uint) error
	// 批量重置节点流量
	BatchResetNodeTraffic(ids []uint) error
}

// CreateNodeRequest 创建节点请求结构
type CreateNodeRequest struct {
	Name                string  `json:"name"`
	Type                string  `json:"type"`
	Address             string  `json:"address"`
	Port                int     `json:"port"`
	ServerInfo          string  `json:"server_info"`
	GroupID             uint    `json:"group_id"`
	Rate                float64 `json:"rate"`
	ParentID            uint    `json:"parent_id"`
	Sort                int     `json:"sort"`
	Show                int     `json:"show"`
	Tags                string  `json:"tags"`
	HealthCheckPort     int     `json:"health_check_port"`
	HealthCheckInterval int     `json:"health_check_interval"`
	HealthCheckTimeout  int     `json:"health_check_timeout"`
	HealthCheckType     string  `json:"health_check_type"`
}

// UpdateNodeRequest 更新节点请求结构
type UpdateNodeRequest struct {
	Name                string  `json:"name"`
	Type                string  `json:"type"`
	Address             string  `json:"address"`
	Port                int     `json:"port"`
	ServerInfo          string  `json:"server_info"`
	GroupID             uint    `json:"group_id"`
	Rate                float64 `json:"rate"`
	Status              int     `json:"status"`
	ParentID            uint    `json:"parent_id"`
	Sort                *int    `json:"sort"` // 指针区分0值和未设置
	Show                *int    `json:"show"`
	Tags                string  `json:"tags"`
	HealthCheckPort     int     `json:"health_check_port"`
	HealthCheckInterval int     `json:"health_check_interval"`
	HealthCheckTimeout  int     `json:"health_check_timeout"`
	HealthCheckType     string  `json:"health_check_type"`
}

type nodeService struct {
	nodeRepo repository.NodeRepository
}

// NewNodeService 创建节点服务
func NewNodeService(nodeRepo repository.NodeRepository) NodeService {
	return &nodeService{
		nodeRepo: nodeRepo,
	}
}

// CreateNode 创建节点
func (s *nodeService) CreateNode(name, nodeType, address string, port int, serverInfo string, groupID uint, rate float64, parentID uint) (*model.Node, error) {
	if name == "" {
		return nil, errors.New("node name is required")
	}
	if nodeType == "" {
		return nil, errors.New("node type is required")
	}
	if address == "" {
		return nil, errors.New("node address is required")
	}
	if port <= 0 || port > 65535 {
		return nil, errors.New("invalid port number")
	}

	// 验证节点类型
	validTypes := map[string]bool{
		model.NodeTypeVMess:       true,
		model.NodeTypeVLESS:       true,
		model.NodeTypeTrojan:      true,
		model.NodeTypeShadowsocks: true,
		model.NodeTypeHysteria2:   true,
		model.NodeTypeHysteria:    true,
		model.NodeTypeTUIC:        true,
		model.NodeTypeAnyTLS:      true,
		model.NodeTypeNaive:       true,
		model.NodeTypeSocks:       true,
		model.NodeTypeHTTP:        true,
		model.NodeTypeMieru:       true,
	}
	if !validTypes[nodeType] {
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}

	// 验证 serverInfo 是否为有效 JSON
	if serverInfo != "" {
		var tmp map[string]interface{}
		if err := json.Unmarshal([]byte(serverInfo), &tmp); err != nil {
			return nil, errors.New("server_info must be valid JSON")
		}
	}

	if rate <= 0 {
		rate = 1.0
	}

	node := &model.Node{
		Name:       name,
		Type:       nodeType,
		Address:    address,
		Port:       port,
		ServerInfo: serverInfo,
		GroupID:    groupID,
		Rate:       rate,
		Status:     1,
		ParentID:   parentID,
	}

	if err := s.nodeRepo.Create(node); err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	return node, nil
}

// GetNodeByID 根据ID获取节点
func (s *nodeService) GetNodeByID(id uint) (*model.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found")
	}
	return node, nil
}

// UpdateNode 更新节点
func (s *nodeService) UpdateNode(id uint, name, nodeType, address string, port int, serverInfo string, groupID uint, rate float64, status int, parentID uint) (*model.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found")
	}

	if name != "" {
		node.Name = name
	}
	if nodeType != "" {
		node.Type = nodeType
	}
	if address != "" {
		node.Address = address
	}
	if port > 0 && port <= 65535 {
		node.Port = port
	}
	if serverInfo != "" {
		// 验证 JSON
		var tmp map[string]interface{}
		if err := json.Unmarshal([]byte(serverInfo), &tmp); err != nil {
			return nil, errors.New("server_info must be valid JSON")
		}
		node.ServerInfo = serverInfo
	}
	if groupID > 0 {
		node.GroupID = groupID
	}
	if rate > 0 {
		node.Rate = rate
	}
	if status >= 0 && status <= 2 {
		node.Status = status
	}
	if parentID >= 0 {
		node.ParentID = parentID
	}

	if err := s.nodeRepo.Update(node); err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}

	return node, nil
}

// CreateNodeEx 扩展创建节点（支持新字段）
func (s *nodeService) CreateNodeEx(req CreateNodeRequest) (*model.Node, error) {
	if req.Name == "" {
		return nil, errors.New("node name is required")
	}
	if req.Type == "" {
		return nil, errors.New("node type is required")
	}
	if req.Address == "" {
		return nil, errors.New("node address is required")
	}
	if req.Port <= 0 || req.Port > 65535 {
		return nil, errors.New("invalid port number")
	}

	validTypes := map[string]bool{
		model.NodeTypeVMess: true, model.NodeTypeVLESS: true, model.NodeTypeTrojan: true,
		model.NodeTypeShadowsocks: true, model.NodeTypeHysteria2: true, model.NodeTypeHysteria: true,
		model.NodeTypeTUIC: true, model.NodeTypeAnyTLS: true, model.NodeTypeNaive: true,
		model.NodeTypeSocks: true, model.NodeTypeHTTP: true, model.NodeTypeMieru: true,
	}
	if !validTypes[req.Type] {
		return nil, fmt.Errorf("unsupported node type: %s", req.Type)
	}

	if req.ServerInfo != "" {
		var tmp map[string]interface{}
		if err := json.Unmarshal([]byte(req.ServerInfo), &tmp); err != nil {
			return nil, errors.New("server_info must be valid JSON")
		}
	}

	if req.Rate <= 0 {
		req.Rate = 1.0
	}
	show := req.Show
	if show != 0 && show != 1 {
		show = 1
	}

	node := &model.Node{
		Name:                req.Name,
		Type:                req.Type,
		Address:             req.Address,
		Port:                req.Port,
		ServerInfo:          req.ServerInfo,
		GroupID:             req.GroupID,
		Rate:                req.Rate,
		Status:              1,
		ParentID:            req.ParentID,
		Sort:                req.Sort,
		Show:                show,
		Tags:                req.Tags,
		HealthCheckPort:     req.HealthCheckPort,
		HealthCheckInterval: req.HealthCheckInterval,
		HealthCheckTimeout:  req.HealthCheckTimeout,
		HealthCheckType:     req.HealthCheckType,
	}

	if node.HealthCheckType == "" {
		node.HealthCheckType = "tcp"
	}

	if err := s.nodeRepo.Create(node); err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}
	return node, nil
}

// UpdateNodeEx 扩展更新节点（支持新字段）
func (s *nodeService) UpdateNodeEx(id uint, req UpdateNodeRequest) (*model.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found")
	}

	if req.Name != "" {
		node.Name = req.Name
	}
	if req.Type != "" {
		node.Type = req.Type
	}
	if req.Address != "" {
		node.Address = req.Address
	}
	if req.Port > 0 && req.Port <= 65535 {
		node.Port = req.Port
	}
	if req.ServerInfo != "" {
		var tmp map[string]interface{}
		if err := json.Unmarshal([]byte(req.ServerInfo), &tmp); err != nil {
			return nil, errors.New("server_info must be valid JSON")
		}
		node.ServerInfo = req.ServerInfo
	}
	if req.Rate > 0 {
		node.Rate = req.Rate
	}
	if req.Status >= 0 && req.Status <= 2 {
		node.Status = req.Status
	}
	// Sort 和 Show 用指针判断是否传入
	if req.Sort != nil {
		node.Sort = *req.Sort
	}
	if req.Show != nil {
		node.Show = *req.Show
	}
	if req.Tags != "" {
		node.Tags = req.Tags
	}
	if req.HealthCheckPort > 0 {
		node.HealthCheckPort = req.HealthCheckPort
	}
	if req.HealthCheckInterval > 0 {
		node.HealthCheckInterval = req.HealthCheckInterval
	}
	if req.HealthCheckTimeout > 0 {
		node.HealthCheckTimeout = req.HealthCheckTimeout
	}
	if req.HealthCheckType != "" {
		node.HealthCheckType = req.HealthCheckType
	}

	if err := s.nodeRepo.Update(node); err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}
	return node, nil
}

// DeleteNode 删除节点
func (s *nodeService) DeleteNode(id uint) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("node not found")
	}

	return s.nodeRepo.Delete(id)
}

// ListNodes 节点列表
func (s *nodeService) ListNodes(page, pageSize int, groupID uint) ([]model.Node, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.nodeRepo.List(page, pageSize, groupID)
}

// ListOnlineNodes 获取所有在线节点
func (s *nodeService) ListOnlineNodes() ([]model.Node, error) {
	return s.nodeRepo.ListAllOnline()
}

// UpdateNodeStatus 更新节点状态
func (s *nodeService) UpdateNodeStatus(id uint, status int) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("node not found")
	}

	node.Status = status
	return s.nodeRepo.Update(node)
}

// BatchUpdateStatus 批量更新节点状态
func (s *nodeService) BatchUpdateStatus(ids []uint, status int) error {
	if len(ids) == 0 {
		return errors.New("no node IDs provided")
	}
	return s.nodeRepo.BatchUpdateStatus(ids, status)
}

// BatchDelete 批量删除节点
func (s *nodeService) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return errors.New("no node IDs provided")
	}
	return s.nodeRepo.BatchDelete(ids)
}

// BatchMoveGroup 批量移动节点分组
func (s *nodeService) BatchMoveGroup(ids []uint, groupID uint) error {
	if len(ids) == 0 {
		return errors.New("no node IDs provided")
	}
	return s.nodeRepo.BatchMoveGroup(ids, groupID)
}

// CopyNode 复制节点
func (s *nodeService) CopyNode(id uint) (*model.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found")
	}

	// 创建节点副本
	newNode := &model.Node{
		Name:                node.Name + " (副本)",
		Type:                node.Type,
		Address:             node.Address,
		Port:                node.Port,
		ServerInfo:          node.ServerInfo,
		GroupID:             node.GroupID,
		Rate:                node.Rate,
		Status:              0, // 新节点默认离线
		ParentID:            node.ParentID,
		HealthCheckPort:     node.HealthCheckPort,
		HealthCheckInterval: node.HealthCheckInterval,
		HealthCheckTimeout:  node.HealthCheckTimeout,
		HealthCheckType:     node.HealthCheckType,
	}

	if err := s.nodeRepo.Create(newNode); err != nil {
		return nil, fmt.Errorf("failed to copy node: %w", err)
	}

	return newNode, nil
}

// UpdateNodeSort 更新节点排序
func (s *nodeService) UpdateNodeSort(id uint, sort int) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("node not found")
	}

	return s.nodeRepo.UpdateSort(id, sort)
}

// ResetNodeTraffic 重置节点流量
func (s *nodeService) ResetNodeTraffic(id uint) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("node not found")
	}

	return s.nodeRepo.ResetTraffic(id)
}

// BatchResetNodeTraffic 批量重置节点流量
func (s *nodeService) BatchResetNodeTraffic(ids []uint) error {
	if len(ids) == 0 {
		return errors.New("no node IDs provided")
	}

	return s.nodeRepo.BatchResetTraffic(ids)
}
