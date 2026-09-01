package model

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// 节点类型
const (
	NodeTypeVMess       = "vmess"
	NodeTypeVLESS       = "vless"
	NodeTypeTrojan      = "trojan"
	NodeTypeShadowsocks = "shadowsocks"
	NodeTypeHysteria2   = "hysteria2"
	NodeTypeHysteria    = "hysteria"
	NodeTypeTUIC        = "tuic"
	NodeTypeAnyTLS      = "anytls"
	NodeTypeNaive       = "naive"
	NodeTypeSocks       = "socks"
	NodeTypeHTTP        = "http"
	NodeTypeMieru       = "mieru"
)

// Node 节点模型
type Node struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	Name                string     `gorm:"type:varchar(100);not null" json:"name"`
	Type                string     `gorm:"type:varchar(20);not null" json:"type"`
	Host                string     `gorm:"type:varchar(255)" json:"host"`               // 主机地址
	Port                int        `gorm:"not null" json:"port"`                        // 客户端连接端口
	ServerPort          int        `gorm:"default:0" json:"server_port"`               // 后端服务端口
	Code                string     `gorm:"type:varchar(255)" json:"code"`              // 节点标识符（原版: spectific_key/code）
	ServerInfo          string     `gorm:"type:text" json:"server_info"`               // JSON格式，存储协议设置
	GroupIDs            string     `gorm:"type:varchar(255);default:'';index" json:"group_ids"` // 逗号分隔的权限组ID，如 "1,2,3"
	Rate                float64    `gorm:"type:decimal(5,2);default:1.0" json:"rate"`
	Status              int        `gorm:"default:1" json:"status"`
	ParentID            uint       `gorm:"default:0" json:"parent_id"`
	MachineID           *uint      `gorm:"index" json:"machine_id"`
	Enabled             bool       `gorm:"default:true" json:"enabled"`
	Sort                int        `gorm:"default:0;index" json:"sort"`
	Show                int        `gorm:"default:1" json:"show"`
	Tags                string     `gorm:"type:varchar(500)" json:"tags"`
	OnlineUserCount     int        `gorm:"default:0" json:"online_user_count"`
	UploadTraffic       int64      `gorm:"default:0" json:"u"`                         // 上行流量（原版字段名 u）
	DownloadTraffic     int64      `gorm:"default:0" json:"d"`                         // 下行流量（原版字段名 d）
	TransferEnable      int64      `gorm:"default:0" json:"transfer_enable"`            // 流量上限，0表示不限制
	RateTimeEnable      bool       `gorm:"default:false" json:"rate_time_enable"`
	RateTimeRanges      string     `gorm:"type:text" json:"rate_time_ranges"`           // JSON: [{start, end, rate}]
	CustomOutbounds     string     `gorm:"type:text" json:"custom_outbounds"`           // JSON: 自定义出站
	CustomRoutes        string     `gorm:"type:text" json:"custom_routes"`              // JSON: 自定义路由
	CertConfig          string     `gorm:"type:text" json:"cert_config"`                // JSON: 证书配置
	Ports               string     `gorm:"type:varchar(255)" json:"ports"`              // 端口范围
	HealthCheckPort     int        `gorm:"default:0" json:"health_check_port"`
	HealthCheckInterval int        `gorm:"default:60" json:"health_check_interval"`
	HealthCheckTimeout  int        `gorm:"default:5" json:"health_check_timeout"`
	HealthCheckType     string     `gorm:"type:varchar(20);default:'tcp'" json:"health_check_type"`
	LastOnline          *time.Time `json:"last_online"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (Node) TableName() string {
	return "nodes"
}

// IsOnline 节点是否在线
func (n *Node) IsOnline() bool {
	return n.Status == 1
}

// IsVisible 是否在订阅中显示
func (n *Node) IsVisible() bool {
	return n.Show == 1 && n.Status != 2 // 显示且非维护中
}

// ParseServerInfo 解析ServerInfo JSON为map
func (n *Node) ParseServerInfo() map[string]interface{} {
	if n.ServerInfo == "" {
		return make(map[string]interface{})
	}
	var info map[string]interface{}
	if err := json.Unmarshal([]byte(n.ServerInfo), &info); err != nil {
		return make(map[string]interface{})
	}
	return info
}

// ServerInfoConfig 通用节点协议配置（从ServerInfo解析）
type ServerInfoConfig struct {
	// 通用
	Network       string `json:"network"`         // 传输协议：tcp, ws, grpc, h2, quic
	NetworkType   string `json:"network_type"`    // 同 network，兼容字段
	TLS           bool   `json:"tls"`             // 是否启用TLS
	TLSServerName string `json:"tls_server_name"` // TLS SNI
	ALPN          string `json:"alpn"`            // ALPN 协议
	AllowInsecure bool   `json:"allow_insecure"`  // 允许不安全连接

	// WebSocket
	WSPath string `json:"ws_path"` // WebSocket路径
	WSHost string `json:"ws_host"` // WebSocket Host

	// gRPC
	GrpcServiceName string `json:"grpc_service_name"` // gRPC 服务名

	// HTTP/2
	H2Path string `json:"h2_path"` // H2路径
	H2Host string `json:"h2_host"` // H2 Host

	// Shadowsocks
	Cipher   string `json:"cipher"`   // 加密方式
	Password string `json:"password"` // 密码（SS专用）

	// VMess
	AlterID  int    `json:"alter_id"` // VMess alterID
	Security string `json:"security"` // VMess加密方式

	// Hysteria2 / Hysteria
	UpMbps   int    `json:"up_mbps"`   // 上行带宽
	DownMbps int    `json:"down_mbps"` // 下行带宽
	Obfs     string `json:"obfs"`      // 混淆密码
	ObfsType string `json:"obfs_type"` // 混淆类型

	// TUIC
	CongestionControl string `json:"congestion_control"` // 拥塞控制

	// 机器ID
	MachineID uint `json:"machine_id"` // 关联的服务器机器ID
}

// ParseServerInfoConfig 解析ServerInfo为结构化配置
func (n *Node) ParseServerInfoConfig() *ServerInfoConfig {
	info := n.ParseServerInfo()
	data, _ := json.Marshal(info)
	var cfg ServerInfoConfig
	_ = json.Unmarshal(data, &cfg)
	// 兼容 network / network_type
	if cfg.Network == "" && cfg.NetworkType != "" {
		cfg.Network = cfg.NetworkType
	}
	return &cfg
}

// GetGroupIDList 解析 GroupIDs 为 []uint 切片
func (n *Node) GetGroupIDList() []uint {
	if n.GroupIDs == "" {
		return nil
	}
	var ids []uint
	for _, s := range strings.Split(n.GroupIDs, ",") {
		s = strings.TrimSpace(s)
		if id, err := strconv.ParseUint(s, 10, 32); err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

// HasGroup 检查节点是否属于指定权限组
func (n *Node) HasGroup(groupID uint) bool {
	for _, id := range n.GetGroupIDList() {
		if id == groupID {
			return true
		}
	}
	return false
}

// SetGroupIDs 设置权限组ID列表
func (n *Node) SetGroupIDs(ids []uint) {
	if len(ids) == 0 {
		n.GroupIDs = ""
		return
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.FormatUint(uint64(id), 10)
	}
	n.GroupIDs = strings.Join(strs, ",")
}

// GroupIDCondition 返回 GORM 查询条件参数，用于按单个 group_id 过滤逗号分隔的 group_ids 字段
// 返回值: condition string, args []interface{}
func GroupIDCondition(groupID uint) (string, []interface{}) {
	idStr := strconv.FormatUint(uint64(groupID), 10)
	cond := "group_ids = ? OR group_ids LIKE ? OR group_ids LIKE ? OR group_ids LIKE ?"
	args := []interface{}{idStr, idStr + ",%", "%," + idStr, "%," + idStr + ",%"}
	return cond, args
}

// GroupIDsContainAny 返回 GORM 查询条件，检查 group_ids 是否包含任意一个指定的组ID
func GroupIDsContainAny(groupIDs []uint) (string, []interface{}) {
	if len(groupIDs) == 0 {
		return "1 = 0", nil // 无组ID时返回不可能的条件
	}
	var conds []string
	var args []interface{}
	for _, gid := range groupIDs {
		idStr := strconv.FormatUint(uint64(gid), 10)
		conds = append(conds, "(group_ids = ? OR group_ids LIKE ? OR group_ids LIKE ? OR group_ids LIKE ?)")
		args = append(args, idStr, idStr+",%", "%,"+idStr, "%,"+idStr+",%")
	}
	return "(" + strings.Join(conds, " OR ") + ")", args
}
