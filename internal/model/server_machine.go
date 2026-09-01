package model

import (
	"time"
)

// ServerMachine 服务器机器模型
type ServerMachine struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"type:varchar(100);not null" json:"name"`    // 机器名称
	Token      string `gorm:"type:varchar(64);uniqueIndex" json:"token,omitempty"`     // 认证令牌
	Notes      string `gorm:"type:text" json:"notes"`                     // 备注
	IsActive   bool   `gorm:"default:true" json:"is_active"`             // 是否活跃
	LastSeenAt int64  `gorm:"default:0" json:"last_seen_at"`             // 最后心跳时间（Unix时间戳）
	LoadStatus string `gorm:"type:text" json:"load_status"`              // 负载状态JSON（cpu, mem_total, mem_used, disk_total, disk_used, net_in_speed, net_out_speed）
	CPU        float64 `gorm:"default:0" json:"cpu"`                      // CPU使用率（denormalized from load_status）
	Memory     float64 `gorm:"default:0" json:"memory"`                  // 内存使用率（denormalized from load_status）
	Disk       float64 `gorm:"default:0" json:"disk"`                    // 磁盘使用率（denormalized from load_status）
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ServerMachine) TableName() string {
	return "server_machines"
}

// IsOnline 是否在线
func (sm *ServerMachine) IsOnline() bool {
	return sm.IsActive
}

// ServerMachineLoadHistory 服务器机器负载历史
type ServerMachineLoadHistory struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	MachineID   uint    `gorm:"index;not null" json:"machine_id"`         // 关联机器ID
	CPU         float64 `gorm:"default:0" json:"cpu"`                    // CPU使用率
	MemTotal    uint64  `gorm:"default:0" json:"mem_total"`              // 内存总量（字节）
	MemUsed     uint64  `gorm:"default:0" json:"mem_used"`               // 已用内存（字节）
	DiskTotal   uint64  `gorm:"default:0" json:"disk_total"`             // 磁盘总量（字节）
	DiskUsed    uint64  `gorm:"default:0" json:"disk_used"`              // 已用磁盘（字节）
	NetInSpeed  float64 `gorm:"default:0" json:"net_in_speed"`          // 入站速度（B/s）
	NetOutSpeed float64 `gorm:"default:0" json:"net_out_speed"`         // 出站速度（B/s）
	RecordedAt  int64   `gorm:"not null" json:"recorded_at"`             // 记录时间（Unix时间戳）
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (ServerMachineLoadHistory) TableName() string {
	return "server_machine_load_histories"
}
