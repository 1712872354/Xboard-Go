package model

import (
	"encoding/json"
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
	Type                string     `gorm:"type:varchar(20);not null" json:"type"` // vmess, vless, trojan, shadowsocks, hysteria2, hysteria, tuic, anytls, naive, socks, http, mieru
	Address             string     `gorm:"type:varchar(255);not null" json:"address"`
	Port                int        `gorm:"not null" json:"port"`
	ServerInfo          string     `gorm:"type:text" json:"server_info"` // JSON格式，存储额外配置（加密方式、传输协议等）
	GroupID             uint       `gorm:"default:0;index" json:"group_id"`
	Rate                float64    `gorm:"type:decimal(5,2);default:1.0" json:"rate"` // 流量倍率
	Status              int        `gorm:"default:1" json:"status"`                   // 1: 在线, 0: 离线, 2: 维护中
	ParentID            uint       `gorm:"default:0" json:"parent_id"`
	Sort                int        `gorm:"default:0;index" json:"sort"`        // 排序值，越小越靠前
	Show                int        `gorm:"default:1" json:"show"`              // 是否在订阅中显示：1显示，0隐藏
	Tags                string     `gorm:"type:varchar(500)" json:"tags"`      // 标签，逗号分隔
	OnlineUserCount     int        `gorm:"default:0" json:"online_user_count"` // 当前在线用户数
	TrafficUsed         int64      `gorm:"default:0" json:"traffic_used"`      // 节点累计已用流量（字节）
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
