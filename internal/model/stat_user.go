package model

import "time"

// StatUser 用户流量统计模型
type StatUser struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	ServerRate float64   `gorm:"default:1" json:"server_rate"`
	Upload     int64     `gorm:"default:0" json:"upload"`
	Download   int64     `gorm:"default:0" json:"download"`
	RecordAt   time.Time `gorm:"index;not null" json:"record_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (StatUser) TableName() string {
	return "stat_users"
}
