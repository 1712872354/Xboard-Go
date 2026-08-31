package model

import (
	"time"
)

// ServerMachine 服务器机器模型
type ServerMachine struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`        // 机器名称
	Host        string    `gorm:"type:varchar(255);not null" json:"host"`        // 主机地址
	Port        int       `gorm:"not null" json:"port"`                          // 端口
	Protocol    string    `gorm:"type:varchar(50)" json:"protocol"`              // 协议
	Token       string    `gorm:"type:varchar(64);uniqueIndex" json:"token"`     // node authentication token
	Status      int       `gorm:"default:1" json:"status"`                       // 状态：1在线，0离线
	CPU         float64   `gorm:"default:0" json:"cpu"`                          // CPU使用率
	Memory      float64   `gorm:"default:0" json:"memory"`                       // 内存使用率
	Disk        float64   `gorm:"default:0" json:"disk"`                         // 磁盘使用率
	Uptime      int64     `gorm:"default:0" json:"uptime"`                       // 运行时间（秒）
	LastCheckAt *time.Time `json:"last_check_at"`                                // 最后检查时间
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ServerMachine) TableName() string {
	return "server_machines"
}

// IsOnline 是否在线
func (sm *ServerMachine) IsOnline() bool {
	return sm.Status == 1
}

// ServerMachineLoadHistory 服务器机器负载历史
type ServerMachineLoadHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MachineID uint      `gorm:"index;not null" json:"machine_id"` // 关联机器ID
	CPU       float64   `gorm:"default:0" json:"cpu"`              // CPU使用率
	Memory    float64   `gorm:"default:0" json:"memory"`           // 内存使用率
	Disk      float64   `gorm:"default:0" json:"disk"`             // 磁盘使用率
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ServerMachineLoadHistory) TableName() string {
	return "server_machine_load_histories"
}
