package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"xboard-go/config"
	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/internal/service"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// NodeServiceServer — server-side interface (mirror of the gRPC service)
// ---------------------------------------------------------------------------

// NodeServiceServer is the interface that the gRPC service implementation
// must satisfy.  Defined manually because we cannot run protoc.
type NodeServiceServer interface {
	Handshake(ctx context.Context, req *HandshakeRequest) (*HandshakeResponse, error)
	GetConfig(ctx context.Context, req *NodeConfigRequest) (*NodeConfig, error)
	GetUsers(ctx context.Context, req *UserListRequest) (*UserList, error)
	Stream(stream NodeService_StreamServer) error
}

// NodeConfigRequest is a lightweight request for GetConfig.
type NodeConfigRequest struct {
	NodeId uint32 `json:"node_id"`
}

// NodeService_StreamServer is the server-side bidirectional stream interface.
// It extends grpc.ServerStream with typed Send/Recv for our message types.
type NodeService_StreamServer interface {
	Send(*PanelMessage) error
	Recv() (*NodeMessage, error)
	grpc.ServerStream
}

// NodeServiceClient is the client-side interface.
type NodeServiceClient interface {
	Handshake(ctx context.Context, req *HandshakeRequest) (*HandshakeResponse, error)
	GetConfig(ctx context.Context, req *NodeConfigRequest) (*NodeConfig, error)
	GetUsers(ctx context.Context, req *UserListRequest) (*UserList, error)
	Stream(ctx context.Context) (NodeService_StreamClient, error)
}

// NodeService_StreamClient is the client-side bidirectional stream interface.
type NodeService_StreamClient interface {
	Send(*NodeMessage) error
	Recv() (*PanelMessage, error)
	grpc.ClientStream
}

// ---------------------------------------------------------------------------
// Stream adapter — wraps grpc.ServerStream with typed Send/Recv
// ---------------------------------------------------------------------------

// nodeStreamServer wraps a grpc.ServerStream to provide typed Send/Recv
// for PanelMessage and NodeMessage.
type nodeStreamServer struct {
	grpc.ServerStream
}

func (s *nodeStreamServer) Send(msg *PanelMessage) error {
	return s.ServerStream.SendMsg(msg)
}

func (s *nodeStreamServer) Recv() (*NodeMessage, error) {
	msg := new(NodeMessage)
	if err := s.ServerStream.RecvMsg(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// NewNodeStreamServer wraps a raw grpc.ServerStream into NodeService_StreamServer.
func NewNodeStreamServer(stream grpc.ServerStream) NodeService_StreamServer {
	return &nodeStreamServer{ServerStream: stream}
}

// ---------------------------------------------------------------------------
// nodeServiceServer — concrete implementation
// ---------------------------------------------------------------------------

type nodeServiceServer struct {
	nodeRepo    repository.NodeRepository
	trafficSvc  service.TrafficService
	uniProxySvc service.UniProxyService
	broadcaster *Broadcaster
	cfg         *config.Config
}

// NewNodeServiceServer creates a new NodeServiceServer.
func NewNodeServiceServer(
	nodeRepo repository.NodeRepository,
	trafficSvc service.TrafficService,
	uniProxySvc service.UniProxyService,
	broadcaster *Broadcaster,
	cfg *config.Config,
) NodeServiceServer {
	return &nodeServiceServer{
		nodeRepo:    nodeRepo,
		trafficSvc:  trafficSvc,
		uniProxySvc: uniProxySvc,
		broadcaster: broadcaster,
		cfg:         cfg,
	}
}

// ---------------------------------------------------------------------------
// Handshake — first call from the node to authenticate & fetch initial state
// ---------------------------------------------------------------------------

func (s *nodeServiceServer) Handshake(ctx context.Context, req *HandshakeRequest) (*HandshakeResponse, error) {
	nodeID, _ := NodeIDFromContext(ctx)

	node, err := s.nodeRepo.GetByID(uint(nodeID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query node: %v", err)
	}
	if node == nil {
		return nil, status.Errorf(codes.NotFound, "node %d not found", nodeID)
	}

	// Mark node as online.
	now := time.Now()
	node.Status = 1
	node.LastOnline = &now
	if err := s.nodeRepo.Update(node); err != nil {
		logger.Sugar().Warnf("Handshake: failed to mark node %d online: %v", nodeID, err)
	}

	// If MachineID is provided, validate and mark the machine online.
	if req.MachineID > 0 {
		machineRepo := repository.NewServerMachineRepository()
		machine, err := machineRepo.GetByID(uint(req.MachineID))
		if err == nil && machine != nil {
			if err := machineRepo.UpdateStatus(uint(req.MachineID), true); err != nil {
				logger.Sugar().Warnf("gRPC Handshake: failed to mark machine %d online: %v", req.MachineID, err)
			} else {
				logger.Sugar().Infof("gRPC Handshake: marked machine %d online", req.MachineID)
			}
		}
	}

	// Build config.
	nodeCfg := buildNodeConfig(node)

	// Fetch active users.
	users, err := getActiveUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query users: %v", err)
	}

	logger.Sugar().Infof("gRPC Handshake: node=%d (%s), version=%s, users=%d",
		nodeID, req.Hostname, req.Version, len(users))

	return &HandshakeResponse{
		Success:       true,
		Message:       "ok",
		PushInterval:  60,
		PullInterval:  60,
		TrackInterval: 10,
		Config:        nodeCfg,
		Users:         users,
	}, nil
}

// ---------------------------------------------------------------------------
// GetConfig — returns the current node configuration
// ---------------------------------------------------------------------------

func (s *nodeServiceServer) GetConfig(ctx context.Context, req *NodeConfigRequest) (*NodeConfig, error) {
	nodeID, _ := NodeIDFromContext(ctx)

	node, err := s.nodeRepo.GetByID(uint(nodeID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query node: %v", err)
	}
	if node == nil {
		return nil, status.Errorf(codes.NotFound, "node %d not found", nodeID)
	}

	return buildNodeConfig(node), nil
}

// ---------------------------------------------------------------------------
// GetUsers — returns the current active user list
// ---------------------------------------------------------------------------

func (s *nodeServiceServer) GetUsers(ctx context.Context, req *UserListRequest) (*UserList, error) {
	users, err := getActiveUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query users: %v", err)
	}
	return &UserList{Users: users}, nil
}

// ---------------------------------------------------------------------------
// Stream — bidirectional node ↔ panel stream
// ---------------------------------------------------------------------------

func (s *nodeServiceServer) Stream(stream NodeService_StreamServer) error {
	nodeID, _ := NodeIDFromContext(stream.Context())
	logger.Sugar().Infof("gRPC Stream: node %d connected", nodeID)

	// Update node status to online.
	markNodeOnline(s.nodeRepo, nodeID)

	// Subscribe to panel events for this node.
	eventCh := s.broadcaster.Subscribe(nodeID)
	defer func() {
		s.broadcaster.Unsubscribe(nodeID)
		markNodeOffline(s.nodeRepo, nodeID)
		logger.Sugar().Infof("gRPC Stream: node %d disconnected", nodeID)
	}()

	// Error channel for the receiver goroutine.
	recvErrCh := make(chan error, 1)

	// Goroutine: receive NodeMessages from the node.
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logger.Sugar().Infof("gRPC Stream: node %d sent EOF", nodeID)
				} else {
					logger.Sugar().Warnf("gRPC Stream: node %d recv error: %v", nodeID, err)
				}
				recvErrCh <- err
				return
			}
			s.handleNodeMessage(stream.Context(), nodeID, msg)
		}
	}()

	// Heartbeat ticker (send ping every 30s).
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Main loop.
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()

		case err := <-recvErrCh:
			return err

		case event, ok := <-eventCh:
			if !ok {
				return status.Error(codes.Unavailable, "event channel closed")
			}
			if err := stream.Send(event); err != nil {
				logger.Sugar().Warnf("gRPC Stream: node %d send error: %v", nodeID, err)
				return err
			}

		case <-pingTicker.C:
			ping := &PanelMessage{
				Payload: &Ping{Timestamp: time.Now().Unix()},
			}
			if err := stream.Send(ping); err != nil {
				logger.Sugar().Warnf("gRPC Stream: node %d ping send error: %v", nodeID, err)
				return err
			}
		}
	}
}

// ---------------------------------------------------------------------------
// handleNodeMessage dispatches incoming NodeMessages to the right handler.
// ---------------------------------------------------------------------------

func (s *nodeServiceServer) handleNodeMessage(ctx context.Context, nodeID uint32, msg *NodeMessage) {
	if msg == nil || msg.Payload == nil {
		return
	}
	switch p := msg.Payload.(type) {
	case *TrafficReport:
		s.handleTrafficReport(ctx, nodeID, p)
	case *AliveReport:
		s.handleAliveReport(ctx, nodeID, p)
	case *StatusReport:
		s.handleStatusReport(ctx, nodeID, p)
	case *DeviceReport:
		s.handleDeviceReport(ctx, nodeID, p)
	case *Pong:
		logger.Sugar().Debugf("pong from node %d", nodeID)
	}
}

func (s *nodeServiceServer) handleTrafficReport(ctx context.Context, nodeID uint32, report *TrafficReport) {
	if len(report.Deltas) == 0 {
		return
	}
	items := make([]service.UniProxyTrafficItem, 0, len(report.Deltas))
	for _, d := range report.Deltas {
		items = append(items, service.UniProxyTrafficItem{
			UserID:   uint(d.UserID),
			Upload:   d.Upload,
			Download: d.Download,
		})
	}
	if err := s.uniProxySvc.PushTraffic(uint(nodeID), items); err != nil {
		logger.Sugar().Errorf("gRPC TrafficReport: node=%d error=%v", nodeID, err)
	}
}

func (s *nodeServiceServer) handleAliveReport(ctx context.Context, nodeID uint32, report *AliveReport) {
	// Collect all IPs from all users in the Alive map.
	var allIPs []string
	for _, aliveIPs := range report.Alive {
		allIPs = append(allIPs, aliveIPs.IPs...)
	}
	if len(allIPs) == 0 {
		return
	}
	if err := s.uniProxySvc.ReportAlive(uint(nodeID), allIPs); err != nil {
		logger.Sugar().Errorf("gRPC AliveReport: node=%d error=%v", nodeID, err)
	}
}

func (s *nodeServiceServer) handleStatusReport(ctx context.Context, nodeID uint32, report *StatusReport) {
	logger.Sugar().Infof("gRPC StatusReport: node=%d cpu=%.1f%% mem=%d/%d disk=%d/%d uptime=%d conns=%d users=%d goroutines=%d",
		nodeID, report.CPU, report.MemUsed, report.MemTotal, report.DiskUsed, report.DiskTotal,
		report.Uptime, report.ActiveConns, report.ActiveUsers, report.Goroutines)

	// Update node's LastOnline timestamp.
	markNodeOnline(s.nodeRepo, nodeID)

	// Update metrics cache for monitoring dashboard.
	GlobalMetricsCache.UpdateMetrics(nodeID, report)

	// Update associated server_machine record if machine_id is present in the node's server_info.
	node, err := s.nodeRepo.GetByID(uint(nodeID))
	if err == nil && node != nil && node.ServerInfo != "" {
		if machineID := parseMachineIDFromServerInfo(node.ServerInfo); machineID > 0 {
			cpuPercent := float64(report.CPU)
			var memPercent, diskPercent float64
			if report.MemTotal > 0 {
				memPercent = float64(report.MemUsed) / float64(report.MemTotal) * 100
			}
			if report.DiskTotal > 0 {
				diskPercent = float64(report.DiskUsed) / float64(report.DiskTotal) * 100
			}
			db := database.Get()
			db.Model(&model.ServerMachine{}).Where("id = ?", machineID).Updates(map[string]interface{}{
				"cpu":          cpuPercent,
				"memory":       memPercent,
				"disk":         diskPercent,
				"is_active":    true,
				"last_seen_at": time.Now().Unix(),
			})
		}
	}
}

func (s *nodeServiceServer) handleDeviceReport(ctx context.Context, nodeID uint32, report *DeviceReport) {
	logger.Sugar().Debugf("gRPC DeviceReport: node=%d users=%d", nodeID, len(report.Devices))
}

// ---------------------------------------------------------------------------
// Helpers (package-level)
// ---------------------------------------------------------------------------

// getActiveUsers queries all active, non-expired users.
// Uses database.Get() directly, mirroring the UniProxyService pattern.
// Now also loads each user's latest paid plan to get speed/device limits.
func getActiveUsers() ([]*User, error) {
	db := database.Get()
	var dbUsers []model.User
	err := db.
		Where("status = ?", 1).
		Where("(expired_at IS NULL OR expired_at > ?)", time.Now()).
		Find(&dbUsers).Error
	if err != nil {
		return nil, fmt.Errorf("query active users: %w", err)
	}

	users := make([]*User, 0, len(dbUsers))
	for _, u := range dbUsers {
		speedLimit := int32(0)
		deviceLimit := int32(0)

		// 查询用户最近一笔已支付订单的套餐，获取速度和设备限制
		var latestOrder model.Order
		if err := db.Where("user_id = ? AND status = ?", u.ID, model.OrderStatusPaid).
			Order("paid_at DESC").
			Preload("Plan").
			First(&latestOrder).Error; err == nil && latestOrder.Plan.ID > 0 {
			deviceLimit = int32(latestOrder.Plan.DeviceLimit)
		}

		user := &User{
			ID:           uint32(u.ID),
			UUID:         u.SubscribeToken,
			Email:        u.Email,
			Passwd:       u.SubscribeToken,
			SpeedLimit:   speedLimit,
			DeviceLimit:  deviceLimit,
			TrafficLimit: u.TrafficLimit,
			UsedTraffic:  u.UsedTraffic,
			Status:       int32(u.Status),
		}
		if u.ExpiredAt != nil {
			user.ExpiredAt = u.ExpiredAt.Unix()
		}
		users = append(users, user)
	}
	return users, nil
}

// markNodeOnline sets the node's status to online.
func markNodeOnline(nodeRepo repository.NodeRepository, nodeID uint32) {
	node, err := nodeRepo.GetByID(uint(nodeID))
	if err != nil || node == nil {
		return
	}
	now := time.Now()
	node.Status = 1
	node.LastOnline = &now
	if err := nodeRepo.Update(node); err != nil {
		logger.Sugar().Warnf("markNodeOnline: failed to update node %d: %v", nodeID, err)
	}
}

// markNodeOffline sets the node's status to offline.
func markNodeOffline(nodeRepo repository.NodeRepository, nodeID uint32) {
	node, err := nodeRepo.GetByID(uint(nodeID))
	if err != nil || node == nil {
		return
	}
	node.Status = 0
	if err := nodeRepo.Update(node); err != nil {
		logger.Sugar().Warnf("markNodeOffline: failed to update node %d: %v", nodeID, err)
	}
}

// GetActiveUsersForBroadcast queries all active, non-expired users and
// returns them as proto-style User structs for broadcasting to nodes.
// This is called from cmd/server/main.go to avoid circular imports.
func GetActiveUsersForBroadcast() ([]*User, error) {
	return getActiveUsers()
}

// buildNodeConfig converts a model.Node to a proto-style NodeConfig.
func buildNodeConfig(node *model.Node) *NodeConfig {
	groupIDs := make([]uint32, 0, len(node.GetGroupIDList()))
	for _, id := range node.GetGroupIDList() {
		groupIDs = append(groupIDs, uint32(id))
	}
	return &NodeConfig{
		ID:         uint32(node.ID),
		Name:       node.Name,
		Protocol:   node.Type,
		Address:    node.Host,
		Port:       int32(node.Port),
		ServerInfo: node.ServerInfo,
		Rate:       float32(node.Rate),
		GroupIDs:   groupIDs,
		ParentID:   uint32(node.ParentID),
	}
}

// parseMachineIDFromServerInfo extracts machine_id from the node's server_info JSON.
// Expected format: {"machine_id": 1, ...}
func parseMachineIDFromServerInfo(serverInfo string) uint {
	if serverInfo == "" {
		return 0
	}
	var info struct {
		MachineID uint `json:"machine_id"`
	}
	if err := json.Unmarshal([]byte(serverInfo), &info); err != nil {
		return 0
	}
	return info.MachineID
}
