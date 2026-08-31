package model

import "time"

// StatServer 节点流量统计模型
type StatServer struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ServerID   uint      `gorm:"index;not null" json:"server_id"`
	ServerType string    `gorm:"type:varchar(20)" json:"server_type"`
	Upload     int64     `gorm:"default:0" json:"upload"`
	Download   int64     `gorm:"default:0" json:"download"`
	RecordAt   time.Time `gorm:"index;not null" json:"record_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (StatServer) TableName() string {
	return "stat_servers"
}
