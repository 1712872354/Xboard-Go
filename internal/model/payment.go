package model

import (
	"time"
)

// Payment 支付方式模型
type Payment struct {
	ID            uint                   `gorm:"primaryKey" json:"id"`
	Name          string                 `gorm:"type:varchar(100);not null" json:"name"`
	Icon          string                 `gorm:"type:varchar(255)" json:"icon"`
	Payment       string                 `gorm:"type:varchar(50);uniqueIndex;not null" json:"payment"`
	Config        map[string]interface{} `gorm:"type:json" json:"config"`
	NotifyDomain  string                 `gorm:"type:varchar(255)" json:"notify_domain"`
	Status        int                    `gorm:"default:1" json:"status"` // 1: 启用, 0: 禁用
	Sort          int                    `gorm:"default:0" json:"sort"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// TableName 指定表名
func (Payment) TableName() string {
	return "payments"
}

// IsActive 是否激活
func (p *Payment) IsActive() bool {
	return p.Status == 1
}
