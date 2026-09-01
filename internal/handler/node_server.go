package handler

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"xboard-go/config"
	"xboard-go/internal/middleware"
	"xboard-go/internal/model"
	"xboard-go/internal/service"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// NodeServerHandler 处理 Xboard-Node 的 REST API 请求。
// 实现 Xboard-Node 期望的端点格式：
//   - POST /api/v2/server/handshake
//   - GET  /api/v1/server/UniProxy/config
//   - GET  /api/v1/server/UniProxy/user
//   - POST /api/v1/server/UniProxy/push
//   - POST /api/v1/server/UniProxy/alive
//   - POST /api/v1/server/UniProxy/status
//   - POST /api/v2/server/report
type NodeServerHandler struct {
	uniProxySvc service.UniProxyService
}

// NewNodeServerHandler 创建 NodeServerHandler。
func NewNodeServerHandler(uniProxySvc service.UniProxyService) *NodeServerHandler {
	return &NodeServerHandler{uniProxySvc: uniProxySvc}
}

// ---------------------------------------------------------------------------
// POST /api/v2/server/handshake
// ---------------------------------------------------------------------------

// Handshake 处理节点握手请求。
func (h *NodeServerHandler) Handshake(c *gin.Context) {
	nodeID := middleware.GetNodeID(c)
	machineID := middleware.GetMachineID(c)

	// 标记节点在线
	if nodeID > 0 {
		if _, err := markNodeOnlineREST(nodeID); err != nil {
			logger.Sugar().Warnf("REST Handshake: failed to mark node %d online: %v", nodeID, err)
		}
	}

	// Machine mode: 如果是 machine 握手，验证 machine 存在
	if machineID > 0 {
		db := database.Get()
		var machine model.ServerMachine
		if err := db.First(&machine, machineID).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "machine not found"})
			return
		}
		// 标记 machine 在线
		db.Model(&machine).Updates(map[string]interface{}{
			"is_active":   true,
			"last_seen_at": time.Now().Unix(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"websocket": gin.H{
			"enabled": true,
			"ws_url":  buildWSURL(c),
		},
		"settings": gin.H{
			"push_interval": 60,
			"pull_interval": 60,
		},
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/server/UniProxy/config  (legacy)
// GET /api/v2/server/config           (v2)
// ---------------------------------------------------------------------------

// GetConfig 返回节点配置。
func (h *NodeServerHandler) GetConfig(c *gin.Context) {
	nodeID := middleware.GetNodeID(c)

	db := database.Get()
	var node model.Node
	if err := db.First(&node, nodeID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "node not found"})
		return
	}

	nodeCfg := buildRESTNodeConfig(&node)

	// ETag 支持
	etag := computeETag(nodeCfg)
	if match := c.GetHeader("If-None-Match"); match == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.Header("ETag", etag)

	c.JSON(http.StatusOK, nodeCfg)
}

// ---------------------------------------------------------------------------
// GET /api/v1/server/UniProxy/user  (legacy)
// GET /api/v2/server/user           (v2)
// ---------------------------------------------------------------------------

// GetUsers 返回用户列表。
func (h *NodeServerHandler) GetUsers(c *gin.Context) {
	db := database.Get()
	var dbUsers []model.User
	err := db.
		Where("status = ?", 1).
		Where("(expired_at IS NULL OR expired_at > ?)", time.Now()).
		Find(&dbUsers).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to query users"})
		return
	}

	users := make([]gin.H, 0, len(dbUsers))
	for _, u := range dbUsers {
		deviceLimit := 0
		var latestOrder model.Order
		if err := db.Where("user_id = ? AND status = ?", u.ID, model.OrderStatusPaid).
			Order("paid_at DESC").
			Preload("Plan").
			First(&latestOrder).Error; err == nil && latestOrder.Plan.ID > 0 {
			deviceLimit = latestOrder.Plan.DeviceLimit
		}

		users = append(users, gin.H{
			"id":           u.ID,
			"uuid":         u.SubscribeToken,
			"speed_limit":  0,
			"device_limit": deviceLimit,
		})
	}

	// ETag 支持
	etag := computeUsersETag(users)
	if match := c.GetHeader("If-None-Match"); match == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.Header("ETag", etag)

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// ---------------------------------------------------------------------------
// POST /api/v1/server/UniProxy/push  (legacy 流量上报)
// POST /api/v2/server/report         (v2 综合上报)
// ---------------------------------------------------------------------------

// Report 处理节点综合上报（流量 + 状态 + 在线IP）。
func (h *NodeServerHandler) Report(c *gin.Context) {
	nodeID := middleware.GetNodeID(c)
	body := middleware.GetNodeBody(c)

	// 处理流量上报
	if traffic, ok := body["traffic"].(map[string]interface{}); ok && len(traffic) > 0 {
		items := make([]service.UniProxyTrafficItem, 0, len(traffic))
		for uidStr, v := range traffic {
			uid, err := strconv.ParseUint(uidStr, 10, 32)
			if err != nil {
				continue
			}
			if arr, ok := v.([]interface{}); ok && len(arr) >= 2 {
				up, _ := arr[0].(float64)
				down, _ := arr[1].(float64)
				items = append(items, service.UniProxyTrafficItem{
					UserID:   uint(uid),
					Upload:   int64(up),
					Download: int64(down),
				})
			}
		}
		if len(items) > 0 {
			if err := h.uniProxySvc.PushTraffic(uint(nodeID), items); err != nil {
				logger.Sugar().Errorf("REST Report: traffic error node=%d: %v", nodeID, err)
			}
		}
	}

	// 处理在线IP上报
	if alive, ok := body["alive"].(map[string]interface{}); ok && len(alive) > 0 {
		var allIPs []string
		for _, ips := range alive {
			if arr, ok := ips.([]interface{}); ok {
				for _, ip := range arr {
					if s, ok := ip.(string); ok {
						allIPs = append(allIPs, s)
					}
				}
			}
		}
		if len(allIPs) > 0 {
			if err := h.uniProxySvc.ReportAlive(uint(nodeID), allIPs); err != nil {
				logger.Sugar().Errorf("REST Report: alive error node=%d: %v", nodeID, err)
			}
		}
	}

	// 处理系统状态上报（status 字段）
	if status, ok := body["status"].(map[string]interface{}); ok {
		processNodeStatus(nodeID, status)
	}

	// 处理在线用户数（online 字段）
	if online, ok := body["online"].(map[string]interface{}); ok {
		processNodeOnline(nodeID, online)
	}

	// 更新节点在线状态
	markNodeOnlineREST(nodeID)

	c.JSON(http.StatusOK, gin.H{})
}

// PushTraffic 处理 legacy 流量推送。
func (h *NodeServerHandler) PushTraffic(c *gin.Context) {
	nodeID := middleware.GetNodeID(c)
	body := middleware.GetNodeBody(c)

	items := make([]service.UniProxyTrafficItem, 0)
	for uidStr, v := range body {
		if uidStr == "token" || uidStr == "node_id" || uidStr == "node_type" {
			continue
		}
		uid, err := strconv.ParseUint(uidStr, 10, 32)
		if err != nil {
			continue
		}
		if arr, ok := v.([]interface{}); ok && len(arr) >= 2 {
			up, _ := arr[0].(float64)
			down, _ := arr[1].(float64)
			items = append(items, service.UniProxyTrafficItem{
				UserID:   uint(uid),
				Upload:   int64(up),
				Download: int64(down),
			})
		}
	}

	if len(items) > 0 {
		if err := h.uniProxySvc.PushTraffic(uint(nodeID), items); err != nil {
			logger.Sugar().Errorf("REST PushTraffic: error node=%d: %v", nodeID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{})
}

// PushAlive 处理 legacy 在线IP推送。
func (h *NodeServerHandler) PushAlive(c *gin.Context) {
	nodeID := middleware.GetNodeID(c)
	body := middleware.GetNodeBody(c)

	var allIPs []string
	for uidStr, v := range body {
		if uidStr == "token" || uidStr == "node_id" || uidStr == "node_type" {
			continue
		}
		_ = uidStr
		if arr, ok := v.([]interface{}); ok {
			for _, ip := range arr {
				if s, ok := ip.(string); ok {
					allIPs = append(allIPs, s)
				}
			}
		}
	}

	if len(allIPs) > 0 {
		if err := h.uniProxySvc.ReportAlive(uint(nodeID), allIPs); err != nil {
			logger.Sugar().Errorf("REST PushAlive: error node=%d: %v", nodeID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{})
}

// PushStatus 处理 legacy 状态推送。
func (h *NodeServerHandler) PushStatus(c *gin.Context) {
	nodeID := middleware.GetNodeID(c)
	markNodeOnlineREST(nodeID)
	c.JSON(http.StatusOK, gin.H{})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// restNodeConfig 是 Xboard-Node 期望的配置格式。
type restNodeConfig struct {
	Protocol        string                 `json:"protocol"`
	ListenIP        string                 `json:"listen_ip"`
	ServerPort      int                    `json:"server_port"`
	Network         string                 `json:"network,omitempty"`
	NetworkSettings map[string]interface{} `json:"networkSettings,omitempty"`
	TLS             int                    `json:"tls,omitempty"`
	TLSSettings     map[string]interface{} `json:"tls_settings,omitempty"`
	Host            string                 `json:"host,omitempty"`
	ServerName      string                 `json:"server_name,omitempty"`
	Cipher          string                 `json:"cipher,omitempty"`
	Plugin          string                 `json:"plugin,omitempty"`
	PluginOpt       string                 `json:"plugin_opts,omitempty"`
	ServerKey       string                 `json:"server_key,omitempty"`
	Flow            string                 `json:"flow,omitempty"`
	Decryption      string                 `json:"decryption,omitempty"`
	Version         int                    `json:"version,omitempty"`
	UpMbps          int                    `json:"up_mbps,omitempty"`
	DownMbps        int                    `json:"down_mbps,omitempty"`
	Obfs            string                 `json:"obfs,omitempty"`
	ObfsPassword    string                 `json:"obfs-password,omitempty"`
	CongestionControl string              `json:"congestion_control,omitempty"`
	Routes          []interface{}          `json:"routes"`
	AcceptProxyProtocol bool               `json:"accept_proxy_protocol,omitempty"`
	BaseConfig      map[string]int         `json:"base_config"`

	// Xboard 扩展字段
	KernelType       string                   `json:"kernel_type,omitempty"`
	KernelLogLevel   string                   `json:"kernel_log_level,omitempty"`
	CustomOutbounds  []map[string]interface{} `json:"custom_outbounds,omitempty"`
	CustomRoutes     []map[string]interface{} `json:"custom_routes,omitempty"`
	CustomRouteRules []map[string]interface{} `json:"custom_route_rules,omitempty"`
	CertConfig       map[string]interface{}   `json:"cert_config,omitempty"`
	AutoTLS          bool                     `json:"auto_tls,omitempty"`
	Domain           string                   `json:"domain,omitempty"`
	Multiplex        map[string]interface{}   `json:"multiplex,omitempty"`
	PaddingScheme    interface{}              `json:"padding_scheme,omitempty"`
	Transport        string                   `json:"transport,omitempty"`
	TrafficPattern   string                   `json:"traffic_pattern,omitempty"`
}

// buildRESTNodeConfig 将面板的 Node 模型转换为 Xboard-Node 期望的配置格式。
func buildRESTNodeConfig(node *model.Node) *restNodeConfig {
	info := node.ParseServerInfo()
	siCfg := node.ParseServerInfoConfig()

	cfg := &restNodeConfig{
		Protocol:   node.Type,
		ListenIP:   "::",
		ServerPort: node.Port,
		Routes:     []interface{}{},
		BaseConfig: map[string]int{
			"push_interval": 60,
			"pull_interval": 60,
		},
	}

	// 传输协议
	if siCfg.Network != "" {
		cfg.Network = siCfg.Network
	}

	// 网络设置
	netSettings := make(map[string]interface{})
	if siCfg.WSPath != "" {
		netSettings["path"] = siCfg.WSPath
		if siCfg.WSHost != "" {
			netSettings["headers"] = map[string]string{"Host": siCfg.WSHost}
		}
	}
	if siCfg.GrpcServiceName != "" {
		netSettings["serviceName"] = siCfg.GrpcServiceName
	}
	if siCfg.H2Path != "" {
		netSettings["path"] = siCfg.H2Path
		if siCfg.H2Host != "" {
			netSettings["host"] = []string{siCfg.H2Host}
		}
	}
	if len(netSettings) > 0 {
		cfg.NetworkSettings = netSettings
	}

	// TLS
	if siCfg.TLS {
		cfg.TLS = 1
		if siCfg.TLSServerName != "" {
			cfg.TLSSettings = map[string]interface{}{
				"server_name": siCfg.TLSServerName,
			}
		}
		if siCfg.AllowInsecure {
			if cfg.TLSSettings == nil {
				cfg.TLSSettings = make(map[string]interface{})
			}
			cfg.TLSSettings["allow_insecure"] = true
		}
		if siCfg.ALPN != "" {
			if cfg.TLSSettings == nil {
				cfg.TLSSettings = make(map[string]interface{})
			}
			cfg.TLSSettings["alpn"] = siCfg.ALPN
		}
	}

	// 从 ServerInfo 中读取 TLS 相关字段（panel 可能直接存储这些字段）
	if tlsSettings, ok := info["tls_settings"].(map[string]interface{}); ok {
		cfg.TLSSettings = tlsSettings
	}
	if tlsVal, ok := info["tls"].(float64); ok {
		cfg.TLS = int(tlsVal)
	}

	// Shadowsocks
	if siCfg.Cipher != "" {
		cfg.Cipher = siCfg.Cipher
	}
	if v, ok := info["cipher"].(string); ok && cfg.Cipher == "" {
		cfg.Cipher = v
	}
	if siCfg.Password != "" {
		cfg.Cipher = siCfg.Cipher // password is per-user, not in config
	}
	if v, ok := info["plugin"].(string); ok {
		cfg.Plugin = v
	}
	if v, ok := info["plugin_opts"].(string); ok {
		cfg.PluginOpt = v
	}
	if v, ok := info["server_key"].(string); ok {
		cfg.ServerKey = v
	}

	// VMess / VLESS
	if v, ok := info["flow"].(string); ok {
		cfg.Flow = v
	}
	if v, ok := info["decryption"].(string); ok {
		cfg.Decryption = v
	}

	// Trojan
	if v, ok := info["host"].(string); ok {
		cfg.Host = v
	}
	if v, ok := info["server_name"].(string); ok {
		cfg.ServerName = v
	}

	// Hysteria / Hysteria2
	if siCfg.UpMbps > 0 {
		cfg.UpMbps = siCfg.UpMbps
	} else if v, ok := info["up_mbps"].(float64); ok {
		cfg.UpMbps = int(v)
	}
	if siCfg.DownMbps > 0 {
		cfg.DownMbps = siCfg.DownMbps
	} else if v, ok := info["down_mbps"].(float64); ok {
		cfg.DownMbps = int(v)
	}
	if siCfg.Obfs != "" {
		cfg.ObfsPassword = siCfg.Obfs
	}
	if v, ok := info["obfs"].(string); ok {
		cfg.Obfs = v
	}
	if v, ok := info["obfs-password"].(string); ok {
		cfg.ObfsPassword = v
	}
	if v, ok := info["version"].(float64); ok {
		cfg.Version = int(v)
	}

	// TUIC
	if siCfg.CongestionControl != "" {
		cfg.CongestionControl = siCfg.CongestionControl
	} else if v, ok := info["congestion_control"].(string); ok {
		cfg.CongestionControl = v
	}

	// Proxy Protocol
	if v, ok := info["accept_proxy_protocol"].(bool); ok {
		cfg.AcceptProxyProtocol = v
	}

	// Xboard 扩展字段
	if v, ok := info["kernel_type"].(string); ok {
		cfg.KernelType = v
	}
	if v, ok := info["kernel_log_level"].(string); ok {
		cfg.KernelLogLevel = v
	}
	if v, ok := info["custom_outbounds"].([]interface{}); ok {
		outbounds := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				outbounds = append(outbounds, m)
			}
		}
		cfg.CustomOutbounds = outbounds
	}
	if v, ok := info["custom_routes"].([]interface{}); ok {
		routes := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				routes = append(routes, m)
			}
		}
		cfg.CustomRoutes = routes
	}
	if v, ok := info["custom_route_rules"].([]interface{}); ok {
		rules := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				rules = append(rules, m)
			}
		}
		cfg.CustomRouteRules = rules
	}
	if v, ok := info["cert_config"].(map[string]interface{}); ok {
		cfg.CertConfig = v
	}
	if v, ok := info["auto_tls"].(bool); ok {
		cfg.AutoTLS = v
	}
	if v, ok := info["domain"].(string); ok {
		cfg.Domain = v
	}
	if v, ok := info["multiplex"].(map[string]interface{}); ok {
		cfg.Multiplex = v
	}
	if v, ok := info["padding_scheme"]; ok {
		cfg.PaddingScheme = v
	}
	if v, ok := info["transport"].(string); ok {
		cfg.Transport = v
	}
	if v, ok := info["traffic_pattern"].(string); ok {
		cfg.TrafficPattern = v
	}

	// 路由规则
	if v, ok := info["routes"].([]interface{}); ok && len(v) > 0 {
		cfg.Routes = v
	}

	return cfg
}

// buildWSURL 构建 WebSocket URL。
// WS 端点在 HTTP 服务器上（与面板同端口），使用请求中的 Host。
func buildWSURL(c *gin.Context) string {
	// 从请求中推断 scheme
	scheme := "ws"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "wss"
	}
	// 直接使用请求的 Host（已包含端口），WS 端点在同一 HTTP 服务器上
	host := c.Request.Host
	if host == "" {
		host = config.Get().Server.Addr()
	}
	return fmt.Sprintf("%s://%s/api/v2/server/ws", scheme, host)
}

func netSplitHostPort(hostport string) (host, port string, err error) {
	for i := len(hostport) - 1; i >= 0; i-- {
		if hostport[i] == ':' {
			return hostport[:i], hostport[i+1:], nil
		}
	}
	return hostport, "", nil
}

func netHostOnly(hostport string) string {
	host, _, _ := netSplitHostPort(hostport)
	return host
}

// computeETag 为配置生成 ETag。
func computeETag(cfg interface{}) string {
	data, _ := json.Marshal(cfg)
	return fmt.Sprintf(`"%x"`, md5.Sum(data))
}

// computeUsersETag 为用户列表生成 ETag。
func computeUsersETag(users []gin.H) string {
	data, _ := json.Marshal(users)
	return fmt.Sprintf(`"%x"`, md5.Sum(data))
}

// processNodeStatus 处理节点系统状态上报（CPU/内存/磁盘）。
func processNodeStatus(nodeID uint32, status map[string]interface{}) {
	cpu, _ := status["cpu"].(float64)

	var memTotal, memUsed uint64
	if mem, ok := status["mem"].(map[string]interface{}); ok {
		memTotal, _ = toUint64(mem["total"])
		memUsed, _ = toUint64(mem["used"])
	}

	var diskTotal, diskUsed uint64
	if disk, ok := status["disk"].(map[string]interface{}); ok {
		diskTotal, _ = toUint64(disk["total"])
		diskUsed, _ = toUint64(disk["used"])
	}

	logger.Sugar().Infof("REST StatusReport: node=%d cpu=%.1f%% mem=%d/%d disk=%d/%d",
		nodeID, cpu, memUsed, memTotal, diskUsed, diskTotal)

	// 更新关联的 server_machine 状态
	db := database.Get()
	var node model.Node
	if err := db.First(&node, nodeID).Error; err == nil && node.ServerInfo != "" {
		info := node.ParseServerInfo()
		if machineID, ok := info["machine_id"].(float64); ok && machineID > 0 {
			var memPercent, diskPercent float64
			if memTotal > 0 {
				memPercent = float64(memUsed) / float64(memTotal) * 100
			}
			if diskTotal > 0 {
				diskPercent = float64(diskUsed) / float64(diskTotal) * 100
			}
			db.Model(&model.ServerMachine{}).Where("id = ?", uint(machineID)).Updates(map[string]interface{}{
				"cpu":           cpu,
				"memory":        memPercent,
				"disk":          diskPercent,
				"is_active":     true,
				"last_seen_at":  time.Now().Unix(),
			})
		}
	}
}

// processNodeOnline 处理节点在线用户数上报。
func processNodeOnline(nodeID uint32, online map[string]interface{}) {
	count := len(online)
	db := database.Get()
	db.Model(&model.Node{}).Where("id = ?", nodeID).Update("online_user_count", count)
}

// toUint64 将 interface{} 转换为 uint64。
func toUint64(v interface{}) (uint64, bool) {
	switch val := v.(type) {
	case float64:
		return uint64(val), true
	case uint64:
		return val, true
	case int64:
		return uint64(val), true
	case int:
		return uint64(val), true
	}
	return 0, false
}

// GetMachineNodes 处理 machine mode 节点发现。
func (h *NodeServerHandler) GetMachineNodes(c *gin.Context) {
	machineID := middleware.GetMachineID(c)
	if machineID == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "missing machine_id"})
		return
	}

	db := database.Get()
	var nodes []model.Node
	if err := db.Where("machine_id = ?", machineID).Find(&nodes).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to query nodes"})
		return
	}

	nodeList := make([]gin.H, 0, len(nodes))
	for _, n := range nodes {
		nodeList = append(nodeList, gin.H{
			"id":      n.ID,
			"type":    n.Type,
			"name":    n.Name,
			"enabled": n.Enabled,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodeList,
		"base_config": gin.H{
			"push_interval": 60,
			"pull_interval": 60,
		},
	})
}

// ReportMachineStatus 处理 machine mode 状态上报。
func (h *NodeServerHandler) ReportMachineStatus(c *gin.Context) {
	machineID := middleware.GetMachineID(c)
	if machineID == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "missing machine_id"})
		return
	}

	body := middleware.GetNodeBody(c)

	db := database.Get()
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"is_active":   true,
		"last_seen_at": now,
	}

	loadStatus := make(map[string]interface{})

	if cpu, ok := body["cpu"].(float64); ok {
		updates["cpu"] = cpu
		loadStatus["cpu"] = cpu
	}
	var memTotal, memUsed uint64
	if mem, ok := body["mem"].(map[string]interface{}); ok {
		memTotal, _ = toUint64(mem["total"])
		memUsed, _ = toUint64(mem["used"])
		if memTotal > 0 {
			updates["memory"] = float64(memUsed) / float64(memTotal) * 100
		}
		loadStatus["mem_total"] = memTotal
		loadStatus["mem_used"] = memUsed
	}
	var diskTotal, diskUsed uint64
	if disk, ok := body["disk"].(map[string]interface{}); ok {
		diskTotal, _ = toUint64(disk["total"])
		diskUsed, _ = toUint64(disk["used"])
		if diskTotal > 0 {
			updates["disk"] = float64(diskUsed) / float64(diskTotal) * 100
		}
		loadStatus["disk_total"] = diskTotal
		loadStatus["disk_used"] = diskUsed
	}
	if net, ok := body["net"].(map[string]interface{}); ok {
		inSpeed, _ := toUint64(net["in_speed"])
		outSpeed, _ := toUint64(net["out_speed"])
		loadStatus["net_in_speed"] = inSpeed
		loadStatus["net_out_speed"] = outSpeed
	}

	if loadStatusJSON, err := json.Marshal(loadStatus); err == nil {
		updates["load_status"] = string(loadStatusJSON)
	}

	db.Model(&model.ServerMachine{}).Where("id = ?", machineID).Updates(updates)

	// 记录负载历史
	cpuVal, _ := updates["cpu"].(float64)
	history := &model.ServerMachineLoadHistory{
		MachineID:  uint(machineID),
		CPU:        cpuVal,
		MemTotal:   memTotal,
		MemUsed:    memUsed,
		DiskTotal:  diskTotal,
		DiskUsed:   diskUsed,
		RecordedAt: now,
	}
	if v, ok := loadStatus["net_in_speed"].(uint64); ok {
		history.NetInSpeed = float64(v)
	}
	if v, ok := loadStatus["net_out_speed"].(uint64); ok {
		history.NetOutSpeed = float64(v)
	}
	db.Create(history)

	c.JSON(http.StatusOK, gin.H{})
}

// markNodeOnlineREST 标记节点为在线。
func markNodeOnlineREST(nodeID uint32) (*model.Node, error) {
	db := database.Get()
	var node model.Node
	if err := db.First(&node, nodeID).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	node.Status = 1
	node.LastOnline = &now
	if err := db.Save(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}
