package model

import (
	"time"
)

// Theme 主题模型
type Theme struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Version     string    `gorm:"type:varchar(20)" json:"version"`
	Author      string    `gorm:"type:varchar(100)" json:"author"`
	Homepage    string    `gorm:"type:varchar(255)" json:"homepage"`
	Preview     string    `gorm:"type:varchar(255)" json:"preview"`
	Config      string    `gorm:"type:json" json:"config"`
	Status      int       `gorm:"default:1" json:"status"` // 1: 启用, 0: 禁用
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Theme) TableName() string {
	return "themes"
}

// IsActive 是否激活
func (t *Theme) IsActive() bool {
	return t.Status == 1
}
