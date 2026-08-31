package model

import (
	"time"
)

// ServerStat 服务器统计模型
type ServerStat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServerID  uint      `gorm:"index;not null" json:"server_id"`     // 关联服务器ID
	Date      string    `gorm:"type:varchar(10);index;not null" json:"date"` // 日期（YYYY-MM-DD）
	Upload    int64     `gorm:"default:0" json:"upload"`              // 上传流量（字节）
	Download  int64     `gorm:"default:0" json:"download"`            // 下载流量（字节）
	Total     int64     `gorm:"default:0" json:"total"`               // 总流量（字节）
	Users     int       `gorm:"default:0" json:"users"`               // 用户数
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ServerStat) TableName() string {
	return "server_stats"
}

// ServerLog 服务器日志模型
type ServerLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServerID  uint      `gorm:"index;not null" json:"server_id"`     // 关联服务器ID
	Level     string    `gorm:"type:varchar(20);not null" json:"level"` // 日志级别：info, warn, error
	Message   string    `gorm:"type:text;not null" json:"message"`    // 日志内容
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ServerLog) TableName() string {
	return "server_logs"
}

// 日志级别常量
const (
	ServerLogLevelInfo  = "info"
	ServerLogLevelWarn  = "warn"
	ServerLogLevelError = "error"
)
