package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"xboard-go/config"
	"xboard-go/internal/grpc"
	"xboard-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSMessage 是 Xboard-Node 期望的 WebSocket 消息格式。
type WSMessage struct {
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

// NodeWSHandler 处理 Xboard-Node 的 WebSocket 连接。
type NodeWSHandler struct {
	upgrader websocket.Upgrader
	bcast    *grpc.Broadcaster
}

// NewNodeWSHandler 创建 NodeWSHandler。
func NewNodeWSHandler(bcast *grpc.Broadcaster) *NodeWSHandler {
	return &NodeWSHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源（节点连接）
			},
			ReadBufferSize:  1024 * 64,
			WriteBufferSize: 1024 * 64,
		},
		bcast: bcast,
	}
}

// Handle 处理 WebSocket 连接升级和消息循环。
// 路由: GET /api/v2/server/ws?token=xxx&node_id=xxx 或 ?token=xxx&machine_id=xxx
func (h *NodeWSHandler) Handle(c *gin.Context) {
	// 认证（从 query params）
	expectedKey := config.Get().App.NodeAPIKey
	token := c.Query("token")
	nodeIDStr := c.Query("node_id")
	machineIDStr := c.Query("machine_id")

	if token == "" || token != expectedKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 支持 node_id 或 machine_id
	var nodeID uint32
	if nodeIDStr != "" {
		nodeID64, err := strconv.ParseUint(nodeIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node_id"})
			return
		}
		nodeID = uint32(nodeID64)
	} else if machineIDStr != "" {
		// machine mode: machine_id 用于认证，node_id 在后续消息中传递
		// 这里先用 machine_id 作为标识
		machineID64, err := strconv.ParseUint(machineIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid machine_id"})
			return
		}
		_ = machineID64 // machine mode uses per-message node_id
		// 对于 machine mode，node_id 在后续消息中动态设置
		nodeID = 0 // 标识为 machine mode
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing node_id or machine_id"})
		return
	}

	// 升级为 WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Sugar().Errorf("WS upgrade failed for node %d: %v", nodeID, err)
		return
	}
	defer conn.Close()

	logger.Sugar().Infof("WS: node %d connected", nodeID)

	// 发送 auth.success
	authMsg := WSMessage{
		Event:     "auth.success",
		Timestamp: time.Now().Unix(),
	}
	if err := conn.WriteJSON(authMsg); err != nil {
		logger.Sugar().Warnf("WS: failed to send auth.success to node %d: %v", nodeID, err)
		return
	}

	// 订阅广播事件（复用 gRPC Broadcaster）
	eventCh := h.bcast.Subscribe(nodeID)
	defer h.bcast.Unsubscribe(nodeID)

	// 标记节点在线
	markNodeOnlineREST(nodeID)

	// 读写循环
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// 错误通道
	errCh := make(chan error, 1)

	// 读取 goroutine（处理客户端消息）
	go func() {
		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			var msg WSMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				continue
			}
			h.handleClientMessage(nodeID, msg)
		}
	}()

	// Ping ticker（保持连接活跃）
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// 主循环
	for {
		select {
		case <-ctx.Done():
			return

		case err := <-errCh:
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logger.Sugar().Infof("WS: node %d disconnected normally", nodeID)
			} else {
				logger.Sugar().Warnf("WS: node %d read error: %v", nodeID, err)
			}
			return

		case event, ok := <-eventCh:
			if !ok {
				return
			}
			// 将 gRPC PanelMessage 转换为 WS 消息格式
			wsMsg, err := convertPanelMessageToWS(event)
			if err != nil {
				logger.Sugar().Warnf("WS: failed to convert message for node %d: %v", nodeID, err)
				continue
			}
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteJSON(wsMsg); err != nil {
				logger.Sugar().Warnf("WS: node %d write error: %v", nodeID, err)
				return
			}

		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteJSON(WSMessage{
				Event:     "ping",
				Timestamp: time.Now().Unix(),
			}); err != nil {
				logger.Sugar().Warnf("WS: node %d ping error: %v", nodeID, err)
				return
			}
		}
	}
}

// handleClientMessage 处理从节点收到的 WebSocket 消息。
func (h *NodeWSHandler) handleClientMessage(nodeID uint32, msg WSMessage) {
	switch msg.Event {
	case "pong":
		// 心跳响应，忽略
	case "node.status":
		// 节点状态更新 — 解析并记录系统指标
		var statusData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &statusData); err == nil {
			// machine mode: 从消息中提取 node_id
			actualNodeID := nodeID
			if nid, ok := statusData["node_id"].(float64); ok && nid > 0 {
				actualNodeID = uint32(nid)
			}
			if actualNodeID > 0 {
				processNodeStatus(actualNodeID, statusData)
				markNodeOnlineREST(actualNodeID)
			}
		}
	case "report.devices":
		// 设备上报 — 解析设备数据
		var deviceData map[string][]string
		if err := json.Unmarshal(msg.Data, &deviceData); err == nil {
			logger.Sugar().Debugf("WS: node %d reported devices for %d users", nodeID, len(deviceData))
			// TODO: 存储设备状态供设备限制使用
		}
	default:
		logger.Sugar().Debugf("WS: node %d sent unknown event: %s", nodeID, msg.Event)
	}
}

// convertPanelMessageToWS 将 gRPC 的 PanelMessage 转换为 Xboard-Node 期望的 WS 消息格式。
func convertPanelMessageToWS(msg *grpc.PanelMessage) (WSMessage, error) {
	if msg == nil || msg.Payload == nil {
		return WSMessage{}, fmt.Errorf("nil message")
	}

	var event string
	var data interface{}

	switch p := msg.Payload.(type) {
	case *grpc.ConfigUpdate:
		event = "sync.config"
		// 将 gRPC NodeConfig 转换为 REST NodeConfig 格式
		restCfg := convertGRPCConfigToREST(p.Config)
		data = map[string]interface{}{
			"config":    restCfg,
			"timestamp": time.Now().Unix(),
		}

	case *grpc.UserListUpdate:
		event = "sync.users"
		users := convertGRPCUsersToREST(p.Users)
		data = map[string]interface{}{
			"users":     users,
			"timestamp": time.Now().Unix(),
		}

	case *grpc.UserDelta:
		event = "sync.user.delta"
		action := "add"
		if len(p.Removed) > 0 {
			action = "remove"
		}
		users := convertGRPCUsersToREST(p.Added)
		data = map[string]interface{}{
			"action":    action,
			"users":     users,
			"timestamp": time.Now().Unix(),
		}

	case *grpc.Ping:
		return WSMessage{
			Event:     "ping",
			Timestamp: p.Timestamp,
		}, nil

	default:
		return WSMessage{}, fmt.Errorf("unsupported payload type: %T", msg.Payload)
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return WSMessage{}, err
	}

	return WSMessage{
		Event:     event,
		Data:      dataBytes,
		Timestamp: time.Now().Unix(),
	}, nil
}

// convertGRPCConfigToREST 将 gRPC NodeConfig 转换为 Xboard-Node 期望的 REST 格式。
func convertGRPCConfigToREST(cfg *grpc.NodeConfig) map[string]interface{} {
	if cfg == nil {
		return nil
	}
	result := map[string]interface{}{
		"protocol":   cfg.Protocol,
		"listen_ip":  "::",
		"server_port": cfg.Port,
		"routes":     []interface{}{},
		"base_config": map[string]int{
			"push_interval": 60,
			"pull_interval": 60,
		},
	}

	// 解析 ServerInfo 获取详细配置
	if cfg.ServerInfo != "" {
		var info map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.ServerInfo), &info); err == nil {
			// 传输协议
			if v, ok := info["network"].(string); ok {
				result["network"] = v
			}
			if v, ok := info["network_type"].(string); ok {
				if _, exists := result["network"]; !exists {
					result["network"] = v
				}
			}

			// 网络设置
			netSettings := make(map[string]interface{})
			if v, ok := info["ws_path"].(string); ok {
				netSettings["path"] = v
			}
			if v, ok := info["ws_host"].(string); ok {
				netSettings["headers"] = map[string]string{"Host": v}
			}
			if v, ok := info["grpc_service_name"].(string); ok {
				netSettings["serviceName"] = v
			}
			if v, ok := info["h2_path"].(string); ok {
				netSettings["path"] = v
			}
			if v, ok := info["h2_host"].(string); ok {
				netSettings["host"] = []string{v}
			}
			if len(netSettings) > 0 {
				result["networkSettings"] = netSettings
			}

			// TLS
			if v, ok := info["tls"].(bool); ok && v {
				result["tls"] = 1
			}
			if v, ok := info["tls"].(float64); ok {
				result["tls"] = int(v)
			}
			if v, ok := info["tls_settings"].(map[string]interface{}); ok {
				result["tls_settings"] = v
			}
			if v, ok := info["tls_server_name"].(string); ok {
				if _, exists := result["tls_settings"]; !exists {
					result["tls_settings"] = map[string]interface{}{"server_name": v}
				}
			}

			// 协议特定字段
			copyStringFields := []string{
				"cipher", "plugin", "plugin_opts", "server_key",
				"flow", "decryption", "host", "server_name",
				"obfs", "obfs-password", "congestion_control",
				"transport", "traffic_pattern",
			}
			for _, field := range copyStringFields {
				if v, ok := info[field]; ok {
					result[field] = v
				}
			}

			copyIntFields := []string{"up_mbps", "down_mbps", "version"}
			for _, field := range copyIntFields {
				if v, ok := info[field].(float64); ok {
					result[field] = int(v)
				}
			}

			if v, ok := info["accept_proxy_protocol"].(bool); ok {
				result["accept_proxy_protocol"] = v
			}

			// Xboard 扩展字段
			extensionFields := []string{
				"kernel_type", "kernel_log_level",
				"custom_outbounds", "custom_routes", "custom_route_rules",
				"cert_config", "multiplex", "padding_scheme",
			}
			for _, field := range extensionFields {
				if v, ok := info[field]; ok {
					result[field] = v
				}
			}
			if v, ok := info["auto_tls"].(bool); ok {
				result["auto_tls"] = v
			}
			if v, ok := info["domain"].(string); ok {
				result["domain"] = v
			}

			// 路由规则
			if v, ok := info["routes"].([]interface{}); ok && len(v) > 0 {
				result["routes"] = v
			}
		}
	}

	return result
}

// convertGRPCUsersToREST 将 gRPC User 列表转换为 Xboard-Node 期望的格式。
func convertGRPCUsersToREST(users []*grpc.User) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		result = append(result, map[string]interface{}{
			"id":           u.ID,
			"uuid":         u.UUID,
			"speed_limit":  u.SpeedLimit,
			"device_limit": u.DeviceLimit,
		})
	}
	return result
}

// WSNodeBroadcaster 是一个扩展的 Broadcaster，同时支持 gRPC 和 WebSocket 节点。
// 它包装了原有的 gRPC Broadcaster，并额外管理 WS 连接。
type WSNodeBroadcaster struct {
	grpc   *grpc.Broadcaster
	wsMu   sync.RWMutex
	wsConns map[uint32][]*wsConn // nodeID -> WS connections
}

// NewWSNodeBroadcaster 创建 WSNodeBroadcaster。
func NewWSNodeBroadcaster(grpcBcast *grpc.Broadcaster) *WSNodeBroadcaster {
	return &WSNodeBroadcaster{
		grpc:   grpcBcast,
		wsConns: make(map[uint32][]*wsConn),
	}
}

type wsConn struct {
	conn    *websocket.Conn
	sendCh  chan WSMessage
	nodeID  uint32
}

// BroadcastConfig 向节点发送配置更新（同时支持 gRPC 和 WS）。
func (b *WSNodeBroadcaster) BroadcastConfig(nodeID uint32, config *grpc.NodeConfig) {
	// 发送给 gRPC 节点
	b.grpc.BroadcastConfig(nodeID, config)

	// 发送给 WS 节点
	restCfg := convertGRPCConfigToREST(config)
	data, _ := json.Marshal(map[string]interface{}{
		"config":    restCfg,
		"timestamp": time.Now().Unix(),
	})
	b.broadcastToWS(nodeID, WSMessage{
		Event:     "sync.config",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// BroadcastUsers 向节点发送用户列表更新（同时支持 gRPC 和 WS）。
func (b *WSNodeBroadcaster) BroadcastUsers(nodeID uint32, users []*grpc.User) {
	// 发送给 gRPC 节点
	b.grpc.BroadcastUsers(nodeID, users)

	// 发送给 WS 节点
	restUsers := convertGRPCUsersToREST(users)
	data, _ := json.Marshal(map[string]interface{}{
		"users":     restUsers,
		"timestamp": time.Now().Unix(),
	})
	b.broadcastToWS(nodeID, WSMessage{
		Event:     "sync.users",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// broadcastToWS 向指定节点的所有 WS 连接广播消息。
func (b *WSNodeBroadcaster) broadcastToWS(nodeID uint32, msg WSMessage) {
	b.wsMu.RLock()
	conns := b.wsConns[nodeID]
	b.wsMu.RUnlock()

	for _, wc := range conns {
		select {
		case wc.sendCh <- msg:
		default:
			logger.Sugar().Warnf("WS broadcast: channel full for node %d, dropping", nodeID)
		}
	}
}

// RegisterWSConn 注册一个 WS 连接。
func (b *WSNodeBroadcaster) RegisterWSConn(nodeID uint32, conn *websocket.Conn) *wsConn {
	wc := &wsConn{
		conn:   conn,
		sendCh: make(chan WSMessage, 16),
		nodeID: nodeID,
	}
	b.wsMu.Lock()
	b.wsConns[nodeID] = append(b.wsConns[nodeID], wc)
	b.wsMu.Unlock()
	return wc
}

// UnregisterWSConn 注销一个 WS 连接。
func (b *WSNodeBroadcaster) UnregisterWSConn(nodeID uint32, wc *wsConn) {
	b.wsMu.Lock()
	conns := b.wsConns[nodeID]
	for i, c := range conns {
		if c == wc {
			b.wsConns[nodeID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(b.wsConns[nodeID]) == 0 {
		delete(b.wsConns, nodeID)
	}
	b.wsMu.Unlock()
}
