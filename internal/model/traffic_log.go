package model

import "time"

// TrafficLog 流量使用日志
type TrafficLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index:idx_user_recorded;not null" json:"user_id"`
	NodeID     uint      `gorm:"not null" json:"node_id"`
	Upload     int64     `gorm:"default:0" json:"upload"`   // 上传流量（字节）
	Download   int64     `gorm:"default:0" json:"download"` // 下载流量（字节）
	RecordedAt time.Time `gorm:"index:idx_user_recorded" json:"recorded_at"`
}

// TableName 指定表名
func (TrafficLog) TableName() string {
	return "traffic_logs"
}

// Total 总流量
func (t *TrafficLog) Total() int64 {
	return t.Upload + t.Download
}
